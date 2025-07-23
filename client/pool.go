package client

import (
	"github.com/injoyai/ios"
	"sync"
)

var (
	bufferPool = sync.Pool{New: func() any {
		x := ios.NewBufferReader(nil, make([]byte, ios.DefaultBufferSize))
		x.Reset()
		return x
	}}
)
