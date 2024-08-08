package ios

import "io"

// MultiCloser 多个关闭合并
func MultiCloser(closer ...io.Closer) io.Closer {
	return &multiCloser{closer: closer}
}

type multiCloser struct {
	closer []io.Closer
}

func (this *multiCloser) Close() (err error) {
	for _, v := range this.closer {
		if er := v.Close(); er != nil {
			err = er
		}
	}
	return
}
