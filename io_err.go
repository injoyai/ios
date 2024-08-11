package ios

import "errors"

var (
	ErrHandClose          = errors.New("主动关闭")
	ErrRemoteClose        = errors.New("远程端主动关闭连接")
	ErrRemoteCloseUnusual = errors.New("远程端意外关闭连接")
	ErrCloseClose         = errors.New("关闭已关闭连接")
	ErrWriteClosed        = errors.New("写入已关闭连接")
	ErrReadClosed         = errors.New("读取已关闭连接")
	ErrPortNull           = errors.New("端口不足或绑定未释放")
	ErrRemoteOff          = errors.New("远程服务可能未开启")
	ErrNetworkUnusual     = errors.New("网络异常")
	ErrWithContext        = errors.New("上下文关闭")
	ErrWithTimeout        = errors.New("超时")
	ErrWithConnectTimeout = errors.New("连接超时")
	ErrReadTimeout        = errors.New("读超时")
	ErrWriteTimeout       = errors.New("写超时")
	ErrInvalidReadFunc    = errors.New("无效数据读取函数")
	ErrMaxConnect         = errors.New("到达最大连接数")
	ErrUseReadMessage     = errors.New("不支持,请使用ReadMessage")
	ErrUseReadAck         = errors.New("不支持,请使用ReadAck")
)
