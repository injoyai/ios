package serial

import (
	"context"
	"github.com/goburrow/serial"
	"github.com/injoyai/ios"
)

type Config = serial.Config

func NewDial(cfg *Config) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(cfg)
		return c, cfg.Address, err
	}
}

func Dial(cfg *Config) (*Client, error) {
	port, err := serial.Open(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{port}, nil
}

type Client struct {
	serial.Port
}
