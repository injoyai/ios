package ios

import "io"

func NewReadWithBuffer(buf []byte) func(r io.Reader) ([]byte, error) {
	return func(r io.Reader) ([]byte, error) {
		n, err := r.Read(buf)
		if err != nil {
			return nil, err
		}
		return buf[:n], nil
	}
}

type Bytes []byte

func (this Bytes) ReadFrom(r io.Reader) ([]byte, error) {
	n, err := r.Read(this)
	if err != nil {
		return nil, err
	}
	return this[:n], nil
}
