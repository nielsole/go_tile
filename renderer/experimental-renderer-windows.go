//go:build windows
// +build windows

package renderer

import "errors"

func LoadData(path string, maxZ uint32, tempFile *os.File) (*Data, error) {
	return _, errors.New("not implemented on Windows")
}

func Mmap(path string) (*[]byte, *os.File, error) {
	return _, _, errors.New("not implemented on Windows")
}

func HandleRenderRequest(w http.ResponseWriter, r *http.Request, duration time.Duration, data *Data, maxTreeDepth uint32, mmapData *[]byte) {
	return _, errors.New("not implemented on Windows")
}

func Munmap(data *[]byte) error {
	return errors.New("not implemented on Windows")
}
