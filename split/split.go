package split

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"regexp"
)

type Checker interface {
	Check([]byte) (match bool, invalid bool, err error)
}

// Split 通用分包配置,适用99%的协议,性能一般O(n²)
type Split struct {
	Checker []Checker             //数据校验,crc,sum等
	OnErr   func(err error) error //处理错误信息,可以重置成nil,例如超时
}

func (s *Split) ReadFrom(r *bufio.Reader) (result []byte, err error) {
	defer func() {
		if s.OnErr != nil && err != nil {
			err = s.OnErr(err)
		}
	}()

clear:
	for {
		// 阻塞等待至少 1 个字节
		if _, err := r.Peek(1); err != nil {
			if err == io.EOF {
				return nil, nil
			}
			return nil, err
		}

	move:
		for i := 0; ; i++ {
			data, err := r.Peek(i + 1)
			if err != nil {
				// 数据不足，等待更多
				if errors.Is(err, bufio.ErrBufferFull) || err == io.EOF {
					return nil, nil
				}
				return nil, err
			}

			for _, c := range s.Checker {
				if c == nil {
					continue
				}
				match, invalid, err := c.Check(data)
				if err != nil {
					return nil, err
				}
				if invalid {
					// 前 i+1 字节无效，丢弃
					r.Discard(i + 1)
					continue clear
				}
				if !match {
					// 数据还不够，继续增加 i
					continue move
				}
			}

			// 找到完整包，直接返回 bufio.Reader 内部 slice
			r.Discard(i + 1)
			return data, nil
		}
	}
}

/*



 */

type Prefixes []Prefix

func (this Prefixes) Check(bs []byte) (match bool, invalid bool, err error) {
	invalid = len(this) > 0
	_invalid := false
	for _, prefix := range this {
		match, _invalid, err = prefix.Check(bs)
		if err != nil {
			return
		}
		if !_invalid {
			invalid = false
		}
		if match {
			return
		}
	}
	return
}

/*



 */

type Prefix []byte

func (this Prefix) Check(bs []byte) (bool, bool, error) {
	if len(bs) >= len(this) && !bytes.HasPrefix(bs, this) {
		return false, true, nil
	}
	if len(bs) < len(this) {
		return false, false, nil
	}
	return true, false, nil
}

/*



 */

type Suffix []byte

func (this Suffix) Check(bs []byte) (match bool, invalid bool, err error) {
	match = bytes.HasSuffix(bs, this)
	return
}

/*



 */

type Regular struct {
	*regexp.Regexp
}

func (this Regular) Check(bs []byte) (bool, bool, error) {
	if this.Regexp == nil {
		return true, false, nil
	}
	return this.Regexp.Match(bs), false, nil
}

/*



 */

type Length struct {
	LittleEndian bool //支持大端小端(默认false,大端)
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
	length := 0
	if this.LittleEndian {
		for i := this.End; i >= this.Start; i-- {
			length = (length << 8) | int(bs[i])
		}
	} else {
		for i := this.Start; i <= this.End; i++ {
			length = (length << 8) | int(bs[i])
		}
	}

	//增加附加长度
	length += this.Fixed

	match = length == len(bs)
	invalid = length < len(bs)

	return
}

/*



 */

// CRC16Modbus 校验器
type CRC16Modbus struct{}

func (c CRC16Modbus) Check(bs []byte) (bool, bool, error) {
	if len(bs) < 3 {
		return false, false, nil
	}
	payload := bs[:len(bs)-2]
	crc := crc16Modbus(payload)
	crcLow := byte(crc & 0xFF)
	crcHigh := byte(crc >> 8)
	return bs[len(bs)-2] == crcLow && bs[len(bs)-1] == crcHigh, false, nil
}

// CRC16-Modbus 算法
func crc16Modbus(data []byte) uint16 {
	var crc uint16 = 0xFFFF
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc >>= 1
				crc ^= 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

/*



 */

type SumLast struct{}

func (s SumLast) Check(bs []byte) (bool, bool, error) {
	if len(bs) < 2 {
		return false, false, nil
	}
	sum := byte(0)
	for _, b := range bs[:len(bs)-1] {
		sum += b
	}
	return sum == bs[len(bs)-1], false, nil
}
