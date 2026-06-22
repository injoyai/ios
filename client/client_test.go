package client

import (
	"bufio"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/injoyai/ios/v2"
)

type testReadWriteCloser struct{}

func (testReadWriteCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (testReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (testReadWriteCloser) Close() error {
	return nil
}

func TestClient_RecyclesReaderPoolBufferOnClose(t *testing.T) {
	oldPool := DefaultReaderPool
	pool := NewPool(1, func() *bufio.Reader {
		return bufio.NewReaderSize(nil, 4096)
	})
	DefaultReaderPool = pool
	defer func() { DefaultReaderPool = oldPool }()

	c := New(nil)
	conn := testReadWriteCloser{}
	c.SetReadWriteCloser("test", conn)

	if c.buf == nil {
		t.Fatalf("expected client buf to be assigned")
	}

	buf := c.buf
	if err := c.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	got := pool.Get()
	if got != buf {
		t.Fatalf("expected pooled reader to be recycled")
	}
}

type blockingReadWriteCloser struct {
	closeCh chan struct{}
}

func (b *blockingReadWriteCloser) Read(p []byte) (int, error) {
	<-b.closeCh
	return 0, io.EOF
}

func (b *blockingReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (b *blockingReadWriteCloser) Close() error {
	select {
	case <-b.closeCh:
	default:
		close(b.closeCh)
	}
	return nil
}

type stagedReadWriteCloser struct {
	reads   chan []byte
	closeCh chan struct{}
}

func (s *stagedReadWriteCloser) Read(p []byte) (int, error) {
	select {
	case <-s.closeCh:
		return 0, io.EOF
	case bs, ok := <-s.reads:
		if !ok {
			return 0, io.EOF
		}
		n := copy(p, bs)
		return n, nil
	}
}

func (s *stagedReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (s *stagedReadWriteCloser) Close() error {
	select {
	case <-s.closeCh:
	default:
		close(s.closeCh)
	}
	return nil
}

func TestClientRunTimeoutStopsWithSingleRunLifecycle(t *testing.T) {
	c := New(nil)
	c.SetReadTimeout(20 * time.Millisecond)
	conn := &blockingReadWriteCloser{closeCh: make(chan struct{})}
	c.SetDial(func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		return conn, "test", nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- c.run(ctx)
	}()

	time.Sleep(5 * time.Millisecond)
	manualErr := errors.New("manual close")
	c.CloseWithErr(manualErr)

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected run to exit")
	}

	time.Sleep(40 * time.Millisecond)
	if errors.Is(c.Closer.Err(), ios.ErrReadTimeout) {
		t.Fatalf("expected timeout goroutine to stop after run exit")
	}
	if errors.Is(c.Closer.Err(), manualErr) {
		return
	}
	if errors.Is(c.Closer.Err(), io.EOF) {
		return
	}
	t.Fatalf("expected closer err to remain manual close or EOF, got %v", c.Closer.Err())
}

func TestClientRunTimeoutKeepsAliveOnReadActivity(t *testing.T) {
	c := New(nil)
	c.SetReadTimeout(30 * time.Millisecond)
	conn := &stagedReadWriteCloser{reads: make(chan []byte, 4), closeCh: make(chan struct{})}
	c.SetDial(func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		return conn, "test", nil
	})
	c.OnDealMessage(func(_ *Client, ack ios.Acker) {
		ack.Ack()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- c.run(ctx)
	}()

	conn.reads <- []byte("a")
	time.Sleep(10 * time.Millisecond)
	conn.reads <- []byte("b")
	time.Sleep(10 * time.Millisecond)
	conn.reads <- []byte("c")
	time.Sleep(10 * time.Millisecond)

	if errors.Is(c.Closer.Err(), ios.ErrReadTimeout) {
		t.Fatalf("expected active reads to prevent timeout")
	}

	_ = conn.Close()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected run to exit after close")
	}
}

func TestClientRunTimeoutClosesWhenReadStops(t *testing.T) {
	c := New(nil)
	c.SetReadTimeout(20 * time.Millisecond)
	conn := &blockingReadWriteCloser{closeCh: make(chan struct{})}
	c.SetDial(func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		return conn, "test", nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- c.run(ctx)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected run to exit on timeout")
	}

	if !errors.Is(c.Closer.Err(), ios.ErrReadTimeout) {
		t.Fatalf("expected read timeout, got %v", c.Closer.Err())
	}
}
