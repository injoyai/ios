package ios

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/injoyai/conv"
	"io"
)

type Checker interface {
	Check([]byte) (bool, bool)
}

var _ MReader = (*Split)(nil)

// Split 数据分包
type Split struct {
	io.Reader
	Check []Checker
	buf   *bufio.Reader
}

func (this *Split) ReadMessage() ([]byte, error) {
	if this.Reader == nil {
		return nil, errors.New("reader is nil")
	}
	if this.buf == nil {
		this.buf = bufio.NewReader(this.Reader)
	}

loop:
	for {
		result := []byte(nil)
		for {
			b, err := this.buf.ReadByte()
			if err != nil {
				return nil, err
			}
			result = append(result, b)
			allMeet := true
			for _, v := range this.Check {
				meet, invalid := v.Check(result)
				if !invalid {
					//表示是无效数据,重新开始读取
					goto loop
				}
				if !meet {
					//暂时还不满足所有要求,等待读取一字节继续判断
					allMeet = false
					break
				}
			}
			if allMeet {
				return result, nil
			}
		}
	}
}

type SplitStartEnd struct {
	Start, End []byte //帧头,帧尾
}

func (this *SplitStartEnd) Check(bs []byte) (bool, bool) {
	if this == nil {
		return true, true
	}

	if len(bs) <= len(this.Start) {
		//如果小于帧头,判断是否符合包含,是否是有效数据
		return false, bytes.HasPrefix(this.Start, bs)
	}

	//帧头不满足说明是无效数据,帧尾不满足说明暂时还不满足
	return bytes.HasSuffix(bs, this.End), bytes.HasPrefix(bs, this.Start)
}

type SplitLength struct {
	LittleEndian     bool //支持大端小端(默认false,大端),暂不支持2143,3412...
	LenStart, LenEnd uint //长度起始位置,长度结束位置
	LenFixed         int  //固定增加长度,有些不计入长度字段
}

func (this *SplitLength) Check(bs []byte) (bool, bool) {
	if this == nil {
		return true, true
	}

	//设置了错误的参数
	if this.LenStart >= this.LenEnd {
		return true, true
	}

	//数据还不满足条件
	if len(bs) <= int(this.LenEnd) {
		return false, true
	}

	//获取长度字节
	lenBytes := bs[this.LenStart : this.LenEnd+1]
	if this.LittleEndian {
		lenBytes = Reverse(lenBytes)
	}
	length := conv.Int(lenBytes) + this.LenFixed

	//返回结果
	return length == len(bs), len(bs) <= length
}

type SplitTotal struct {
	Least uint //至少
	Most  uint //至多
}

func (this *SplitTotal) Check(bs []byte) (bool, bool) {
	if this == nil {
		return true, true
	}
	return len(bs) >= int(this.Least),
		this.Most == 0 || len(bs) <= int(this.Most)
}

func Reverse(bs []byte) []byte {
	x := make([]byte, len(bs))
	for i, v := range bs {
		x[len(bs)-i-1] = v
	}
	return x
}
