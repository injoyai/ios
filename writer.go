package ios

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"sync"

	"github.com/injoyai/conv"
)

var _ MoreWriter = (*MoreWrite)(nil)

func NewMoreWrite(w io.Writer, onWrite ...func(p []byte, write func(p []byte) error) error) *MoreWrite {
	return &MoreWrite{
		Writer:  w,
		onWrite: conv.Default(nil, onWrite...),
	}
}

type MoreWrite struct {
	io.Writer
	onWrite func(p []byte, write func(p []byte) error) error
}

func (this *MoreWrite) Write(p []byte) (n int, err error) {
	if this.onWrite != nil {
		er := this.onWrite(p, func(bs []byte) error {
			n, err = this.Writer.Write(bs)
			return err
		})
		return n, er
	}
	return this.Writer.Write(p)
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
	s = strings.ReplaceAll(s, " ", "")
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

/*



 */

func NewPlanWrite(w io.Writer, onWrite func(*Plan)) *PlanWrite {
	return &PlanWrite{
		Writer:  w,
		onWrite: onWrite,
	}
}

type PlanWrite struct {
	io.Writer
	plan    *Plan
	once    sync.Once
	onWrite func(*Plan)
}

func (this *PlanWrite) Write(p []byte) (n int, err error) {
	this.once.Do(func() {
		this.plan = &Plan{}
	})
	this.plan.Index++
	this.plan.LastBytes = p
	if this.onWrite != nil {
		this.onWrite(this.plan)
	}
	return this.Writer.Write(p)
}

type Plan struct {
	Index     int64  //写入次数
	TotalSize int64  //已写入的字节大小
	LastBytes []byte //最后的数据内容
}

func (this *Plan) SetTotal(total int64) {
	this.TotalSize = total
}

func (this *Plan) Rate(total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(this.TotalSize) / float64(total)
}
