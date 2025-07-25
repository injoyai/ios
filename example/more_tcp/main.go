package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"time"
)

/*
测试很多tcp连接的性能,
是否有内存泄漏,是否CPU占用过高

 1. 1个服务端,1万个客户端,客户端写入频率1次/s,服务端每条数据都响应,CPU型号为(i7-7700)
    内存占用725.8MB,CPU使用率(5.0%~17.6%)
*/
func main() {

	go func() {
		err := listen.RunTCP(20001, func(s *server.Server) {
			s.SetClientOption(func(c *client.Client) {
				c.Logger.Debug(false)
				c.Event.OnDealMessage = func(c *client.Client, msg ios.Acker) {
					c.Write(msg.Payload())
				}
			})
			go s.Timer(time.Second*5, func(s *server.Server) {
				logs.Debug("客户端连接数量:", s.GetClientLen())
			})
		})
		panic(err)
	}()

	for i := 0; i < 10000; i++ {
		go func() {
			redial.TCP("127.0.0.1:20001", func(c *client.Client) {
				c.Logger.Debug(false)
				c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
					return w.WriteAny(time.Now().String())
				})
			}).Run(context.Background())
		}()
	}

	select {}

}
