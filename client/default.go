package client

import (
	"bufio"
	"sync"
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
	DefaultReaderPool = sync.Pool{
		New: func() interface{} {
			return bufio.NewReaderSize(nil, 4096)
		},
	}
)
