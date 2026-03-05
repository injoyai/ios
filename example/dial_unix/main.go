package main

import (
	"log"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/dial"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
	"github.com/injoyai/logs"
)

func main() {

	filename := "/tmp/motor/x12.sock"

	go func() {
		<-time.After(time.Second)
		dial.RunUnix(filename, func(c *client.Client) {
			c.OnConnected(func(c *client.Client) error {
				c.GoTimerWriter(time.Minute, func(w ios.MoreWriter) error {
					return w.WriteAny(time.Now())
				})
				return nil
			})
		})
	}()

	err := listen.RunUnix(filename, func(s *server.Server) {
		s.OnClient(func(c *client.Client) {
			c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
				log.Println("1:", string(msg.Bytes()))
			})
		})
	})
	logs.Err(err)

}
