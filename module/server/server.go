package server

import (
	"context"
	"fmt"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/module/client"
	"sync"
)

type Option func(s *Server)

func Run(listen ios.ListenFunc, op ...Option) error {
	s, err := New(listen, op...)
	if err != nil {
		return err
	}
	return s.Run()
}

func New(listen ios.ListenFunc, op ...Option) (*Server, error) {
	return NewWithContext(context.Background(), listen, op...)
}

func NewWithContext(ctx context.Context, listen ios.ListenFunc, op ...Option) (*Server, error) {
	listener, err := listen()
	if err != nil {
		return nil, err
	}
	defaultLogger.Infof("[%s] 开启服务成功...\n", listener.Addr())
	s := &Server{
		Logger:   defaultLogger,
		key:      listener.Addr(),
		Closer:   safe.NewCloser(),
		Runner:   safe.NewRunnerWithContext(ctx, nil),
		listener: listener,
		timeout:  safe.NewRunnerWithContext(ctx, nil),
		client:   make(map[string]*client.Client),
	}
	s.Runner.SetFunc(s.run)
	for _, v := range op {
		v(s)
	}
	return s, nil
}

type Server struct {
	Logger
	*safe.Closer
	*safe.Runner
	key           string
	listener      ios.Listener              //listener
	timeout       *safe.Runner              //超时机制
	clientOptions []client.Option           //客户端选项
	client        map[string]*client.Client //客户端
	clientMu      sync.RWMutex              //锁
}

func (this *Server) SetOption(op ...client.Option) *Server {
	this.clientOptions = append(this.clientOptions, op...)
	return this
}

func (this *Server) run(ctx context.Context) error {
	for {
		c, k, err := this.listener.Accept()
		if err != nil {
			return err
		}
		go func(k string, c ios.ReadWriteCloser) {
			cli := client.NewWithContext(ctx)
			cli.SetReadWriteCloser(k, c)
			cli.SetOption(this.clientOptions...)

			cli.Infof("[%s] 新的客户端连接...\n", cli.GetKey())
			if cli.Event != nil && cli.Event.OnConnected != nil {
				if err := cli.Event.OnConnected(cli); err != nil {
					cli.CloseWithErr(err)
					return
				}
			}

			//设置修改key事件
			cli.Event.OnKeyChange = this.onChangeKey
			//取消重试,客户端是被连接
			cli.SetRedial(false)
			//取消读取超时机制,取消客户端,实现服务端
			cli.SetReadTimeout(0)

			//设置到缓存
			this.onChangeKey(cli, "")
			cli.Run()

		}(k, c)
	}
}

func (this *Server) onChangeKey(c *client.Client, oldKey string) {
	//判断是否存在老连接,存在则关闭老连接(被挤下线)
	this.clientMu.RLock()
	old, ok := this.client[c.GetKey()]
	this.clientMu.RUnlock()
	if ok && old != c {
		old.CloseWithErr(fmt.Errorf("重复标识(%s),关闭老客户端", old.GetKey()))
	}
	//保存到缓存中
	this.clientMu.Lock()
	delete(this.client, oldKey)
	this.client[c.GetKey()] = c
	this.clientMu.Unlock()
}
