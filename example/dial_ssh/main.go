package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/common"
	"github.com/injoyai/ios/v2/module/ssh"
	"github.com/injoyai/logs"
)

func main() {

	c := client.Redial(ssh.NewDial(&ssh.Config{
		Address:  "192.168.10.9:22",
		User:     "root",
		Password: "root",
		Timeout:  time.Second * 5,
	}))

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
		c.OnDealMessage(func(c *client.Client, msg ios.Acker) {
			fmt.Printf("\r" + string(msg.Bytes()))
		})
	})

	c.Run(context.Background())

}
