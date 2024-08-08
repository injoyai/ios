package serial

import (
	"github.com/goburrow/serial"
)

type (
	Config = serial.Config
)

func Dial(cfg *Config) (*Client, error) {
	serial, err := serial.OpenPort(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{serial.Port}
}

type Client struct {
	*serial.Port
}
