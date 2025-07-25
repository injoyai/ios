package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/client/redial"
	"github.com/injoyai/ios/module/common"
	"github.com/injoyai/ios/module/ssh"
	"github.com/injoyai/logs"
	"os"
	"time"
)

func main() {

	c := redial.SSH(&ssh.Config{
		Address:  "192.168.10.9:22",
		User:     "root",
		Password: "root",
		Timeout:  time.Second * 5,
	})

	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			if _, err := c.Write(append(s.Bytes(), '\n')); err != nil {
				logs.Err(err)
				return
			}
		}
	}()

	c.SetOption(func(c *client.Client) {
		c.Logger.SetLevel(common.LevelError)
		c.Event.OnDealMessage = func(c *client.Client, msg ios.Acker) {
			fmt.Printf("\r" + string(msg.Payload()))
		}
	})

	c.Run(context.Background())

}
