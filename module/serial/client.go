package serial

import (
	"context"

	"github.com/goburrow/serial"
	"github.com/injoyai/ios/v2"
)

type Config = serial.Config

func NewDial(cfg *Config) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(cfg)
		return c, cfg.Address, err
	}
}

func Dial(cfg *Config) (serial.Port, error) {
	return serial.Open(cfg)
}
