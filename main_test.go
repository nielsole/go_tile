package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nielsole/go_tile/utils"
)

// cpu: Intel(R) Core(TM) i5-6300U CPU @ 2.40GHz
// BenchmarkPngRead-4                 88027             13260 ns/op
// BenchmarkWriteTileResponse-4       55233             21662 ns/op

// 22 microsecond and 13microsecond avg. response time is really nothing worth optimizing
func BenchmarkPngRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		readTile("mock_data/0.meta", 0)
	}
}

func BenchmarkParsePath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _, _, err := utils.ParsePath("/tile/4/3/2.png")
		if err != nil {
			b.Error(err)
		}
	}
}

func TestWriteTileResponse(t *testing.T) {

	// Prepare test request and response
	req := httptest.NewRequest("GET", "/tile/0/0/0.png", nil)
	w := httptest.NewRecorder()

	// Call writeTileResponse function
	modTime := time.Now()
	err := writeTileResponse(w, req, "mock_data/0.meta", 0, modTime, "png")
	if err != nil {
		t.Fatal(err)
	}

	// Check response status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d but got %d", http.StatusOK, w.Code)
	}

	// Check response header
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("unexpected content type: %s", w.Header().Get("Content-Type"))
	}

	if w.Body.Len() < 100 {
		t.Errorf("expected body,but got %q", w.Body.String())
	}

	// Check response caching headers
	expectedCacheControl := "no-cache"
	if w.Header().Get("Cache-Control") != expectedCacheControl {
		t.Errorf("unexpected Cache-Control header: %s", w.Header().Get("Cache-Control"))
	}
	expectedLastModified := modTime.UTC().Format(http.TimeFormat)
	if w.Header().Get("Last-Modified") != expectedLastModified {
		t.Errorf("unexpected Last-Modified header: %s", w.Header().Get("Last-Modified"))
	}
}

func TestWriteTileResponse404(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/404.meta", 0, time.Now(), "png"); err != nil {
		t.Error(err)
	}
	if resp.Code != 404 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestWriteTileResponseOutOfBounds(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/0.meta", 65, time.Now(), "png"); err != nil {
		t.Error(err)
	}
	if resp.Code != 500 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}
