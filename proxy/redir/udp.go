package redir

import (
	"errors"
	"net"

	adapters "github.com/Dreamacro/clash/adapters/inbound"
	"github.com/Dreamacro/clash/common/pool"
	"github.com/Dreamacro/clash/component/socks5"
	C "github.com/Dreamacro/clash/constant"
	"github.com/Dreamacro/clash/log"
	"github.com/Dreamacro/clash/tunnel"
)

type RedirUDPListener struct {
	net.PacketConn
	address string
	closed  bool
}

func NewRedirUDPProxy(addr string) (*RedirUDPListener, error) {
	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, err
	}

	rl := &RedirUDPListener{l, addr, false}

	c, ok := l.(*net.UDPConn)
	if !ok {
		return nil, errors.New("only work with UDP connection")
	}

	err = setsockopt(c, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		oob := make([]byte, 1024)
		for {
			var origDst *net.UDPAddr
			buf := pool.BufPool.Get().([]byte)

			n, oobn, _, remoteAddr, err := c.ReadMsgUDP(buf, oob)
			if err != nil {
				pool.BufPool.Put(buf[:cap(buf)])
				if rl.closed {
					break
				}
				continue
			}

			origDst, err = getOrigDst(oob, oobn)
			if err != nil {
				log.Debugln("redir udp: getorigDst error: %s", err)
				continue
			}
			handleRedirUDP(l, buf[:n], remoteAddr, origDst)
		}
	}()

	return rl, nil
}

func (l *RedirUDPListener) Close() error {
	l.closed = true
	return l.PacketConn.Close()
}

func (l *RedirUDPListener) Address() string {
	return l.address
}

func handleRedirUDP(pc net.PacketConn, buf []byte, addr *net.UDPAddr, origDst *net.UDPAddr) {
	var origDstAddr net.Addr = origDst
	var addrAddr net.Addr = addr
	target := socks5.ParseAddrToSocksAddr(origDstAddr)

	packet := &fakeConn{
		PacketConn: pc,
		rAddr:      addrAddr,
		buf:        buf,
	}
	tunnel.AddPacket(adapters.NewPacket(target, packet, C.REDIR))
}
