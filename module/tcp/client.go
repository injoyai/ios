package tcp

import (
	"github.com/injoyai/ios"
	"io"
	"net"
	"time"
)

var _ ios.AReadWriteCloser = (*Client)(nil)

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
		buf := make([]byte, 1024*4)
		this.Handler = ios.NewReadWithBuffer(buf)
	}
	bs, err := this.Handler(this)
	return ios.Ack(bs), err
}
