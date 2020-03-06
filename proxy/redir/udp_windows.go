package redir

import (
	"errors"
	"net"
)

func setsockopt(c *net.UDPConn, addr string) (error) {
	return errors.New("Windows not support yet")
}

func getOrigDst(oob []byte, oobn int) (*net.UDPAddr, error) {
	return nil, errors.New("Windows not support yet")
}
