package main

import (
	"context"
	"errors"
	"time"

	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	go func() {
		listen.RunTCP(10086, func(s *server.Server) {
			s.Logger.Enable(false)
			s.OnConnected(func(c *client.Client) {
				c.OnConnected(func(c *client.Client) error {
					logs.Debug("新的客户端连接")
					c.Logger.Enable(false)
					go func() {
						<-time.After(time.Second * 5)
						c.CloseWithErr(errors.New("手动断开"))
					}()
					return nil
				})
			})
		})
	}()

	c := redial.TCP("127.0.0.1:10086")
	go func() {
		logs.Err(c.Run(context.Background()))
	}()
	go func() {
		<-time.After(time.Second * 10)
		c.Stop()
	}()
	<-c.Done()
	logs.Debug("结束客户端生命周期")
	<-time.After(time.Second * 10)
	c.Run(context.Background())

}
