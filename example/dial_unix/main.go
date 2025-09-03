package main

import (
	"context"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/dial"
	"github.com/injoyai/ios/server"
	"github.com/injoyai/ios/server/listen"
	"log"
	"time"
)

func main() {

	filename := "/tmp/test.sock"

	s, err := listen.Unix(filename, func(s *server.Server) {
		s.SetClientOption(func(c *client.Client) {
			c.OnDealMessage = func(c *client.Client, msg ios.Acker) {
				log.Println("1:", string(msg.Payload()))
			}
		})
	})
	if err != nil {
		panic(err)
	}
	go s.Run(context.Background())

	<-time.After(time.Second)
	go func() {
		err := listen.RunUnix(filename, func(s *server.Server) {
			s.SetClientOption(func(c *client.Client) {
				c.OnDealMessage = func(c *client.Client, msg ios.Acker) {
					log.Println("2:", string(msg.Payload()))
				}
			})
		})
		log.Println(err)
	}()

	//<-time.After(time.Second)

	go func() {
		dial.RunUnix(filename, func(c *client.Client) {
			c.OnConnected = func(c *client.Client) error {
				c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
					return w.WriteAny(time.Now())
				})
				return nil
			}
		})
	}()

	err = dial.RunUnix(filename, func(c *client.Client) {
		c.OnConnected = func(c *client.Client) error {
			c.GoTimerWriter(time.Second, func(w ios.MoreWriter) error {
				return w.WriteAny(time.Now())
			})
			return nil
		}
	})
	log.Println(err)
}
