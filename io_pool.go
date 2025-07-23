package ios

import (
	"io"
	"sync"
)

// DefaultFreeFReaderPool 默认使用该读取方式,缓存4KB
var DefaultFreeFReaderPool = NewFreeFReaderPoolKB(4)

func NewFreeFReaderPoolKB(n int) *FreeFReaderPool {
	return &FreeFReaderPool{
		kb: n,
		pool: sync.Pool{New: func() any {
			return make([]byte, 1024*n)
		}},
	}
}

type FreeFReaderPool struct {
	kb   int
	pool sync.Pool
}

func (this *FreeFReaderPool) Get() FreeFReader {
	buffer := this.pool.Get().([]byte)
	return &fromRead{
		buffer:  buffer,
		handler: NewReadFrom(buffer),
		free: func() {
			if cap(buffer) <= this.kb<<10 {
				this.pool.Put(buffer)
			}
		},
	}
}

type fromRead struct {
	buffer  []byte
	handler func(r Reader) ([]byte, error)
	free    func()
	once    sync.Once
}

func (this *fromRead) ReadFrom(r io.Reader) ([]byte, error) {
	return this.handler(r)
}

func (this *fromRead) Free() {
	this.once.Do(this.free)
}
