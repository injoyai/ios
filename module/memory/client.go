package memory

import (
	"context"
	"github.com/injoyai/base/chans"
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
		toServer:   chans.NewIO(1),
		fromServer: chans.NewIO(1),
	}

	if timeout <= 0 {
		val.Ch <- c
		return c, nil
	}

	select {
	case val.Ch <- c:
	case <-time.After(timeout):
		return nil, ios.ErrWithTimeout
	}

	return c, nil
}

type Client struct {
	toServer   *chans.IO
	fromServer *chans.IO
}

func (this *Client) Read(p []byte) (int, error) {
	return this.fromServer.Read(p)
}

func (this *Client) Write(p []byte) (int, error) {
	return this.toServer.Write(p)
}

func (this *Client) Close() error {
	this.toServer.Close()
	this.fromServer.Close()
	return nil
}

func (this *Client) sIO() io.ReadWriteCloser {
	return &IO{
		ReadFunc:  this.toServer.Read,
		WriteFunc: this.fromServer.Write,
		CloseFunc: this.Close,
	}
}

type IO struct {
	ios.ReadFunc
	ios.WriteFunc
	ios.CloseFunc
}
