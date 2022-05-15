package main

import (
	"errors"
	"io"
)

type SubFileReaderSeeker struct {
	file io.ReadSeeker
	// field needs to be modifiable by Seek and Read but they may not be pointer receivers
	offset    *int64
	minOffset int64
	maxOffset int64
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func (readerSeeker SubFileReaderSeeker) Read(p []byte) (n int, err error) {
	availableLength := min(readerSeeker.maxOffset-*readerSeeker.offset, int64(len(p)))
	n, err = readerSeeker.file.Read(p[:availableLength])
	*readerSeeker.offset += int64(n)
	return n, err
}

func (readerSeeker SubFileReaderSeeker) Seek(offset int64, whence int) (n int64, err error) {
	var requestedFileOffset int64
	switch whence {
	case 0:
		requestedFileOffset = readerSeeker.minOffset + offset
		if requestedFileOffset > readerSeeker.maxOffset {
			return offset, errors.New("out of bounds exception")
		}
	case 1:
		requestedFileOffset = *readerSeeker.offset + offset
		if requestedFileOffset < readerSeeker.minOffset {
			return offset, errors.New("out of bounds exception")
		}
	case 2:
		requestedFileOffset = readerSeeker.maxOffset - offset
		whence = 0
		if requestedFileOffset < readerSeeker.minOffset {
			return offset, errors.New("out of bounds exception")
		}
	}
	n, err = readerSeeker.file.Seek(requestedFileOffset, whence)
	*readerSeeker.offset = int64(n)
	return n - readerSeeker.minOffset, err
}

func NewSubFileReaderSeeker(file io.ReadSeeker, offset int64, length int64) io.ReadSeeker {
	// this smells:
	file.Seek(offset, 0)
	return SubFileReaderSeeker{
		file:      file,
		offset:    &offset,
		minOffset: offset,
		maxOffset: offset + length,
	}
}
