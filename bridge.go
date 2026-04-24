package ios

import (
	"io"
)

// Bridge 桥接,桥接两个ReadWriteCloser
// 例如,桥接串口(客户端)和网口(tcp客户端),可以实现通过串口上网
func Bridge(r1, r2 io.ReadWriteCloser) error {
	defer func() {
		r1.Close()
		r2.Close()
	}()

	// 创建通道监听错误
	errCh := make(chan error, 2)

	// 从 r1 到 r2
	go func() {
		_, err := io.Copy(r2, r1)
		errCh <- err
	}()

	// 从 r2 到 r1
	go func() {
		_, err := io.Copy(r1, r2)
		errCh <- err
	}()

	// 等待任一方向出错或关闭
	err := <-errCh
	return err
}
