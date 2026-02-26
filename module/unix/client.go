package unix

import (
	"context"
	"net"

	"github.com/injoyai/ios/v2"
)

func NewDial(addr string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		var d net.Dialer
		c, err := d.DialContext(ctx, "unix", addr)
		return c, addr, err
	}
}
