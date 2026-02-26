package client

import "github.com/injoyai/ios/v2"

var (
	DefaultPool = NewPool(1000, func() *ios.BufferReader {
		return ios.NewBufferReader(nil, make([]byte, ios.DefaultBufferSize))
	})
)

func NewPool(max int, new func() *ios.BufferReader) *Pool {
	return &Pool{
		ch:   make(chan *ios.BufferReader, max),
		new:  new,
		item: new(),
	}
}

// Pool 内存复用池,用来替代sync.Pool
// 用来解决sync.Pool内存不能被系统回收的问题
type Pool struct {
	ch   chan *ios.BufferReader
	new  func() *ios.BufferReader
	item *ios.BufferReader
}

// Get 获取对象,如果没有会新申请
func (this *Pool) Get() *ios.BufferReader {
	select {
	case buf := <-this.ch:
		return buf
	default:
		return this.new()
	}
}

// Put 回收变量,需要变量的内存地址,方便解除引用让系统回收
// 每个变量的申明都会有一个固定的内存地址,注意变量,指针,内存地址的区别
func (this *Pool) Put(buf *ios.BufferReader) {
	if buf == nil {
		return
	}
	if buf.Cap() > this.item.Cap() {
		buf = nil
		return
	}
	buf.Clear()
	select {
	case this.ch <- buf:
	default:
		buf = nil
	}
}
