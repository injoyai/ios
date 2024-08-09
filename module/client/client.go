package client

import (
	"context"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"time"
)

type (
	Option  func(c *Client)
	Message = bytes.Entity
)

func Run(op ...Option) error {
	return New(op...).Run()
}

func New(op ...Option) *Client {
	ctx := context.Background()
	c := &Client{
		Key:              "",
		ReadeWriteCloser: nil,
		tag:              maps.NewSafe(),
		Logger:           nil,
		Base:             Base{CreateTime: time.Now()},
		Event:            Event{},
		Closer:           safe.NewCloser(),
		Runner:           safe.NewRunnerWithContext(ctx, nil),
	}

	c.Closer.SetCloseFunc(func(err error) error {
		c.Runner.Stop(true)
		*c = *New()
		return nil
	})
	c.Runner.SetFunc(c.run)

	for _, f := range op {
		f(c)
	}

	return c
}

type Client struct {
	Key                  string     //自定义标识
	ios.ReadeWriteCloser            //IO实例
	tag                  *maps.Safe //标签,用于记录连接的一些信息

	Logger       //日志
	Base         //基本信息
	Event        //事件
	*safe.Closer //关闭
	*safe.Runner //运行
}

func (this *Client) run(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			////读取数据
			//ack, err := this.ReadAck()
			//if err != nil || len(ack.Payload()) == 0 {
			//	return err
			//}
			//
			////处理数据
			//for _, dealFunc := range this.dealFunc {
			//	if dealFunc != nil && dealFunc(this, ack.Payload()) {
			//		ack.Ack()
			//	}
			//}

		}
	}

}

type Base struct {
	CreateTime time.Time //创建时间,对象创建时间,重连不会改变
	DialTime   time.Time //连接时间,每次重连会改变
	ReadTime   time.Time //本次连接,最后读取到数据的时间
	WriteTime  time.Time //本次连接,最后写入数据时间
}
