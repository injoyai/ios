package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/module/tcp"
	"github.com/injoyai/logs"
	"time"
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
			c.Event.OnReconnect = client.NewReconnectInterval(time.Second * 3)
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
