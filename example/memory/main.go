package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"github.com/injoyai/logs"
	"time"
)

func main() {

	go listen.RunMemory("test", func(s *server.Server) {
		//s.Logger.Debug(false)
		s.SetClientOption(func(c *client.Client) {
			c.OnDealMessage = func(c *client.Client, msg ios.Acker) {
				_, err := c.Write(msg.Payload())
				logs.PrintErr(err)
			}
		})
	})

	<-time.After(time.Second)

	c, err := dial.Memory("test", func(c *client.Client) {
		c.Logger.Debug(false)
		c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
			return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
		})
	})
	logs.PanicErr(err)
	c.Run()

}
