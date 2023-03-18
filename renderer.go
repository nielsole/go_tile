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

func getSocketConnection(renderd_socket string) (net.Conn, error) {
	renderd_socket_type := getSocketType(renderd_socket)
	if renderd_socket_type == "tcp" {
		tcp_addr, _ := net.ResolveTCPAddr("tcp", renderd_socket)
		return net.DialTCP("tcp", nil, tcp_addr)
	} else {
		return net.Dial("unix", renderd_socket)
	}
}

func requestRender(x, y, z uint32, map_name, renderd_socket string, renderd_timeout time.Duration, priority int) error {
	c, err := getSocketConnection(renderd_socket)
	if err != nil {
		return err
	}
	defer c.Close()
	if renderd_timeout > 0 {
		c.SetDeadline(time.Now().Add(renderd_timeout))
	}
	request := Request{
		Version:     3,
		CmdPriority: uint32(priority),
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
