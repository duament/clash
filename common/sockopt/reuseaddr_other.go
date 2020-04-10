// +build !linux

package sockopt

import (
	"net"
)

func Reuseaddr(c *net.UDPConn) (err error) {
	return
}
