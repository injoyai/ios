package split

import (
	"bytes"
	"errors"
	"github.com/injoyai/conv"
	"github.com/injoyai/ios"
	"io"
)

// Split 通用分包配置,适用99%的协议
type Split struct {
	Prefix                        //匹配帧头
	Suffix                        //匹配帧尾
	Length                        //匹配长度
	Checker                       //数据校验,crc,sum等
	OnErr   func(err error) error //处理错误信息,可以重置成nil,例如超时
}

func (this *Split) ReadFrom(r io.Reader) (result []byte, err error) {

	defer func() {
		if this.OnErr != nil {
			err = this.OnErr(err)
		}
	}()

loop:
	for {
		result = []byte(nil)

		for {
			b, err := ios.ReadByte(r)
			if err != nil {
				return result, err
			}
			result = append(result, b)

			/*

			 */

			//校验数据是否满足帧头
			preMatch, invalid, err := this.Prefix.Check(result)
			if err != nil {
				return result, err
			}

			if invalid {
				//表示是无效数据,重新开始读取
				continue loop
			}

			if !preMatch {
				//暂时还不满足所有要求,等待读取一字节继续判断
				continue
			}

			/*

			 */

			//校验数据是否满足帧尾
			sufMatch, invalid, err := this.Suffix.Check(result)
			if err != nil {
				return result, err
			}

			if invalid {
				//表示是无效数据,重新开始读取
				continue loop
			}

			if !sufMatch {
				//暂时还不满足所有要求,等待读取一字节继续判断
				continue
			}

			/*

			 */

			//校验数据长度
			lenMatch, invalid, err := this.Length.Check(result)
			if err != nil {
				return result, err
			}

			if invalid {
				//表示是无效数据,重新开始读取
				continue loop
			}

			if !lenMatch {
				//暂时还不满足所有要求,等待读取一字节继续判断
				continue
			}

			/*

			 */

			if this.Checker != nil && !this.Checker.Check(result) {
				//表示是无效数据,重新开始读取
				continue loop
			}

			return result, nil

		}
	}
}

type Prefix []byte

func (this Prefix) Check(bs []byte) (match bool, invalid bool, err error) {
	for i, b := range bs {
		if i < len(this) && this[i] != b {
			return false, true, nil
		}
	}
	match = bytes.HasPrefix(bs, this)
	invalid = !match && len(bs) > len(this)
	return
}

type Suffix []byte

func (this Suffix) Check(bs []byte) (match bool, invalid bool, err error) {
	match = bytes.HasSuffix(bs, this)
	return
}

type Length struct {
	LittleEndian bool //支持大端小端(默认false,大端),暂不支持2143,3412...
	Start, End   uint //长度起始位置,长度结束位置
	Fixed        int  //固定增加长度,例如部分包长度指的后续长度,需补充总长度
}

func (this Length) Check(bs []byte) (match bool, invalid bool, err error) {

	//设置了错误的参数
	if this.Start > this.End {
		return false, false, errors.New("参数长度起始结束设置有误")
	}

	//未设置
	if this.Start == 0 && this.End == 0 {
		return true, false, nil
	}

	//数据还不满足条件
	if len(bs) <= int(this.End) {
		return false, false, nil
	}

	//获取数据总长度
	lenBytes := bs[this.Start : this.End+1]
	if this.LittleEndian {
		lenBytes = Reverse(lenBytes)
	}

	//增加附加长度
	length := conv.Int(lenBytes) + this.Fixed

	match = length == len(bs)
	invalid = length < len(bs)

	return
}

// Reverse 倒序
func Reverse(bs []byte) []byte {
	x := make([]byte, len(bs))
	for i, v := range bs {
		x[len(bs)-i-1] = v
	}
	return x
}
