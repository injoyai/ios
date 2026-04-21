package ios

import (
	"io"
)

// Bridge 桥接,桥接两个ReadWriter
// 例如,桥接串口(客户端)和网口(tcp客户端),可以实现通过串口上网
func Bridge(i1, i2 io.ReadWriter) error {
	return Swap(i1, i2)
}

func Swap(r1, r2 io.ReadWriter) error {
	go io.Copy(r1, r2)
	_, err := io.Copy(r2, r1)
	return err
}
