package ios

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type errAckReader struct{}

func (errAckReader) ReadAck() (Acker, error) {
	return nil, errors.New("read ack failed")
}

func TestAllReadReadBytesReturnsErrorWhenAckIsNil(t *testing.T) {
	reader := NewAllReader(errAckReader{}, nil)

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("expected no panic, got %v", recovered)
		}
	}()

	bs, err := reader.ReadBytes()
	if err == nil {
		t.Fatalf("expected error")
	}
	if bs != nil {
		t.Fatalf("expected nil bytes, got %v", bs)
	}
}

func TestReadPrefixMatchesDirectly(t *testing.T) {
	got, err := ReadPrefix(bytes.NewBufferString("abc123"), []byte("abc"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if string(got) != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
}

func TestReadPrefixMatchesOverlap(t *testing.T) {
	got, err := ReadPrefix(bytes.NewBufferString("abababc"), []byte("ababc"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if string(got) != "ababc" {
		t.Fatalf("expected ababc, got %q", got)
	}
}

func TestReadPrefixReturnsCacheOnEOF(t *testing.T) {
	got, err := ReadPrefix(bytes.NewBufferString("ab"), []byte("abc"))
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF, got %v", err)
	}
	if string(got) != "ab" {
		t.Fatalf("expected ab, got %q", got)
	}
}
