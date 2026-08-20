package server

import (
	"sync"

	"github.com/injoyai/ios/client"
)

var (
	clientPool = &sync.Pool{New: func() any {
		return client.New(nil)
	}}
)
