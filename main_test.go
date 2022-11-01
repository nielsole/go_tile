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

func TestReadPngTile(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/0.meta", 0, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 200 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestReadPngTile404(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/404.meta", 0, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 404 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}

func TestReadPngTileOutOfBounds(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/", bytes.NewReader([]byte{}))
	resp := httptest.ResponseRecorder{}
	if err := writeTileResponse(&resp, req, "mock_data/0.meta", 65, time.Now()); err != nil {
		t.Error(err)
	}
	if resp.Code != 500 {
		t.Errorf("Unexpected response code: %d", resp.Code)
	}
}
