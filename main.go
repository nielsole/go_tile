package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nielsole/go_tile/renderer"
	"github.com/nielsole/go_tile/utils"
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

func readTile(metatile_path string, metatile_offset uint32) (*os.File, *io.SectionReader, error) {
	file, err := os.Open(metatile_path)
	if err != nil {
		return nil, nil, err
	}
	file.Seek(4, 0)
	tile_count, err := readInt(file)
	if err != nil {
		return nil, nil, err
	}
	if metatile_offset >= tile_count {
		return nil, nil, errors.New("requested offset exceeded bounds of metatile")
	}
	file.Seek(int64(20+metatile_offset*2*4), 0)
	tile_offset, err := readInt(file)
	if err != nil {
		return nil, nil, err
	}
	tile_length, err := readInt(file)
	if err != nil {
		return nil, nil, err
	}
	return file, io.NewSectionReader(file, int64(tile_offset), int64(tile_length)), nil
}

func writeTileResponse(writer http.ResponseWriter, req *http.Request, metatile_path string, metatile_offset uint32, modTime time.Time, ext string) error {
	file, tileReader, err := readTile(metatile_path, metatile_offset)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			logWarningf("Could not open file %s: %v", metatile_path, err)
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return nil
	}
	writer.Header().Add("Cache-Control", "no-cache")
	http.ServeContent(writer, req, "file."+ext, modTime, tileReader)
	return nil
}

func handleRequest(resp http.ResponseWriter, req *http.Request, data_dir, map_name, renderd_socket string, renderd_timeout time.Duration, tile_expiration time.Duration) {
	z, x, y, ext, err := utils.ParsePath(req.URL.Path)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Header().Add("Content-Type", "image/"+ext)
	metatile_path, metatile_offset := findPath(data_dir, map_name, z, x, y)
	fileInfo, statErr := os.Stat(metatile_path)
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			if len(renderd_socket) == 0 {
				logErrorf("Metatile does not exist: %s", metatile_path)
				resp.WriteHeader(http.StatusNotFound)
				return
			}
			renderErr := requestRender(x, y, z, map_name, renderd_socket, renderd_timeout, 5)
			if renderErr != nil {
				logWarningf("Could not generate tile for coordinates %d, %d, %d (x,y,z). '%s'", x, y, z, renderErr)
				// Not returning as we are hoping and praying that rendering did nonetheless produce a file
			}
			if fileInfo, statErr = os.Stat(metatile_path); statErr != nil {
				if renderErr == nil {
					logWarningf("Metatile could not be found after successful render. Are the paths matching? Tried %s", metatile_path)
				}
				// we haven't checked if this was actually a NotFound error, and even then, this is not a client error, so a 5xx is warranted
				resp.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if tile_expiration > 0 {
		modTime := fileInfo.ModTime()
		if modTime.Add(tile_expiration).Before(time.Now()) {
			go requestRender(x, y, z, map_name, renderd_socket, renderd_timeout, 7)
		}
	}
	modTime := fileInfo.ModTime()
	errTile := writeTileResponse(resp, req, metatile_path, metatile_offset, modTime, ext)
	if errTile != nil {
		resp.WriteHeader(http.StatusInternalServerError)
	}
}

func getSocketType(renderd_socket string) (string, *net.TCPAddr) {
	tcp_addr, _ := net.ResolveTCPAddr("tcp", renderd_socket)
	if tcp_addr != nil {
		return "tcp", tcp_addr
	}
	return "unix", nil
}

func main() {
	http_listen_host := flag.String("host", "0.0.0.0", "HTTP Listening host")
	http_listen_port := flag.Int("port", 8080, "HTTP Listening port")
	https_listen_port := flag.Int("tls_port", 8443, "HTTPS Listening port. This listener is only enabled if both tls cert and key are set.")
	//data_dir := flag.String("data", "./data", "Path to directory containing tiles")
	static_dir := flag.String("static", "./static/", "Path to static file directory")
	renderd_socket := flag.String("socket", "", "Unix domain socket path or hostname:port for contacting renderd. Rendering disabled by default.")
	renderd_timeout_duration := flag.Duration("renderd-timeout", (time.Duration(60) * time.Second), "Timeout duration after which renderd returns an error to the client (I.E. '30s' for thirty seconds). Set negative to disable")
	map_name := flag.String("map", "ajt", "Name of map. This value is also used to determine the metatile subdirectory")
	tls_cert_path := flag.String("tls_cert_path", "", "Path to TLS certificate")
	tls_key_path := flag.String("tls_key_path", "", "Path to TLS key")
	//tile_expiration_duration := flag.Duration("tile_expiration", 0, "Duration after which tiles are considered stale (I.E. '168h' for one week). Tile expiration disabled by default")
	verbose := flag.Bool("verbose", false, "Output debug log messages")

	flag.Parse()

	// Renderd expects at most 64 bytes.
	// 64 - (5 * 4 bytes - 1 zero byte of null-terminated string) = 43
	if len(*map_name) > 43 {
		logFatalf("Map name may not be longer than 43 characters")
	}
	if len(*renderd_socket) > 0 {
		renderd_socket_type, renderd_tcp_addr := getSocketType(*renderd_socket)
		if renderd_socket_type == "tcp" {
			c, err := net.DialTCP("tcp", nil, renderd_tcp_addr)
			if err != nil {
				logFatalf("There was an error with the renderd %s socket at '%s': %v", renderd_socket_type, *renderd_socket, err)
			}
			c.Close()
		} else {
			_, err := os.Stat(*renderd_socket)
			if err != nil {
				logFatalf("There was an error with the renderd %s socket at '%s': %v", renderd_socket_type, *renderd_socket, err)
			}
		}
		logInfof("Using renderd %s socket at '%s'\n", renderd_socket_type, *renderd_socket)
	} else {
		logInfof("Rendering is disabled")
	}

	// Create a temp file.
	var err error
	tempFile, err := ioutil.TempFile("", "example")
	if err != nil {
		fmt.Println("Cannot create temp file:", err)
		os.Exit(1)
	}

	data, err := renderer.LoadData("/home/nokadmin/projects/go_tile/mock_data/test.osm.pbf", 15, tempFile)
	if err != nil {
		logFatalf("There was an error loading data: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()

	// Memory-map the file
	mmapData, mmapFile, err := renderer.Mmap(tempFileName)
	if err != nil {
		logFatalf("There was an error memory-mapping temp file: %v", err)
	}
	defer syscall.Munmap(*mmapData)
	defer mmapFile.Close()

	// HTTP request multiplexer
	httpServeMux := http.NewServeMux()

	// Tile HTTP request handler
	httpServeMux.HandleFunc("/tile/", func(w http.ResponseWriter, r *http.Request) {
		if *verbose {
			logDebugf("%s request received: %s", r.Method, r.RequestURI)
		}
		if r.Method != "GET" {
			http.Error(w, "Only GET requests allowed", http.StatusMethodNotAllowed)
			return
		}
		renderer.HandleRenderRequest(w, r, *renderd_timeout_duration, data, 15, mmapData)
		//handleRequest(w, r, *data_dir, *map_name, *renderd_socket, *renderd_timeout_duration, *tile_expiration_duration)
	})

	// Static HTTP request handler
	httpServeMux.Handle("/", http.FileServer(http.Dir(*static_dir)))

	// HTTP Server
	httpServer := http.Server{
		Handler: httpServeMux,
	}

	go func() {

		// HTTPS listener
		if len(*tls_cert_path) > 0 && len(*tls_key_path) > 0 {
			go func() {
				httpsAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", *http_listen_host, *https_listen_port))
				if err != nil {
					logFatalf("Failed to resolve TCP address: %v", err)
				}
				httpsListener, err := net.ListenTCP("tcp", httpsAddr)
				if err != nil {
					logFatalf("Failed to start TCP listener: %v", err)
				} else {
					logInfof("Started HTTPS listener on %s\n", httpsAddr)
				}
				err = httpServer.ServeTLS(httpsListener, *tls_cert_path, *tls_key_path)
				if err != nil && err != http.ErrServerClosed {
					logFatalf("Failed to start HTTPS server: %v", err)
				}
			}()
		} else {
			logInfof("TLS is disabled")
		}

		// HTTP listener
		httpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", *http_listen_host, *http_listen_port))
		if err != nil {
			logFatalf("Failed to resolve TCP address: %v", err)
		}
		httpListener, err := net.ListenTCP("tcp", httpAddr)
		if err != nil {
			logFatalf("Failed to start TCP listener: %v", err)
		} else {
			logInfof("Started HTTP listener on %s\n", httpAddr)
		}
		err = httpServer.Serve(httpListener)
		if err != nil && err != http.ErrServerClosed {
			logFatalf("Failed to start HTTP server: %v", err)
		}
	}()
	// Setup signal capturing.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Waiting for SIGINT (Ctrl+C)
	select {
	case <-stop:
		fmt.Println("\nShutting down the server...")

		// Create a deadline for the shutdown process.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start shutdown.
		if err := httpServer.Shutdown(ctx); err != nil {
			fmt.Println("Error during server shutdown:", err)
		}

		// Cleanup the temp file.
		if err := os.Remove(tempFile.Name()); err != nil {
			fmt.Println("Failed to remove temp file:", err)
		} else {
			fmt.Println("Temp file removed.")
		}

		// Additional cleanup code here...

		fmt.Println("Server gracefully stopped.")
	}
}
