package main

import (
	"context"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/logs"
)

func main() {
	ctx := context.Background()

	if false {

		c := client.Redial(tcp.NewDial(":10086"), func(c *client.Client) {
			c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
			})
		})

		c.Run(ctx)
	}

	{

		c := client.New(tcp.NewDial(":10086"), func(c *client.Client) {
			c.OnReconnect(client.NewReconnectInterval(time.Second * 3))
			c.SetKey(":10087")
			c.SetRedial()
			//c.SetReadTimeout(time.Second * 10)
			c.GoTimerWriter(time.Second*3, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now().Format("2006-01-02 15:04:05"))
			})
		})
		logs.Err(c.Run(ctx))
	}

}
