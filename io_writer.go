package ios

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/injoyai/conv"
	"io"
)

var _ MoreWriter = (*MoreWrite)(nil)

type WriteOption func(p []byte) ([]byte, error)
type WriteResult func(err error)

func NewMoreWriter(w io.Writer, op ...WriteOption) MoreWriter {
	return &MoreWrite{
		Writer: w,
		Option: op,
	}
}

type MoreWrite struct {
	io.Writer
	Option []WriteOption
	Result []WriteResult
}

func (this *MoreWrite) Write(p []byte) (n int, err error) {
	for _, f := range this.Option {
		if f != nil {
			p, err = f(p)
			if err != nil {
				return 0, err
			}
		}
	}
	n, err = this.Writer.Write(p)
	for _, f := range this.Result {
		if f != nil {
			f(err)
		}
	}
	return
}

func (this *MoreWrite) WriteString(s string) (n int, err error) {
	return this.Write([]byte(s))
}

func (this *MoreWrite) WriteByte(c byte) error {
	_, err := this.Write([]byte{c})
	return err
}

func (this *MoreWrite) WriteBase64(s string) error {
	bs, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	_, err = this.Write(bs)
	return err
}

func (this *MoreWrite) WriteHEX(s string) error {
	bs, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	_, err = this.Write(bs)
	return err
}

func (this *MoreWrite) WriteJson(a any) error {
	bs, err := json.Marshal(a)
	if err != nil {
		return err
	}
	_, err = this.Write(bs)
	return err
}

func (this *MoreWrite) WriteAny(a any) error {
	bs := conv.Bytes(a)
	_, err := this.Write(bs)
	return err
}

func (this *MoreWrite) WriteChan(c chan any) error {
	for {
		v, ok := <-c
		if !ok {
			return nil
		}
		_, err := this.Write(conv.Bytes(v))
		if err != nil {
			return err
		}
	}
}

type PlanWrite struct {
	io.Writer
	*Plan
	OnWrite func(*Plan)
}

func (this *PlanWrite) Write(p []byte) (n int, err error) {
	if this.Plan == nil {
		this.Plan = &Plan{
			Index:   0,
			Total:   0,
			Current: 0,
			Bytes:   nil,
		}
	}
	this.Plan.Index++
	this.Plan.Current += int64(len(p))
	this.Plan.Bytes = p
	if this.OnWrite != nil {
		this.OnWrite(this.Plan)
	}
	if this.Plan.Err != nil {
		return 0, this.Plan.Err
	}
	return this.Writer.Write(this.Plan.Bytes)
}

type Plan struct {
	Index   int64
	Total   int64
	Current int64
	Bytes   []byte //数据内容
	Err     error  //错误信息
}

func (this *Plan) SetTotal(total int64) {
	this.Total = total
}

func (this *Plan) Rate() float64 {
	if this.Total == 0 {
		return 0
	}
	return float64(this.Current) / float64(this.Total)
}
