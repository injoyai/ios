package main

import (
	"time"

	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/frame"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
)

func main() {

	go listen.RunTCP(10099, func(s *server.Server) {
		s.Logger.Debug(false)
		s.SetClientOption(func(c *client.Client) {
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
