package ios

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

func NewFRead(buf []byte) FReadFunc {
	if buf == nil {
		buf = make([]byte, DefaultBufferSize)
	}
	return Buffer(buf).ReadFrom
}

// NewFRead2 读取函数
func NewFRead2(buf []byte) FRead2Func {
	return NewFRead2WithHandler(NewFRead(buf))
}

// NewFReadLeast 新建读取函数,至少读取设置的字节
func NewFReadLeast(least int) FReadFunc {
	buf := make([]byte, least)
	return func(r io.Reader) ([]byte, error) {
		_, err := io.ReadAtLeast(r, buf, least)
		return buf, err
	}
}

// NewFReadKB 新建读取函数,按KB读取
func NewFReadKB(n int) FReadFunc {
	return NewFRead(make([]byte, 1024*n))
}

// NewFRead4KB 新建读取函数,按4KB读取
func NewFRead4KB() FReadFunc {
	return NewFRead(make([]byte, 1024*4))
}

// NewFRead2WithHandler 读取函数
func NewFRead2WithHandler(f FReadFunc) FRead2Func {
	if f == nil {
		buf := Buffer(make([]byte, DefaultBufferSize))
		f = buf.ReadFrom
	}
	return func(r Reader) ([]byte, error) {
		switch v := r.(type) {
		case MReader:
			return v.ReadMessage()

		case AReader:
			a, err := v.ReadAck()
			if err != nil {
				return nil, err
			}
			defer a.Ack()
			return a.Payload(), nil

		case io.Reader:
			return f(v)

		default:
			return nil, fmt.Errorf("未知类型: %T, 未实现[Reader|MReader|AReader]", r)

		}

	}
}

// ReadByte 读取一字节
func ReadByte(r io.Reader) (byte, error) {
	switch v := r.(type) {
	case io.ByteReader:
		return v.ReadByte()
	default:
		b := make([]byte, 1)
		_, err := io.ReadAtLeast(r, b, 1)
		return b[0], err
	}
}

// ReadPrefix 读取Reader符合的头部,返回成功(nil),或者错误
func ReadPrefix(r io.Reader, prefix []byte) ([]byte, error) {
	cache := []byte(nil)
	b1 := make([]byte, 1)
	for index := 0; index < len(prefix); {
		switch v := r.(type) {
		case io.ByteReader:
			b, err := v.ReadByte()
			if err != nil {
				return cache, err
			}
			cache = append(cache, b)
		default:
			_, err := io.ReadAtLeast(r, b1, 1)
			if err != nil {
				return cache, err
			}
			cache = append(cache, b1[0])
		}
		if cache[len(cache)-1] == prefix[index] {
			index++
		} else {
			for len(cache) > 0 {
				//only one error in this ReadPrefix ,it is EOF,and not important
				cache2, _ := ReadPrefix(bytes.NewReader(cache[1:]), prefix)
				if len(cache2) > 0 {
					cache = cache2
					break
				}
				cache = cache[1:]
			}
			index = len(cache)
		}
	}
	return cache, nil
}

/*



 */

var (
	bufferPool = sync.Pool{New: func() any {
		return Buffer(make([]byte, DefaultBufferSize))
	}}
)

type Buffer []byte

func (this Buffer) ReadFrom(r io.Reader) ([]byte, error) {
	n, err := r.Read(this)
	if err != nil {
		return nil, err
	}
	return this[:n], nil
}
