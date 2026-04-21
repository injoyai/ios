package ios

var (
	_ MoreWriter      = Null
	_ ReadWriteCloser = Null
	_ Closer          = Null
)

const (
	Null    = null(0)
	Discard = Null
)

type null int8

func (null) ReadAck() (Acker, error) { return Ack(nil), nil }

func (null) ReadMessage() ([]byte, error) { return nil, nil }

func (null) Read(p []byte) (int, error) { return 0, nil }

func (null) ReadAt(p []byte, off int64) (int, error) { return 0, nil }

func (null) WriteAt(p []byte, off int64) (int, error) { return len(p), nil }

func (null) Write(p []byte) (int, error) { return len(p), nil }

func (null) WriteString(s string) (int, error) { return len(s), nil }

func (null) WriteByte(c byte) error { return nil }

func (null) WriteBase64(s string) error { return nil }

func (null) WriteHEX(s string) error { return nil }

func (null) WriteJson(a any) error { return nil }

func (null) WriteAny(a any) error { return nil }

func (null) WriteChan(c chan any) error { return nil }

func (null) Close() error { return nil }

func (null) Closed() bool { return false }
