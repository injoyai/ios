package ios

import (
	"bufio"
	"bytes"
	"io"
)

func NewFRead(buf []byte) FReadFunc {
	if buf == nil {
		buf = make([]byte, DefaultBufferSize)
	}
	return Bytes(buf).ReadBuffer
}

// NewFReadLeast 新建读取函数,至少读取设置的字节
func NewFReadLeast(least int) FReadFunc {
	buf := make([]byte, least)
	return func(r *bufio.Reader) ([]byte, error) {
		_, err := io.ReadAtLeast(r, buf, least)
		return buf, err
	}
}

// NewFReadB 新建读取函数,按B读取
func NewFReadB(n int) FReadFunc {
	return NewFRead(make([]byte, n))
}

// NewFReadKB 新建读取函数,按KB读取
func NewFReadKB(n int) FReadFunc {
	return NewFRead(make([]byte, 1024*n))
}

// NewFRead4KB 新建读取函数,按4KB读取
func NewFRead4KB() FReadFunc {
	return NewFRead(make([]byte, 1024*4))
}

/*



 */

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
