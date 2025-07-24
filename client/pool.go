package client

import (
	"github.com/injoyai/ios"
	"sync"
)

var (
	bufferPool = sync.Pool{New: func() any {
		return ios.NewBufferReader(nil, make([]byte, ios.DefaultBufferSize))
	}}
)
