package memory

import (
	"bytes"
	"context"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"io"
	"time"
)

func NewDial(key string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(key)
		return c, key, err
	}
}

func Dial(key string) (*Client, error) {
	return DialTimeout(key, 0)
}

func DialTimeout(key string, timeout time.Duration) (*Client, error) {

	val := manage.MustGet(key)
	if val == nil {
		return nil, ios.ErrRemoteOff
	}

	c := &Client{
		output: bytes.NewBuffer(nil),
		input:  bytes.NewBuffer(nil),
		Closer: safe.NewCloser()}

	if timeout <= 0 {
		val.(*Server).Ch <- c
		return c, nil
	}

	select {
	case val.(*Server).Ch <- c:
	case <-time.After(timeout):
		return nil, ios.ErrWithTimeout
	}

	return c, nil
}

var _ ios.AReadWriteCloser = (*Client)(nil)

type Client struct {
	output *bytes.Buffer
	input  *bytes.Buffer
	*safe.Closer
	Handler func(r io.Reader) ([]byte, error)
}

func (this *Client) ReadAck() (ios.Acker, error) {
	if this.Closed() {
		return nil, this.Closer.Err()
	}
	if this.Handler == nil {
		buf := make([]byte, 1024*4)
		this.Handler = ios.NewReadWithBuffer(buf)
	}
	bs, err := this.Handler(this.output)
	return ios.Ack(bs), err
}

func (this *Client) Write(p []byte) (int, error) {
	if this.Closed() {
		return 0, ios.ErrWriteClosed
	}
	return this.input.Write(p)
}

func (this *Client) Close() error {
	return this.Closer.CloseWithErr(io.EOF)
}

func (this *Client) sRead(p []byte) (int, error) {
	if this.Closed() {
		return 0, this.Closer.Err()
	}
	return this.input.Read(p)
}

func (this *Client) sWrite(p []byte) (int, error) {
	if this.Closed() {
		return 0, ios.ErrWriteClosed
	}
	return this.output.Write(p)
}

func (this *Client) sIO() io.ReadWriteCloser {
	return &IO{
		ReadFunc:  this.sRead,
		WriteFunc: this.sWrite,
		CloseFunc: this.Close,
	}
}

type IO struct {
	ios.ReadFunc
	ios.WriteFunc
	ios.CloseFunc
}
