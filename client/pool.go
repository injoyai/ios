package client

import "github.com/injoyai/ios"

var (
	bufferReadePool = NewPool(1000, func() *ios.BufferReader {
		return ios.NewBufferReader(nil, make([]byte, ios.DefaultBufferSize))
	})
)

func NewPool[T any](max int, new func() *T) *Pool[T] {
	return &Pool[T]{
		ch:  make(chan *T, max),
		new: new,
	}
}

// Pool 内存复用池,用来替代sync.Pool
// 用来解决sync.Pool内存不能被系统回收的问题
type Pool[T any] struct {
	ch  chan *T
	new func() *T
}

// Get 获取对象,如果没有会新申请
func (this *Pool[T]) Get() *T {
	select {
	case buf := <-this.ch:
		return buf
	default:
		return this.new()
	}
}

// Put 回收变量,需要变量的内存地址,方便解除引用让系统回收
// 每个变量的申明都会有一个固定的内存地址,注意变量,指针,内存地址的区别
func (this *Pool[T]) Put(buf *T) {
	if buf == nil {
		return
	}
	select {
	case this.ch <- buf:
	default:
		//解除引用,然后让GC回收
		*buf = *new(T)
	}
}
