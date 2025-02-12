package sse

import (
	"context"
	"errors"
	"github.com/injoyai/ios"
	"io"
	"net/http"
)

func NewDial(url string, body io.Reader) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(url, body)
		return c, url, err
	}
}

func Dial(url string, body io.Reader) (*Client, error) {
	c := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头，表明客户端接受 text/event-stream
	req.Header.Set("Accept", "text/event-stream")

	//发起请求
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return &Client{ReadCloser: resp.Body}, nil
}

type Client struct {
	io.ReadCloser
}

func (this *Client) Write(p []byte) (int, error) {
	return 0, errors.New("not support")
}
