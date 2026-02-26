package main

import (
	"context"

	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/ios/v2/module/tcp"
	"github.com/injoyai/logs"
)

func main() {

	err := client.New(
		tcp.NewDial(":10086"),
		client.WithRedial(),
		client.WithDebug(),
	).Run(context.Background())
	logs.Err(err)
}
