package ios

import (
	"bufio"
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
	if len(prefix) == 0 {
		return nil, nil
	}
	cache := make([]byte, 0, len(prefix))
	failure := buildPrefixFailure(prefix)
	matched := 0
	b1 := make([]byte, 1)
	for {
		var b byte
		switch v := r.(type) {
		case io.ByteReader:
			val, err := v.ReadByte()
			if err != nil {
				return cache, err
			}
			b = val
		default:
			_, err := io.ReadAtLeast(r, b1, 1)
			if err != nil {
				return cache, err
			}
			b = b1[0]
		}
		cache = append(cache, b)
		for matched > 0 && b != prefix[matched] {
			matched = failure[matched-1]
		}
		if b == prefix[matched] {
			matched++
			if matched == len(prefix) {
				return cache[len(cache)-len(prefix):], nil
			}
		}
	}
}

func buildPrefixFailure(prefix []byte) []int {
	failure := make([]int, len(prefix))
	for i, matched := 1, 0; i < len(prefix); i++ {
		for matched > 0 && prefix[i] != prefix[matched] {
			matched = failure[matched-1]
		}
		if prefix[i] == prefix[matched] {
			matched++
			failure[i] = matched
		}
	}
	return failure
}
