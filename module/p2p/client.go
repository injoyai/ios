package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/safe"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/pion/webrtc/v4"
)

func NewDial(relay *client.Client, target string) ios.DialFunc {
	return func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		c, err := Dial(relay, target)
		return c, target, err
	}
}

func Dial(relay *client.Client, target string) (*Client, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		return nil, err
	}

	//创建数据通道,v1版本先默认创建一个数据通道,先执行这个才能收集ICE信息
	dc, err := conn.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}

	offer, err := conn.CreateOffer(nil)
	if err != nil {
		return nil, err
	}

	err = conn.SetLocalDescription(offer)
	if err != nil {
		return nil, err
	}

	// ✅ 用这个通道注册等待 ICE 收集完成
	<-webrtc.GatheringCompletePromise(conn)

	//监听候选地址,向中继服务器发送,由中继转发给目标
	conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		relay.Write(Message{Type: ICE, To: target, Data: candidate.ToJSON().Candidate}.Bytes())
	})

	//像中继服务器发送请求连接数据
	_, err = relay.Write(Message{Type: SDP, To: target, Data: conv.String(conn.LocalDescription())}.Bytes())
	if err != nil {
		return nil, err
	}

	//处理中继服务器返回的数据
	wait := chans.NewSafe[struct{}]()
	dc.OnOpen(func() { wait.Add(struct{}{}) })
	relay.OnDealMessage(func(c *client.Client, msg ios.Acker) {

		var err error
		defer func() { wait.CloseWithErr(err) }()

		m := Message{}
		err = json.Unmarshal(msg.Bytes(), &m)
		if err != nil {
			return
		}

		switch m.Type {
		case ICE:
			//一些ICE信息,用于建立P2P
			err = conn.AddICECandidate(webrtc.ICECandidateInit{Candidate: m.Data})

		case SDP:
			desc := webrtc.SessionDescription{}
			err = json.Unmarshal([]byte(m.Data), &desc)
			if err != nil {
				return
			}
			_, err = desc.Unmarshal()
			if err != nil {
				return
			}
			//设置远程配置描述
			if desc.Type == webrtc.SDPTypeAnswer {
				err = conn.SetRemoteDescription(desc)
				return
			}
		case Error:
			err = errors.New(m.Data)
		}

	})

	//等待中继服务器响应
	select {
	case <-wait.Done():
		return nil, wait.Err()

	case <-time.After(time.Second * 300):
		return nil, errors.New("建立连接超时")

	case <-wait.Chan:

	}

	p := newClient(target, dc)

	//设置关闭函数
	p.Closer.SetCloseFunc(func(err error) error {
		close(p.ch)
		return p.dc.Close()
	})

	//监听连接状态
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
		case webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateClosed:
			//大概10来秒才能监测到连接断开
			p.CloseWithErr(io.EOF)
		}
	})

	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		select {
		case p.ch <- msg.Data:
		default:
		}
	})

	return p, nil
}

var _ ios.MReadWriteCloser = &Client{}

func newClient(key string, channel *webrtc.DataChannel) *Client {
	return &Client{
		key:    key,
		ch:     make(chan []byte, 100),
		dc:     channel,
		offer:  webrtc.SessionDescription{},
		Closer: safe.NewCloser(),
	}
}

type Client struct {
	key   string
	ch    chan []byte
	dc    *webrtc.DataChannel
	offer webrtc.SessionDescription
	*safe.Closer
}

func (this *Client) ReadMessage() ([]byte, error) {
	if this.Closed() {
		return nil, this.Err()
	}
	bs, ok := <-this.ch
	if !ok {
		if this.Err() != nil {
			return nil, this.Err()
		}
		return nil, io.EOF
	}
	return bs, nil
}

func (this *Client) Write(p []byte) (int, error) {
	if this.Closed() {
		return 0, this.Err()
	}
	err := this.dc.Send(p)
	this.CloseWithErr(err)
	return len(p), err
}
