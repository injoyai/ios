package tcp

import (
	"context"
	"github.com/injoyai/ios"
	"io"
	"net"
	"time"
)

var _ ios.AReadWriteCloser = (*Client)(nil)

func NewDial(addr string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := DialTimeout(addr, 0)
		return c, addr, err
	}
}

func Dial(addr string) (*Client, error) {
	return DialTimeout(addr, 0)
}

func DialTimeout(addr string, timeout time.Duration) (*Client, error) {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	return &Client{
		Conn: c,
	}, nil
}

type Client struct {
	net.Conn
	Handler func(r io.Reader) ([]byte, error)
}

func (this *Client) ReadAck() (ios.Acker, error) {
	if this.Handler == nil {
		f := ios.NewReadWithBuffer(make([]byte, 1024*4))
		this.Handler = func(r io.Reader) ([]byte, error) {
			return f(r)
		}
	}
	bs, err := this.Handler(this.Conn)
	return ios.Ack(bs), err
}
