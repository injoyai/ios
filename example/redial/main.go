package main

import (
	"errors"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/ios/module/client/dial"
	"github.com/injoyai/ios/module/server"
	"github.com/injoyai/ios/module/server/listen"
	"time"
)

func main() {

	go func() {
		listen.RunTCP(10086, func(s *server.Server) {
			s.Logger.Debug(false)
			s.SetOption(func(c *client.Client) {
				c.Event.OnConnected = func(c *client.Client) error {
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

	dial.RedialTCP("127.0.0.1:10086", func(c *client.Client) {

	}).Run()

}
