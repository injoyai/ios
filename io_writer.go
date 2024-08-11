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

func NewMoreWriter(w io.Writer, op ...WriteOption) MoreWriter {
	return &MoreWrite{
		Writer: w,
		Option: op,
	}
}

type MoreWrite struct {
	io.Writer
	Option []WriteOption
}

func (this *MoreWrite) Write(p []byte) (n int, err error) {
	for _, f := range this.Option {
		p, err = f(p)
		if err != nil {
			return 0, err
		}
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
