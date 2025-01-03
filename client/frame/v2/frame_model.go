package frame

import (
	"github.com/injoyai/ios"
	"github.com/injoyai/ios/client"
)

/*
可选对象结构


|状态码	|3bit	|Byte	|0~8	推荐:		|
|		|		|		|0:请求 2:成功 5:失败	|
|消息ID	|5bit	|Byte	|0~32				|
|-------------------------------------------|
|帧类型	|1字节	|BIN	|自定义				|
|-------------------------------------------|
|内容	|可变	|Byte	|数据内容			|
|-------------------------------------------|

*/

var (
	Fail uint8 = 5
	Succ uint8 = 2
)

type Model struct {
	Code  uint8  //3位
	MsgID uint8  //5位
	Type  uint8  //1字节
	Data  []byte //数据
}

func (this *Model) Bytes() []byte {
	bs := make([]byte, len(this.Data)+2)
	bs[0] = this.Code<<5 + this.MsgID%32
	bs[1] = this.Type
	copy(bs[2:], this.Data)
	return bs
}

func (this *Model) IsRequest() bool {
	return this.Code == 0
}

func (this *Model) IsResponse() bool {
	return this.Code > 0
}

// IsFail 是否是失败,可选,也可自定义
func (this *Model) IsFail() bool {
	return this.Code != 0 && this.Code != 2
}

// IsSucc 是否成功,可选,也可自定义
func (this *Model) IsSucc() bool {
	return this.Code == 2
}

// RespSucc 响应成功
func (this *Model) RespSucc(data []byte) *Model {
	return this.Resp(Succ, data)
}

// RespErr 响应错误
func (this *Model) RespErr(err error) *Model {
	if err != nil {
		return this.Resp(Fail, []byte(err.Error()))
	}
	return this.Resp(Succ, []byte{})
}

// Resp 生成响应
func (this *Model) Resp(code uint8, data []byte) *Model {
	this.Code = code<<5 + this.MsgID%32
	this.Data = data
	return this
}

func OnMessage(f func(m *Model)) func(c *client.Client, msg ios.Acker) {
	return func(c *client.Client, msg ios.Acker) {
		m := &Model{}
		bs := msg.Payload()
		if len(bs) > 2 {
			m.Code = bs[0] >> 5
			m.MsgID = bs[0] & 0x1F
			m.Type = bs[1]
			m.Data = bs[2:]
		}
		f(m)
	}
}
