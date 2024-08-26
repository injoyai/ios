package ios

import (
	"github.com/injoyai/base/chans"
	"io"
)

// SplitBytesByLength
// 按最大长度分割字节 todo 这个不应该出现在这里
func SplitBytesByLength(p []byte, max int) [][]byte {
	if max == 0 {
		return [][]byte{}
	}
	list := [][]byte(nil)
	for len(p) > max {
		list = append(list, p[:max])
		p = p[max:]
	}
	list = append(list, p)
	return list
}

// Pipe 一个双向通道
func Pipe() (io.ReadWriteCloser, io.ReadWriteCloser) {
	r1 := chans.NewIO(0)
	r2 := chans.NewIO(0)
	type T struct {
		io.Reader
		io.Writer
		io.Closer
	}
	i1 := T{Reader: r1, Writer: r2, Closer: MultiCloser(r1, r2)}
	i2 := T{Reader: r2, Writer: r1, Closer: MultiCloser(r2, r1)}
	return i1, i2

}
