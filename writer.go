package ios

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"

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

func NewPlanWrite(w io.Writer, onWrite func(Plan)) *PlanWrite {
	return &PlanWrite{
		Writer: w,
		plan: &Plan{
			Start: time.Now(),
		},
		onWrite: onWrite,
	}
}

type PlanWrite struct {
	io.Writer
	plan    *Plan
	onWrite func(Plan)
	mu      sync.Mutex
}

func (this *PlanWrite) Write(p []byte) (n int, err error) {
	n, err = this.Writer.Write(p)

	if n > 0 {
		this.mu.Lock()
		this.plan.Count++
		this.plan.Current += int64(n)
		this.plan.Last = p
		plan := *this.plan
		this.mu.Unlock()

		if this.onWrite != nil {
			this.onWrite(plan)
		}
	}

	return
}

type Plan struct {
	Count   int64     //写入次数
	Current int64     //已写入的字节大小
	Last    []byte    //最后的数据内容
	Start   time.Time //开始时间
}

func (this Plan) Percent(total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(this.Current) / float64(total) * 100
}

func (this Plan) AvgRate() float64 {
	sec := time.Now().Sub(this.Start).Seconds()
	if sec <= 0 {
		return 0
	}
	return float64(this.Current) / sec
}

func (this Plan) Remain(total int64) int64 {
	r := total - this.Current
	if r < 0 {
		r = 0
	}
	return r
}

func (this Plan) ETA(total int64) time.Duration {
	remain := this.Remain(total)
	avgRate := this.AvgRate()
	if avgRate <= 0 {
		return -1
	}
	return time.Duration(float64(remain) * float64(time.Second) / avgRate)
}
