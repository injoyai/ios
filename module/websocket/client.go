package websocket

import (
	ws "github.com/gorilla/websocket"
	"github.com/injoyai/ios"
)

var _ ios.AReadWriteCloser = &Client{}

type Client struct {
	*ws.Conn
}

func (c Client) ReadAck() (ios.Acker, error) {
	//TODO implement me
	panic("implement me")
}

func (this *Client) Write(p []byte) (n int, err error) {

	//this.Conn.WriteMessage(websocket.TextMessage, p)

	//TODO implement me
	panic("implement me")
}
