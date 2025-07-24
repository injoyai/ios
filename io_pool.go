package ios

import (
	"io"
	"sync"
)

// DefaultFReaderPool 默认使用该读取方式,缓存4KB
var DefaultFReaderPool = NewFReaderPool(DefaultBufferSize)

func NewFReaderPool(cap int) *FReaderPool {
	return &FReaderPool{
		cap: cap,
		pool: sync.Pool{New: func() any {
			return make([]byte, cap)
		}},
	}
}

type FReaderPool struct {
	cap  int
	pool sync.Pool
}

func (this *FReaderPool) Get() FreeFReader {
	buffer := this.pool.Get().([]byte)
	return &fromRead{
		buffer: buffer,
		free: func() {
			if cap(buffer) <= this.cap {
				this.pool.Put(buffer)
			}
		},
	}
}

type fromRead struct {
	buffer []byte
	free   func()
	once   sync.Once
}

func (this *fromRead) ReadFrom(r io.Reader) ([]byte, error) {
	n, err := r.Read(this.buffer)
	if err != nil {
		return nil, err
	}
	return this.buffer[:n], nil
}

func (this *fromRead) Free() {
	this.once.Do(this.free)
}
