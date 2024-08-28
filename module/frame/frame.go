package frame

import (
	"compress/gzip"
	"fmt"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/base/maps"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios/module/client"
	"github.com/injoyai/logs"
	"hash/crc32"
	"io"
	"time"
)

/*

通用封装包

包构成(大端):
.===================================.
|构成	|字节	|类型	|说明		|
|-----------------------------------|
|帧头 	|2字节 	|Byte	|固定0x8888	|
|-----------------------------------|
|帧长  	|4字节	|HEX	|总字节长度	|
|-----------------------------------|
|帧类型	|2字节	|BIN	|详见帧类型	|
|-----------------------------------|
|消息号	|1字节	|Byte	|消息id		|
|-----------------------------------|
|内容	|可变	|Byte	|数据内容	|
|-----------------------------------|
|校验和	|4字节	|Byte	|crc IEEE	|
|-----------------------------------|
|帧尾 	|2字节	|Byte	|固定0x8989	|
^===================================^

包类型:
.=======================================================================================================.
|bit15				|bit14		|bit13~11				|bit10	|bit9	|bit8							|
|-------------------------------------------------------------------------------------------------------|
|数据方向0请求,1响应	|预留		|压缩方式,0无,1gzip		|预留											|
^=======================================================================================================^
|bit7   								|功能码															|
|-------------------------------------------------------------------------------------------------------|
|数据的读写0读/订阅/接收,1写/发布/发送		|																|
^=======================================================================================================^
*/

var (
	start = []byte{0x88, 0x88} //帧头
	end   = []byte{0x89, 0x89} //帧尾
)

const (
	baseLength        = 15
	ControlCall uint8 = 0x00
	ControlBack uint8 = 0x80
	ControlGzip uint8 = 0x10
)
const (
	// 内置功能码,待定

	FunctionCustom      uint8 = 0x0 //自定义
	FunctionPing        uint8 = 0x1 //测试连接,无数据
	FunctionTime        uint8 = 0x2 //时间(时间戳),同步时间
	FunctionSubscribe   uint8 = 0x3 //订阅
	FunctionIMEI        uint8 = 0x4 //imei
	FunctionICCID       uint8 = 0x5 //iccid
	FunctionIMSI        uint8 = 0x6 //imsi
	FunctionReload      uint8 = 0x7 //重新加载
	FunctionReboot      uint8 = 0x8 //重启设备
	FunctionLinkAddress uint8 = 0x9 //新连接地址
	FunctionSlave       uint8 = 0xA //设置从站地址,主节点开始,后续一次增加长度,例如主节点是 "1" 分配给子节点就是 "1.1","1.2","1.13"
)

func NewPing() []byte {
	//00000001
	return (&Frame{Control: ControlCall, Function: FunctionPing}).Bytes()
}

func NewPong() []byte {
	//10000001
	return (&Frame{Control: ControlBack, Function: FunctionPing}).Bytes()
}

func New(msgID uint8, data []byte) *Frame {
	return NewCall(msgID, data)
}

func NewCall(msgID uint8, data []byte) *Frame {
	return &Frame{
		Control:  ControlCall,
		Function: FunctionCustom,
		MsgID:    msgID,
		Data:     data,
	}
}

func NewBack(msgID uint8, data []byte) *Frame {
	return NewCall(msgID, data).Resp(data)
}

type Frame struct {
	Control  uint8  //控制码
	Function uint8  //功能码
	MsgID    uint8  //消息id
	Data     []byte //数据内容
}

// SetCompress 设置压缩方式,1是gzip,其他不压缩
func (this *Frame) SetCompress(n uint8) *Frame {
	this.Control &= 0xCF
	this.Control |= n
	return this
}

func (this *Frame) String() string {
	return this.Bytes().HEX()
}

func (this *Frame) encodeData() []byte {
	data := this.Data
	switch this.Control & 0x30 {
	case ControlGzip:
		// Gzip 压缩字节
		buf := bytes.NewBuffer(nil)
		gzipWriter := gzip.NewWriter(buf)
		gzipWriter.Write(data)
		gzipWriter.Close()
		data = buf.Bytes()
	default:
	}
	return data
}

func (this *Frame) decodeData() error {
	switch this.Control & 0x30 {
	case ControlGzip:
		// Gzip 解压字节
		reader := bytes.NewReader(this.Data)
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return err
		}
		defer gzipReader.Close()
		buf := new(bytes.Buffer)
		_, err = buf.ReadFrom(gzipReader)
		if err != nil {
			logs.Err(err)
			return err
		}
		this.Data = buf.Bytes()
	default:
	}
	return nil
}

func (this *Frame) Bytes() bytes.Entity {
	data := []byte(nil)
	data = append(data, start...)
	dataBytes := this.encodeData()
	length := uint32(len(dataBytes) + baseLength)
	data = append(data, conv.Bytes(length)...)
	data = append(data, this.Control, this.Function)
	data = append(data, this.MsgID)
	data = append(data, dataBytes...)
	data = append(data, conv.Bytes(crc32.ChecksumIEEE(data))...)
	data = append(data, end...)
	return data
}

// Resp 生成响应包
func (this *Frame) Resp(bs []byte) *Frame {
	this.Control |= ControlBack
	this.Data = bs
	return this
}

// IsCall 是否请求数据
func (this *Frame) IsCall() bool {
	return this.Control&0x80 == 0
}

// IsBack 是否是响应数据
func (this *Frame) IsBack() bool {
	return this.Control&0x80 == 0x80
}

// IsPing 是否是ping,需要响应pong
func (this *Frame) IsPing() bool {
	return this.IsCall() && this.Function&0x7F == FunctionPing
}

// IsPong 是否是pong,不需要处理
func (this *Frame) IsPong() bool {
	return this.IsBack() && this.Function&0x7F == FunctionPing
}

func (this *Frame) IsRead() bool {
	return this.Function&0x80 == 0
}

func (this *Frame) IsWrite() bool {
	return this.Function&0x80 == 0x80
}

func (this *Frame) GetFunction() uint8 {
	return this.Function & 0x7F
}

func (this *Frame) dealWithTag(w io.Writer, tag *maps.Safe, key string) error {
	if this.IsCall() {
		if this.IsRead() {
			if _, err := w.Write(this.Resp(conv.Bytes(tag.GetString(key))).Bytes()); err != nil {
				return err
			}
		}
		if this.IsWrite() {
			tag.Set(key, string(this.Data))
			if _, err := w.Write(this.Resp(nil).Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

// Decode 按自定义的包解析
func Decode(bs []byte) (*Frame, error) {

	//校验基础数据长度
	if len(bs) <= baseLength {
		return nil, fmt.Errorf("数据长度小于(%d)", baseLength)
	}

	//校验帧头
	if bs[0] != start[0] && bs[1] != start[1] {
		return nil, fmt.Errorf("帧头错误,预期(%x),得到(%x)", start, bs[:2])
	}

	//获取总数据长度
	length := conv.Int(bs[2:6])

	//校验总长度
	if len(bs) != length {
		return nil, fmt.Errorf("数据总长度错误,预期(%d),得到(%d)", length, len(bs))
	}

	//校验crc32
	if crc1, crc2 := crc32.ChecksumIEEE(bs[:length-6]), conv.Uint32(bs[length-6:length-2]); crc1 != crc2 {
		return nil, fmt.Errorf("数据CRC校验错误,预期(%x),得到(%x)", crc1, crc2)
	}

	//校验帧尾
	if bs[length-2] != end[0] && bs[length-1] != bs[1] {
		return nil, fmt.Errorf("帧尾错误,预期(%x),得到(%x)", end, bs[length-2:])
	}

	p := &Frame{
		Control:  bs[6],
		Function: bs[7],
		MsgID:    bs[8],
		Data:     bs[9 : length-6],
	}

	return p, p.decodeData()

}

func WithDeal(w io.Writer, tag *maps.Safe, bs client.Message) (*Frame, error) {

	p, err := Decode(bs)
	if err != nil {
		return nil, err
	}
	switch p.GetFunction() {
	case FunctionPing:
		if p.IsCall() {
			if _, err := w.Write(p.Resp(nil).Bytes()); err != nil {
				return nil, err
			}
		}
	case FunctionTime:
		if p.IsCall() && p.IsRead() {
			now := conv.Bytes(time.Now().UnixMilli())
			if _, err := w.Write(p.Resp(now).Bytes()); err != nil {
				return nil, err
			}
		}

	case FunctionIMEI:
		if err := p.dealWithTag(w, tag, "imei"); err != nil {
			return nil, err
		}

	case FunctionICCID:
		if err := p.dealWithTag(w, tag, "iccid"); err != nil {
			return nil, err
		}

	case FunctionIMSI:
		if err := p.dealWithTag(w, tag, "imsi"); err != nil {
			return nil, err
		}

	case FunctionSlave:
		if err := p.dealWithTag(w, tag, "salve"); err != nil {
			return nil, err
		}

	}
	return p, nil
}

func WriteMessage(req []byte) ([]byte, error) {
	return New(0, req).Bytes(), nil
}

func ReadFrom(r io.Reader) (result []byte, err error) {
	var bs []byte
	for {

		bs = make([]byte, 2)
		n, err := r.Read(bs)
		if err != nil {
			return result, err
		}

		if n == 2 && bs[0] == start[0] && bs[1] == start[1] {
			//帧头
			result = append(result, bs...)

			bs = make([]byte, 4)
			n, err = r.Read(bs)
			if err != nil {
				return result, err
			}
			if n == 4 {
				//长度
				length := conv.Int(bs)

				if length > baseLength {
					result = append(result, bs...)
					length -= 6

					bs = make([]byte, length)
					_, err = io.ReadAtLeast(r, bs, length)
					result = append(result, bs...)

					//return result, nil

					p, err := Decode(result)
					if err != nil {
						return nil, err
						result = result[:0]
						continue
					}
					return p.Data, nil
				}
			}
		}
	}
}

var Entity = &_frame{}

type _frame struct{}

func (this *_frame) WriteMessage(p []byte) ([]byte, error) {
	return WriteMessage(p)
}

func (this *_frame) ReadFrom(r io.Reader) ([]byte, error) {
	return ReadFrom(r)
}
