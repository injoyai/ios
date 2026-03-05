package main

import (
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/frame"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
)

func main() {

	go listen.RunTCP(10099, func(s *server.Server) {
		s.Logger.Enable(false)
		s.OnClient(func(c *client.Client) {
			c.WithFrame(frame.Default)
		})
	})

	<-time.After(time.Second)

	redial.RunTCP(":10099", func(c *client.Client) {
		c.WithFrame(frame.Default)
		c.OnConnected(func(c *client.Client) error {
			c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().String())
			})
			return nil
		})
	})

}
