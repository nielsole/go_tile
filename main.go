package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

/*
 * The findPath function is based upon mod_tile code:
 * Copyright (c) 2007 - 2020 by mod_tile contributors (see AUTHORS file)
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; If not, see http://www.gnu.org/licenses/.
 */

func findPath(baseDir string, z, x, y uint32) (metaPath string, offset uint32) {
	var mask uint32
	var hash [5]byte

	// Default value
	var METATILE uint32
	METATILE = 8
	mask = METATILE - 1
	offset = (x&mask)*METATILE + (y & mask)
	x &= ^mask
	y &= ^mask

	for i := 0; i < 5; i++ {
		hash[i] = byte(((x & 0x0f) << 4) | (y & 0x0f))
		x >>= 4
		y >>= 4
	}
	metaPath = fmt.Sprintf("%s/%d/%d/%d/%d/%d/%d.meta", baseDir, z, hash[4], hash[3], hash[2], hash[1], hash[0])
	return
}

func readInt(file *os.File) (uint32, error) {
	b := make([]byte, 4)
	bytesRead, err := file.Read(b)
	if err != nil {
		return 0, err
	} else if bytesRead != 4 {
		return 0, errors.New("incorrect amount of bytes read")
	}
	return binary.LittleEndian.Uint32(b), nil
}

func readPNGTile(writer http.ResponseWriter, req *http.Request, metatile_path string, metatile_offset uint32, modTime time.Time) error {
	file, err := os.Open(metatile_path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			fmt.Println("Error opening file!!!", metatile_path)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return nil
	}
	defer file.Close()
	writer.Header().Add("Cache-Control", "no-cache")

	file.Seek(4, 0)
	tile_count, err := readInt(file)
	if err != nil {
		return err
	}
	if metatile_offset >= tile_count {
		return errors.New("requested offset exceeded bounds of metatile")
	}
	file.Seek(int64(20+metatile_offset*2*4), 0)
	tile_offset, err := readInt(file)
	if err != nil {
		return err
	}
	tile_length, err := readInt(file)
	if err != nil {
		return err
	}
	file.Stat()
	file.Seek(int64(tile_offset), 0)
	http.ServeContent(writer, req, "file.png", modTime, NewSubFileReaderSeeker(file, int(tile_offset), int(tile_length)))
	return nil
}

func parsePath(path string) (z, x, y uint32, err error) {
	matcher := regexp.MustCompile(`/tile/([0-9]+)/([0-9]+)/([0-9]+).png`)
	matches := matcher.FindStringSubmatch(path)
	if len(matches) != 4 {
		return 0, 0, 0, errors.New("could not match path")
	}
	zInt, err := strconv.Atoi(matches[1])
	if err != nil {
		return
	}
	xInt, err := strconv.Atoi(matches[2])
	if err != nil {
		return
	}
	yInt, err := strconv.Atoi(matches[3])
	if err != nil {
		return
	}
	z = uint32(zInt)
	x = uint32(xInt)
	y = uint32(yInt)
	return
}

func requestRender(x, y, z uint32) error {
	c, err := net.Dial("unix", "/var/run/renderd/renderd.sock")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(30 * time.Second))
	// Version
	if err := binary.Write(c, binary.LittleEndian, uint32(3)); err != nil {
		return err
	}
	// RenderPrio
	if err := binary.Write(c, binary.LittleEndian, uint32(5)); err != nil {
		return err
	}
	if err := binary.Write(c, binary.LittleEndian, x); err != nil {
		return err
	}
	if err := binary.Write(c, binary.LittleEndian, y); err != nil {
		return err
	}
	if err := binary.Write(c, binary.LittleEndian, z); err != nil {
		return err
	}
	if n, err := c.Write([]byte("ajt")); n != 3 {
		return errors.New("could not write request. Not all bytes were written")
	} else if err != nil {
		return err
	}
	// Filling up null bytes
	for i := 0; i < 64-(4*5+len("ajt")); i++ {
		c.Write([]byte{0})
	}

	var protocol_version uint32
	err = binary.Read(c, binary.LittleEndian, &protocol_version)
	if err != nil {
		return err
	}
	if protocol_version != 3 {
		return fmt.Errorf("unsupported protocol version: %d", protocol_version)
	}
	var response_code uint32
	err = binary.Read(c, binary.LittleEndian, &response_code)
	if err != nil {
		return err
	}
	if response_code != 3 {
		return fmt.Errorf("render request not successful. Received response code: %d", response_code)
	}
	response := make([]byte, 56)
	n, err := c.Read(response)
	if err != nil {
		return err
	}
	if n != len(response) {
		return errors.New("could not read response. Unexpected number of bytes")
	}
	return nil
}

func handleRequest(resp http.ResponseWriter, req *http.Request, data_dir *string) {
	z, x, y, err := parsePath(req.URL.Path)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Header().Add("Content-Type", "image/png")
	metatile_path, metatile_offset := findPath(*data_dir, z, x, y)
	fileInfo, statErr := os.Stat(metatile_path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			renderErr := requestRender(x, y, z)
			if renderErr != nil {
				fmt.Printf("Could not generate tile for coordinates %d, %d, %d (x,y,z). '%s'\n", x, y, z, renderErr)
				// Not returning as we are hoping and praying that rendering did nonetheless produce a file
			}
			if fileInfo, statErr = os.Stat(metatile_path); statErr != nil {
				if renderErr == nil {
					fmt.Printf("warning: metatile could not be found after successful render. Are the paths matching? Tried %s\n", metatile_path)
				}
				// we haven't checked if this was actually a NotFound error, and even then, this is not a client error, so a 5xx is warranted
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	modTime := fileInfo.ModTime()
	errPng := readPNGTile(resp, req, metatile_path, metatile_offset, modTime)
	if errPng != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	listen_port := flag.String("port", ":8080", "Listening port")
	data_dir := flag.String("data", "./data", "Path to directory containing tiles")
	static_dir := flag.String("static", "./static/", "Path to static file directory")
	flag.Parse()
	http.HandleFunc("/tile/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed) // TODO return 4xx wrong method
			w.Write([]byte("Only GET requests allowed"))
			return
		}
		handleRequest(w, r, data_dir)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(*static_dir)).ServeHTTP(w, r)
	})
	fmt.Printf("Listening on port %s\n", *listen_port)
	err := http.ListenAndServe(*listen_port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
