package server

import (
	"context"
	"fmt"
	"github.com/injoyai/base/maps/timeout"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
	"github.com/injoyai/ios/module/common"
	"sync"
	"time"
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
	s := &Server{
		Event:    &Event{},
		Closer:   safe.NewCloser(),
		Runner:   safe.NewRunnerWithContext(ctx, nil),
		key:      listener.Addr(),
		Logger:   common.NewLogger(),
		Listener: listener,
		Timeout: timeout.New().SetDealFunc(func(key interface{}) error {
			return key.(*client.Client).CloseWithErr(ios.ErrWithTimeout)
		}),
		client: make(map[string]*client.Client),
	}
	s.Runner.SetFunc(s.run)
	s.Timeout.SetTimeout(time.Minute * 3) //3分钟超时(3-检查间隔会超时)
	s.Timeout.SetInterval(time.Minute)    //1分钟检查一次
	s.Closer.SetCloseFunc(func(err error) error {
		//关闭全部客户端,是否关闭?,net包是不关闭已连接的客户端,可以方便热启动
		//s.CloseAllClient(err)
		//服务关闭事件
		s.Logger.Infof("[%s] 关闭服务...\n", listener.Addr())
		if s.Event != nil && s.Event.OnClose != nil {
			s.Event.OnClose(s, err)
		}
		return listener.Close()
	})
	for _, v := range op {
		v(s)
	}
	//放在用户选项之后,方便用户控制是否输出
	s.Logger.Infof("[%s] 开启服务成功...\n", listener.Addr())
	if s.Event.OnOpen != nil {
		s.Event.OnOpen(s)
	}
	return s, nil
}

type Server struct {
	*Event
	*safe.Closer
	*safe.Runner
	key           string
	Logger        common.Logger             //日志
	Listener      ios.Listener              //listener
	Timeout       *timeout.Timeout          //超时机制
	clientOptions []client.Option           //客户端选项
	client        map[string]*client.Client //客户端
	clientMu      sync.RWMutex              //锁
}

// SetClientOption 设置客户端选项
func (this *Server) SetClientOption(op ...client.Option) *Server {
	this.clientOptions = append(this.clientOptions, op...)
	return this
}

// Timer 定时器
func (this *Server) Timer(t time.Duration, f Option) {
	tick := time.NewTicker(t)
	defer tick.Stop()
	for {
		select {
		case <-this.Closer.Done():
			return
		case _, ok := <-tick.C:
			if ok {
				f(this)
			}
		}
	}
}

// GetClient 获取客户端
func (this *Server) GetClient(key string) *client.Client {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	return this.client[key]
}

// GetClientLen 获取客户端数量
func (this *Server) GetClientLen() int {
	return len(this.client)
}

// RangeClient 遍历客户端
func (this *Server) RangeClient(f func(c *client.Client) bool) {
	this.clientMu.RLock()
	defer this.clientMu.RUnlock()
	for _, c := range this.client {
		if !f(c) {
			return
		}
	}
}

// CloseClient 关闭客户端
func (this *Server) CloseClient(key string, err error) {
	c := this.GetClient(key)
	if c != nil {
		c.CloseWithErr(err)
	}
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	delete(this.client, key)
}

// CloseAllClient 关闭全部客户端
func (this *Server) CloseAllClient(err error) {
	this.RangeClient(func(c *client.Client) bool {
		c.CloseWithErr(err)
		return true
	})
	this.clientMu.Lock()
	defer this.clientMu.Unlock()
	this.client = make(map[string]*client.Client)
}

func (this *Server) run(ctx context.Context) error {
	for {
		c, k, err := this.Listener.Accept()
		if err != nil {
			return err
		}
		go func(k string, c ios.ReadWriteCloser) {
			cli := client.NewWithContext(ctx)
			cli.Logger = this.Logger
			cli.SetReadWriteCloser(k, c)
			cli.SetOption(this.clientOptions...)

			//触发服务端连接事件,是否需要2个事件?
			this.Logger.Infof("[%s] 新的客户端连接...\n", cli.GetKey())
			if this.Event != nil && this.Event.OnConnected != nil {
				if err := this.Event.OnConnected(this, cli); err != nil {
					cli.CloseWithErr(err)
					return
				}
			}
			//触发客户端的连接事件,是否需要2个事件?
			if cli.Event != nil && cli.Event.OnConnected != nil {
				if err := cli.Event.OnConnected(cli); err != nil {
					cli.CloseWithErr(err)
					return
				}
			}

			//取消重试,客户端是被连接
			cli.SetRedial(false)
			//取消读取超时机制,取消客户端,实现服务端
			cli.SetReadTimeout(0)

			//设置修改key事件
			onChangeKey := cli.Event.OnKeyChange
			cli.Event.OnKeyChange = func(c *client.Client, oldKey string) {
				if onChangeKey != nil {
					onChangeKey(c, oldKey)
				}
				this.onChangeKey(c, oldKey)
			}
			//把已关闭的客户端从缓存中清除
			onDisconnect := cli.Event.OnDisconnect
			cli.Event.OnDisconnect = func(c *client.Client, err error) {
				if onDisconnect != nil {
					onDisconnect(c, err)
				}
				this.clientMu.Lock()
				delete(this.client, c.GetKey())
				this.clientMu.Unlock()
			}
			//保持读超时状态
			onDealMessage := cli.Event.OnDealMessage
			cli.Event.OnDealMessage = func(c *client.Client, message ios.Acker) {
				this.Timeout.Keep(c)
				if onDealMessage != nil {
					onDealMessage(c, message)
				}
			}
			//保持写超时状态
			onWriteWith := cli.Event.OnWriteWith
			cli.Event.OnWriteWith = func(bs []byte) ([]byte, error) {
				if onWriteWith != nil {
					return onWriteWith(bs)
				}
				this.Timeout.Keep(c)
				return bs, nil
			}

			//设置到缓存
			this.onChangeKey(cli, "")
			cli.Run()

		}(k, c)
	}
}

func (this *Server) onChangeKey(c *client.Client, oldKey string) {
	//判断是否存在老连接,存在则关闭老连接(被挤下线)
	this.clientMu.RLock()
	//查找这个新key是否存在实例,存在则判断2个是否是一个,不是一个则关闭老的那个
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
