package main

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	z, x, y, err := parsePath("/tile/4/3/2.png")
	if err != nil {
		t.Error(err)
	}
	if z != 4 || x != 3 || y != 2 {
		t.Fail()
	}
}

// cpu: Intel(R) Core(TM) i5-6300U CPU @ 2.40GHz
// BenchmarkPngRead-4                 88027             13260 ns/op
// BenchmarkWriteTileResponse-4       55233             21662 ns/op

// 22 microsecond and 13microsecond avg. response time is really nothing worth optimizing
func BenchmarkPngRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readPngTile("mock_data/0.meta", 0)
	}
}

func TestWriteTileResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/0.meta", 0, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 200 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestWriteTileResponse404(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/404.meta", 0, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 404 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestWriteTileResponseOutOfBounds(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/0.meta", 65, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 500 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}
