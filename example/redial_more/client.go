package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/client/redial"
)

func main() {
	go http.ListenAndServe(fmt.Sprintf(":6060"), nil)
	for i := 0; i < 10000; i++ {
		go func() {
			redial.TCP("127.0.0.1:10086", func(c *client.Client) {
				c.Logger.Enable(false)
				c.OnConnected(func(c *client.Client) {
					c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
						return w.WriteAny(time.Now().String())
					})
					go func() {
						<-time.After(time.Second * 6)
						c.Close()
					}()
				})
			}).Run(context.Background())
		}()
	}

	select {}
}
