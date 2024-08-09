package dial

import (
	"context"
	"errors"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"time"
)

func Dial(f Handler) (*client.Client, error) {

	return &client.Client{}, nil
}

type Handler func(ctx context.Context) (ios.ReadeWriteCloser, string, error)

func WithMust(h Handler) Handler {
	return func(ctx context.Context) (ios.ReadeWriteCloser, string, error) {
		if h == nil {
			return nil, "", errors.New("handler is nil")
		}
		wait := time.Second * 0
		for i := 0; ; i++ {
			select {
			case <-ctx.Done():
				return nil, "", ctx.Err()
			case <-time.After(wait):
				c, s, err := h(ctx)
				if err == nil {
					return c, s, nil
				}
				if wait < time.Second {
					wait = time.Second
				} else if wait <= time.Second*16 {
					wait *= 2
				}
			}
		}
	}
}
