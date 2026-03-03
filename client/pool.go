package client

import (
	"bufio"
	"sync"
)

var DefaultReaderPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReaderSize(nil, 4096)
	},
}
