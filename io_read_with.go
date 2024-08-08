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
