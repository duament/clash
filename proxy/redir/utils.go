package redir

import (
	"net"

	"github.com/Dreamacro/clash/common/pool"
	"github.com/Dreamacro/clash/component/resolver"
	"github.com/Dreamacro/clash/dns"
)

type fakeConn struct {
	net.PacketConn
	origDst  net.Addr
	rAddr    net.Addr
	buf      []byte
	ipMapped *bool
}

func (c *fakeConn) Data() []byte {
	return c.buf
}

// WriteBack opens a new socket binding `lAddr` to wirte UDP packet back
func (c *fakeConn) WriteBack(b []byte, addr net.Addr) (n int, err error) {
	lAddr := addr.(*net.UDPAddr)
	if c.IpMapped() {
		lAddr.IP = c.origDst.(*net.UDPAddr).IP
	}

	tc, err := dialUDP("udp", lAddr, c.rAddr.(*net.UDPAddr))
	if err != nil {
		n = 0
		return
	}
	n, err = tc.Write(b)
	tc.Close()
	return
}

// LocalAddr returns the source IP/Port of UDP Packet
func (c *fakeConn) LocalAddr() net.Addr {
	return c.rAddr
}

func (c *fakeConn) Close() error {
	err := c.PacketConn.Close()
	pool.BufPool.Put(c.buf[:cap(c.buf)])
	return err
}

func (c *fakeConn) IpMapped() bool {
	if c.ipMapped != nil {
		return *c.ipMapped
	}

	resolver := resolver.DefaultResolver.(*dns.Resolver)
	_, ipMapped := resolver.IPToHost(c.origDst.(*net.UDPAddr).IP)
	c.ipMapped = &ipMapped
	return ipMapped
}
