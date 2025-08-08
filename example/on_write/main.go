package main

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/server/listen"
	"time"
)

func main() {

	go listen.RunTCP(12658)

	redial.RunTCP(":12658", func(c *client.Client) {
		c.Logger.Debug(false)
		c.OnConnected = func(c *client.Client) error {
			c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now())
			})
			return nil
		}
		c.OnWrite = func(f func() error) error {
			//把写入数据重置掉
			_, err := c.Origin().Write([]byte("hello"))
			return err
		}
	})

}
