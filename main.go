package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
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

func findPath(baseDir, mapName string, z, x, y uint32) (metaPath string, offset uint32) {
	var mask uint32
	var hash [5]byte

	// Default value
	var METATILE = uint32(8)
	mask = METATILE - 1
	offset = (x&mask)*METATILE + (y & mask)
	x &= ^mask
	y &= ^mask

	for i := 0; i < 5; i++ {
		hash[i] = byte(((x & 0x0f) << 4) | (y & 0x0f))
		x >>= 4
		y >>= 4
	}
	metaPath = fmt.Sprintf("%s/%s/%d/%d/%d/%d/%d/%d.meta", baseDir, mapName, z, hash[4], hash[3], hash[2], hash[1], hash[0])
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
			fmt.Println("Could not open file!", metatile_path)
			fmt.Println(err)
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
	http.ServeContent(writer, req, "file.png", modTime, io.NewSectionReader(file, int64(tile_offset), int64(tile_length)))
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

func handleRequest(resp http.ResponseWriter, req *http.Request, data_dir, map_name, renderd_sock_path string, renderd_timeout time.Duration) {
	z, x, y, err := parsePath(req.URL.Path)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Header().Add("Content-Type", "image/png")
	metatile_path, metatile_offset := findPath(data_dir, map_name, z, x, y)
	fileInfo, statErr := os.Stat(metatile_path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			if len(renderd_sock_path) == 0 {
				resp.WriteHeader(http.StatusNotFound)
				return
			}
			renderErr := requestRender(x, y, z, map_name, renderd_sock_path, renderd_timeout)
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
	http_listen_port := flag.String("port", ":8080", "HTTP Listening port")
	tls_listen_port := flag.String("tls_port", ":8443", "HTTPS Listening port. This listener is only enabled if both tls cert and key are set.")
	data_dir := flag.String("data", "./data", "Path to directory containing tiles")
	static_dir := flag.String("static", "./static/", "Path to static file directory")
	renderd_sock_path := flag.String("socket", "/var/run/renderd/renderd.sock", "Path to renderd socket. Set to '' to disable rendering")
	renderd_timeout := flag.Int("renderd-timeout", 60, "time in seconds to wait for renderd before returning an error to the client. Set negative to disable")
	map_name := flag.String("map", "ajt", "Name of map. This value is also used to determine the metatile subdirectory")
	tls_cert_path := flag.String("tls_cert_path", "", "Path to TLS certificate")
	tls_key_path := flag.String("tls_key_path", "", "Path to TLS key")
	var renderd_timeout_duration time.Duration = time.Duration(*renderd_timeout) * time.Second
	flag.Parse()
	// Renderd expects at most 64 bytes.
	// 64 - (5 * 4 bytes - 1 zero byte of null-terminated string) = 43
	if len(*map_name) > 43 {
		log.Fatal("Map name may not be longer than 43 characters")
	}
	if len(*renderd_sock_path) > 0 {
		_, err := os.Stat(*renderd_sock_path)
		if err != nil {
			log.Printf("Sanity Check Warning: There was an error with the renderd socket at '%s': %v", *renderd_sock_path, err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/tile/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed) // TODO return 4xx wrong method
			w.Write([]byte("Only GET requests allowed"))
			return
		}
		handleRequest(w, r, *data_dir, *map_name, *renderd_sock_path, renderd_timeout_duration)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(*static_dir)).ServeHTTP(w, r)
	})

	server := &http.Server{
		Handler: mux,
	}
	if len(*tls_cert_path) > 0 && len(*tls_key_path) > 0 {
		go func() {
			fmt.Printf("Listening on port %s\n", *tls_listen_port)
			tls_listener, err := net.Listen("tcp", *tls_listen_port)
			if err != nil {
				log.Fatal(err)
			}
			err = server.ServeTLS(tls_listener, *tls_cert_path, *tls_key_path)
			if err != nil {
				log.Fatal(err)
			}
		}()
	} else {
		fmt.Println("TLS is disabled")
	}
	fmt.Printf("Listening on port %s\n", *http_listen_port)
	http_listener, err := net.Listen("tcp", *http_listen_port)
	if err != nil {
		log.Fatal(err)
	}
	err = server.Serve(http_listener)
	if err != nil {
		log.Fatal(err)
	}
}
