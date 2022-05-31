package main

import (
	"encoding/binary"
	"errors"
	"net"
	"time"
)

type Response struct {
	Version uint32
	Success uint32
	// There are actually more fields being sent, but do we care?
}

type Request struct {
	Version     uint32
	CmdPriority uint32
	X           uint32
	Y           uint32
	Z           uint32
	Map         [44]byte
}

func requestRender(x, y, z uint32, map_name, renderd_sock_path string, renderd_timeout time.Duration) error {
	c, err := net.Dial("unix", renderd_sock_path)
	if err != nil {
		return err
	}
	defer c.Close()
	if renderd_timeout > 0 {
		c.SetDeadline(time.Now().Add(renderd_timeout))
	}
	request := Request{
		Version:     3,
		CmdPriority: 5,
		X:           x,
		Y:           y,
		Z:           z,
	}
	copy(request.Map[:], []byte(map_name))
	if err := binary.Write(c, binary.LittleEndian, request); err != nil {
		return err
	}
	response := make([]byte, 56)
	n, err := c.Read(response)
	if err != nil {
		return err
	}
	if n != len(response) {
		return errors.New("could not read response. Unexpected number of bytes")
	}
	return nil
}
