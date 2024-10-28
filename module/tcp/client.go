package tcp

import (
	"context"
	"github.com/injoyai/ios"
	"net"
	"time"
)

func NewDial(addr string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := DialTimeout(addr, 0)
		return c, addr, err
	}
}

func Dial(addr string) (net.Conn, error) {
	return DialTimeout(addr, 0)
}

func DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}
