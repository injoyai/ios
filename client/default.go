package client

import (
	"bufio"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/module/common"
)

var (
	// defaultReconnectInterval 默认重连时间间隔
	defaultReconnect = NewReconnectRetreat(time.Second*2, time.Second*32, 2)

	// defaultDealErr 默认处理错误
	defaultDealErr = func(c *Client, err error) error { return common.DealErr(err) }

	// defaultReadFrame 默认读取数据方式
	defaultReadFrame = ios.NewFRead4KB()

	// DefaultReaderPool 默认bufio连接池
	DefaultReaderPool = NewPool(1000, func() *bufio.Reader {
		return bufio.NewReaderSize(nil, 4096)
	})
)

func NewPool(max int, new func() *bufio.Reader) *Pool {
	return &Pool{
		ch:  make(chan *bufio.Reader, max),
		new: new,
	}
}

// Pool 内存复用池,用来替代sync.Pool
// 用来解决sync.Pool内存不能被系统回收的问题
type Pool struct {
	ch  chan *bufio.Reader
	new func() *bufio.Reader
}

// Get 获取对象,如果没有会新申请
func (this *Pool) Get() *bufio.Reader {
	select {
	case buf := <-this.ch:
		return buf
	default:
		return this.new()
	}
}

// Put 回收变量,需要变量的内存地址,方便解除引用让系统回收
// 每个变量的申明都会有一个固定的内存地址,注意变量,指针,内存地址的区别
func (this *Pool) Put(buf *bufio.Reader) {
	if buf == nil {
		return
	}
	buf.Reset(nil)
	select {
	case this.ch <- buf:
	default:
		buf = nil
	}
}
