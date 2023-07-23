//go:build windows
// +build windows

package renderer

import (
	"errors"
	"net/http"
	"os"
	"time"
)

func LoadData(path string, maxZ uint32, tempFile *os.File) (*Data, error) {
	return nil, errors.New("not implemented on Windows")
}

func Mmap(path string) (*[]byte, *os.File, error) {
	return nil, nil, errors.New("not implemented on Windows")
}

func HandleRenderRequest(w http.ResponseWriter, r *http.Request, duration time.Duration, data *Data, maxTreeDepth uint32, mmapData *[]byte) {
	return
}

func Munmap(data *[]byte) error {
	return errors.New("not implemented on Windows")
}
