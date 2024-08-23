package common

import (
	"github.com/injoyai/ios"
	"io"
	"strings"
)

// DealErr 错误处理,常见整理成中文
func DealErr(err error) error {
	if err != nil {
		s := err.Error()
		switch {
		case err == io.EOF:

		case strings.Contains(s, "An existing connection was forcibly closed by the remote host"):
			return ios.ErrRemoteCloseUnusual

		case strings.Contains(s, "use of closed network connection"):
			switch {
			case strings.Contains(s, "close tcp"):
				return ios.ErrCloseClose

			case strings.Contains(s, "write tcp"):
				return ios.ErrWriteClosed

			case strings.Contains(s, "read tcp"):
				return ios.ErrReadClosed

			}

		case strings.Contains(s, "bind: An operation on a socket could not be performed because the system lacked sufficient buffer space or because a queue was full"):
			return ios.ErrPortNull

		case strings.Contains(s, "connectex: No connection could be made because the target machine actively refused it."):
			return ios.ErrRemoteOff

		case strings.Contains(s, "A socket operation was attempted to an unreachable network"):
			return ios.ErrNetworkUnusual

		}
	}
	return err
}
