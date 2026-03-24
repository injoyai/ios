package p2p

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/injoyai/base/chans"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
	"github.com/injoyai/logs"
	"github.com/pion/webrtc/v4"
)

func NewListen(key string, relay *client.Client) ios.ListenFunc {
	return func() (ios.Listener, error) {
		return Listen(key, relay)
	}
}

func Listen(key string, relay *client.Client) (ios.Listener, error) {

	l := &listen{
		key:     key,
		clients: maps.NewGeneric[string, *webrtc.PeerConnection](),
		accept:  chans.NewSafe[*Client](10),
		relay:   relay,
	}

	relay.OnDealMessage(func(c *client.Client, msg ios.Acker) {
		logs.PrintErr(func() error {
			m := Message{}
			err := json.Unmarshal(msg.Bytes(), &m)
			if err != nil {
				return err
			}

			conn, err := l.clients.GetOrSetByHandler(m.From, func() (*webrtc.PeerConnection, error) {
				return l.newPeer(m.From)
			})
			if err != nil {
				return err
			}

			switch m.Type {
			case ICE:
				//一些ICE信息,用于建立P2P
				return conn.AddICECandidate(webrtc.ICECandidateInit{Candidate: m.Data})

			case SDP:
				//客户端尝试建立连接请求

				desc := webrtc.SessionDescription{}
				err = json.Unmarshal([]byte(m.Data), &desc)
				if err != nil {
					return err
				}
				if desc.Type == webrtc.SDPTypeOffer {
					if conn.RemoteDescription() == nil {
						err = conn.SetRemoteDescription(desc)
						if err != nil {
							return err
						}
						answer, err := conn.CreateAnswer(nil)
						if err != nil {
							return err
						}
						err = conn.SetLocalDescription(answer)
						if err != nil {
							return err
						}
						//监听候选地址,向中继服务器发送,由中继转发给目标
						conn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
							if candidate == nil {
								return
							}
							relay.Write(Message{Type: ICE, To: m.From, Data: candidate.ToJSON().Candidate}.Bytes())
						})

					}

					_, err = c.Write(Message{Type: SDP, To: m.From, Data: conv.String(conn.LocalDescription())}.Bytes())
					return err

				}

			}
			return nil
		}())
	})
	go relay.Run(context.Background())

	return l, nil
}

var _ ios.Listener = &listen{}

type listen struct {
	key     string
	clients *maps.Generic[string, *webrtc.PeerConnection]
	accept  *chans.Safe[*Client]
	relay   *client.Client
}

func (this *listen) newPeer(key string) (*webrtc.PeerConnection, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}},
	})
	if err != nil {
		return nil, err
	}
	conn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		switch state {
		case webrtc.PeerConnectionStateConnected:
		case webrtc.PeerConnectionStateFailed,
			webrtc.PeerConnectionStateDisconnected,
			webrtc.PeerConnectionStateClosed:
			this.clients.Del(key)
			conn.Close()
		}
	})
	conn.OnDataChannel(func(channel *webrtc.DataChannel) {
		p := newClient(key, channel)
		channel.OnOpen(func() { this.accept.Must(p) })
		channel.OnMessage(func(msg webrtc.DataChannelMessage) {
			select {
			case p.ch <- msg.Data:
			default:
			}
		})
	})
	return conn, nil
}

func (this *listen) Close() error {
	this.relay.Close()
	this.accept.Close()
	return nil
}

func (this *listen) Accept() (ios.ReadWriteCloser, string, error) {
	r, ok := <-this.accept.Chan
	if !ok {
		return nil, "", errors.New("listener closed")
	}
	return r, r.key, nil
}

func (this *listen) Addr() string {
	return this.key
}
