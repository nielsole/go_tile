package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
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

func findPath(baseDir string, x uint32, y uint32, z int) (metaPath string, offset uint32) {
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
		return 0, errors.New("Incorrect amount of bytes read")
	}
	return binary.LittleEndian.Uint32(b), nil
}

func readPNGTile(writer http.ResponseWriter, req *http.Request, metatile_path string, metatile_offset uint32) error {
	fileInfo, statErr := os.Stat(metatile_path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			writer.WriteHeader(http.StatusNotFound)
			return nil
		}
		return statErr
	}
	modTime := fileInfo.ModTime()
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

	// declare chunk size
	const maxSz = 4

	file.Seek(4, 0)
	tile_count, err := readInt(file)
	if err != nil {
		return err
	}
	if metatile_offset >= tile_count {
		return errors.New("Requested offset exceeded bounds of metatile")
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

func handleRequest(resp http.ResponseWriter, req *http.Request, data_dir *string) {
	resp.Header().Add("Content-Type", "image/png")
	vars := mux.Vars(req)
	// _should_ never fail as the path was already matched by mux
	x, _ := strconv.Atoi(vars["x"])
	y, _ := strconv.Atoi(vars["y"])
	z, _ := strconv.Atoi(vars["z"])
	metatile_path, metatile_offset := findPath(*data_dir, uint32(x), uint32(y), z)
	errPng := readPNGTile(resp, req, metatile_path, metatile_offset)
	if errPng != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func main() {
	listen_port := flag.String("port", ":8080", "Listening port")
	data_dir := flag.String("data", "./data", "Path to directory containing tiles")
	static_dir := flag.String("static", "./static/", "Path to static file directory")
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/tile/{z:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.png", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, data_dir)
	}).Methods("GET")
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.FileServer(http.Dir(*static_dir)).ServeHTTP(w, r)
	})
	http.Handle("/", r)
	fmt.Printf("Listening on port %s\n", *listen_port)
	err := http.ListenAndServe(*listen_port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
