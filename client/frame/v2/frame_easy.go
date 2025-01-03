package frame

import (
	"errors"
	"github.com/injoyai/base/bytes"
	"github.com/injoyai/conv"
	"io"
	"math"
)

/*
简易封装包

包构成(大端):
.===========================================.
|构成	|字节	|类型	|说明				|
|-------------------------------------------|
|帧头 	|2字节 	|Byte	|固定0xFEFE			|
|-------------------------------------------|
|帧长  	|4字节	|HEX	|总字节长度			|
|-------------------------------------------|
|内容	|可变	|Byte	|数据内容			|
|-------------------------------------------|
|校验和	|2字节	|Byte	|校验和 				|
^===========================================^


*/

var (
	Prefix = []byte{0xFE, 0xFE}
)

type Frame struct {
	Data []byte
}

func (this *Frame) Bytes() []byte {
	bs := make([]byte, len(this.Data)+8)
	copy(bs, Prefix)
	copy(bs[2:], conv.Bytes(uint32(len(this.Data)+2)))
	copy(bs[6:], this.Data)
	sum := uint16(0)
	for _, v := range this.Data {
		sum += uint16(v)
	}
	copy(bs[len(bs)-2:], conv.Bytes(sum))
	return bs
}

func ReadFrom(r io.Reader) ([]byte, error) {
	for {

		//读取帧头
		bs := make([]byte, 2)
		n, err := r.Read(bs)
		if err != nil {
			return nil, err
		}
		if n == 2 && bytes.Equal(bs, Prefix) {

			//读取长度
			bs = make([]byte, 4)
			n, err = r.Read(bs)
			if err != nil {
				return nil, err
			}

			if n == 4 {

				length := conv.Int(bs)
				if length >= 2 {

					//读取数据域
					result := make([]byte, length)
					_, err = io.ReadAtLeast(r, result, length)
					if err != nil {
						return nil, err
					}

					//判断校验和
					sum := uint16(0)
					for _, v := range result[:length-2] {
						sum += uint16(v)
					}
					if conv.Uint16(result[length-2:]) == sum {
						return result[:length-2], nil
					}

				}

			}

		}

	}
}

func WriteWith(bs []byte) ([]byte, error) {
	if len(bs) > math.MaxUint32 {
		return nil, errors.New("数据长度太长")
	}
	f := Frame{Data: bs}
	return f.Bytes(), nil
}

var Default = _default{}

type _default struct{}

func (_default) ReadFrom(r io.Reader) ([]byte, error) {
	return ReadFrom(r)
}

func (_default) WriteWith(bs []byte) ([]byte, error) {
	return WriteWith(bs)
}
