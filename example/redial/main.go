package main

import (
	"errors"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {

	go func() {
		listen.RunTCP(10086, func(s *server.Server) {
			s.Logger.Debug(false)
			s.SetClientOption(func(c *client.Client) {
				c.Event.OnConnected = func(c *client.Client) error {
					logs.Debug("新的客户端连接")
					c.Logger.Debug(false)
					go func() {
						<-time.After(time.Second * 5)
						c.CloseWithErr(errors.New("手动断开"))
					}()
					return nil
				}
			})
		})
	}()

	c := dial.RedialTCP("127.0.0.1:10086")
	go func() {
		logs.Err(c.Run())
	}()
	go func() {
		<-time.After(time.Second * 10)
		c.Stop()
	}()
	<-c.Runner.Done()
	logs.Debug("结束客户端生命周期")
	<-time.After(time.Second * 10)
	c.Run()

}
