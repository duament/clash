package redir

import (
	"errors"
	"net"
)

func dialUDP(network string, lAddr *net.UDPAddr, rAddr *net.UDPAddr) (*net.UDPConn, error) {
	return nil, errors.New("Windows not support yet")
}
