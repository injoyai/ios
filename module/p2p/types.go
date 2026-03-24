package p2p

import "encoding/json"

const (
	ICE   = "ice"
	SDP   = "sdp"
	Error = "err" //ws响应错误
)

type Message struct {
	Type string `json:"type"`
	From string `json:"from,omitempty"` //来源,由中继服务器填入
	To   string `json:"to,omitempty"`
	Data string `json:"data"`
}

func (this Message) Bytes() []byte {
	bs, _ := json.Marshal(this)
	return bs
}
