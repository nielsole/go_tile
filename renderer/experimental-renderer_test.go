//go:build !windows
// +build !windows

package renderer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

// Test to check bounding box
func TestBoundingBox(t *testing.T) {
	// Test bounding box
	p := Point{10.068140300000001, 53.6577386}
	tile := Tile{
		X: 138403,
		Y: 84591,
		Z: 18,
	}
	bbox := getBoundingBox(tile)
	if !bbox.contains(p) {
		t.Errorf("Point %v is not in bounding box %v", p, bbox)
	}
}

// How to debug
// go test -benchtime 15s -bench=BenchmarkServe -cpuprofile cpu.out
// go tool pprof cpu.out
// svg

// 26440844f98323a1ea611acd3be00196a7b2a48d 2023-06-27
//
//	752985932 ns/op
//
// 643b79ef31ade931afa5037381eeecd8a2a19d19
//
//	54120852 ns/op
//
// b82ccbeef7bebe3647783ea9b3ed80638d5785cd
//
//				  8390724 ns/op
//	           5443820 ns/op
//
// Full tile
// 3484142532 ns/op
// cec1fd4962a9fd01956f827a7ca74d98560bb6a1
//
//		293297990 ns/op
//	     88056 ns/op
//
// c03af9bccb2499f3f0291c8e8aca22f141f06600
//
//	77898 ns/op
//
// ef29875363c610d536be31e603921c71ef698468
//
//	95696 ns/op
func BenchmarkServeEmptyTile(b *testing.B) {
	b.StopTimer()
	pathTile := "/tile/11/1086/664.png"
	tempFile, err := ioutil.TempFile("", "example")
	if err != nil {
		fmt.Println("Cannot create temp file:", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())
	data, err := LoadData("/home/nokadmin/projects/go_tile/mock_data/test.osm.pbf", 15, tempFile)
	if err != nil {
		b.Error(err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	// Memory-map the file
	mmapData, mmapFile, err := Mmap(tempFileName)
	if err != nil {
		log.Fatalf("There was an error memory-mapping temp file: %v", err)
	}
	defer syscall.Munmap(*mmapData)
	defer mmapFile.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", pathTile, bytes.NewReader([]byte{}))
		resp := httptest.ResponseRecorder{}
		HandleRenderRequest(&resp, req, time.Second, data, 15, mmapData)
	}
}

// b82ccbeef7bebe3647783ea9b3ed80638d5785cd
//
//	2579786011 ns/op
//	2532791281 ns/op
//
// cec1fd4962a9fd01956f827a7ca74d98560bb6a1
// 18045510356 ns/op
// 13405872770 ns/op
//
//	4064247136 ns/op
//
// c03af9bccb2499f3f0291c8e8aca22f141f06600
//
//		2456439762 ns/op
//	 2536369085 ns/op
func BenchmarkServeFullTile(b *testing.B) {
	b.StopTimer()
	pathTile := "/tile/11/1081/661.png"
	tempFile, err := ioutil.TempFile("", "example")
	if err != nil {
		fmt.Println("Cannot create temp file:", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())
	data, err := LoadData("/home/nokadmin/projects/go_tile/mock_data/test.osm.pbf", 15, tempFile)
	if err != nil {
		b.Error(err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	// Memory-map the file
	mmapData, mmapFile, err := Mmap(tempFileName)
	if err != nil {
		log.Fatalf("There was an error memory-mapping temp file: %v", err)
	}
	defer syscall.Munmap(*mmapData)
	defer mmapFile.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", pathTile, bytes.NewReader([]byte{}))
		resp := httptest.ResponseRecorder{}
		HandleRenderRequest(&resp, req, time.Second, data, 15, mmapData)
	}
}

// b82ccbeef7bebe3647783ea9b3ed80638d5785cd
//
//	8440346178 ns/op
//	  90810744 ns/op
//
// cec1fd4962a9fd01956f827a7ca74d98560bb6a1
//
//		41576416793 ns/op
//	  7996205074 ns/op
//
// c03af9bccb2499f3f0291c8e8aca22f141f06600
//
//		4617123956 ns/op
//	  101012328 ns/op
func BenchmarkServeFullTileZ3(b *testing.B) {
	b.StopTimer()
	pathTile := "/tile/3/4/2.png"
	tempFile, err := ioutil.TempFile("", "example")
	if err != nil {
		fmt.Println("Cannot create temp file:", err)
		os.Exit(1)
	}
	defer os.Remove(tempFile.Name())
	data, err := LoadData("/home/nokadmin/projects/go_tile/mock_data/test.osm.pbf", 15, tempFile)
	if err != nil {
		b.Error(err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()
	// Memory-map the file
	mmapData, mmapFile, err := Mmap(tempFileName)
	if err != nil {
		log.Fatalf("There was an error memory-mapping temp file: %v", err)
	}
	defer syscall.Munmap(*mmapData)
	defer mmapFile.Close()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", pathTile, bytes.NewReader([]byte{}))
		resp := httptest.ResponseRecorder{}
		HandleRenderRequest(&resp, req, time.Second, data, 15, mmapData)
	}
}
