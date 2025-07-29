package server

import (
	"github.com/injoyai/ios/client"
	"sync"
)

var (
	clientPool = &sync.Pool{New: func() any {
		return client.New(nil)
	}}
)
