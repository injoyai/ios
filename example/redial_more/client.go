package main

import (
	"fmt"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/logs"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func main() {
	go http.ListenAndServe(fmt.Sprintf(":6060"), nil)
	for i := 0; i < 1000; i++ {
		go func() {
			dial.RedialTCP("127.0.0.1:10086", func(c *client.Client) {
				c.Logger.Debug(false)
				c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
					err := w.WriteAny(time.Now().String())
					logs.PrintErr(err)
					return err
				})
			}).Run()
		}()
	}

	select {}
}
