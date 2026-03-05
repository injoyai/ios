package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/server"
	"github.com/injoyai/ios/v2/server/listen"
)

func main() {
	go http.ListenAndServe(fmt.Sprintf(":6070"), nil)
	listen.RunTCP(10086, func(s *server.Server) {
		go func() {
			for {
				<-time.After(time.Second * 3)
				log.Println("客户端数量:", s.GetClientLen())
			}
		}()
		//s.Logger.Debug(false)
		s.OnClient(func(c *client.Client) {
			c.OnConnected(func(c *client.Client) error {
				c.Logger.Enable(false)
				//go func() {
				//	<-time.After(time.Second * 1)
				//	c.CloseWithErr(errors.New("手动断开"))
				//}()
				return nil
			})
		})
	})
}
