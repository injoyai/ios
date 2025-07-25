package client

import (
	"github.com/injoyai/ios"
	"sync"
)

var (
	bufferReadePool = sync.Pool{New: func() any {
		return ios.NewBufferReader(nil, make([]byte, ios.DefaultBufferSize))
	}}
)
