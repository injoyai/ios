package main

import (
	"errors"
	"fmt"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
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
		s.Logger.Debug(false)
		s.SetClientOption(func(c *client.Client) {
			c.Event.OnConnected = func(c *client.Client) error {
				//c.Logger.Debug(false)
				go func() {
					<-time.After(time.Second * 1)
					c.CloseWithErr(errors.New("手动断开"))
				}()
				return nil
			}
		})
	})
}
