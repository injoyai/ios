package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/module/frame"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"time"
)

func main() {

	go listen.RunTCP(10089, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.Event.WithFrame(frame.Entity)
		})
	})

	dial.RedialTCP(":10089", func(c *client.Client) {
		c.Event.WithFrame(frame.Entity)
		c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().String())
		})
	}).Run()

}
