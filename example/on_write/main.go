package main

import (
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
	"github.com/injoyai/ios/v2/server/listen"
)

func main() {

	go listen.RunTCP(12658)

	redial.RunTCP(":12658", func(c *client.Client) {
		c.Logger.Enable(false)
		c.OnConnected(func(c *client.Client) error {
			c.GoTimerWriter(time.Second*5, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now())
			})
			return nil
		})
		c.OnWrite(func(f func() error) error {
			//把写入数据重置掉
			_, err := c.Origin().Write([]byte("hello"))
			return err
		})
	})

}
