package ios

var (
	_ MoreWriter      = &null{}
	_ ReadWriteCloser = &null{}
	_ Closer          = &null{}
)

var Null = &null{}

type null struct{}

func (this *null) ReadAck() (Acker, error) { return Ack(nil), nil }

func (this *null) ReadMessage() ([]byte, error) { return nil, nil }

func (this *null) Read(p []byte) (int, error) { return 0, nil }

func (this *null) ReadAt(p []byte, off int64) (int, error) { return 0, nil }

func (this *null) WriteAt(p []byte, off int64) (int, error) { return len(p), nil }

func (this *null) Write(p []byte) (int, error) { return len(p), nil }

func (this *null) WriteString(s string) (int, error) { return len(s), nil }

func (this *null) WriteByte(c byte) error { return nil }

func (this *null) WriteBase64(s string) error { return nil }

func (this *null) WriteHEX(s string) error { return nil }

func (this *null) WriteJson(a any) error { return nil }

func (this *null) WriteAny(a any) error { return nil }

func (this *null) WriteChan(c chan any) error { return nil }

func (this *null) Close() error { return nil }

func (this *null) Closed() bool { return false }
