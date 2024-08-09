package serial

import (
	"github.com/goburrow/serial"
)

type (
	Config = serial.Config
)

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
