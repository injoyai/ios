package ios

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestBridgeStopsBothDirectionsAfterOneSideCloses(t *testing.T) {
	leftA, leftB := net.Pipe()
	rightA, rightB := net.Pipe()

	done := make(chan error, 1)
	go func() {
		done <- Bridge(leftA, rightA)
	}()

	payload := []byte("ping")
	if _, err := leftB.Write(payload); err != nil {
		t.Fatalf("write left payload: %v", err)
	}
	buf := make([]byte, len(payload))
	if _, err := io.ReadFull(rightB, buf); err != nil {
		t.Fatalf("read right payload: %v", err)
	}
	if string(buf) != string(payload) {
		t.Fatalf("expected %q, got %q", payload, buf)
	}

	if err := leftB.Close(); err != nil {
		t.Fatalf("close left peer: %v", err)
	}

	select {
	case err := <-done:
		if err == nil {
			return
		}
	case <-time.After(time.Second):
		t.Fatalf("expected Bridge to return after one side closes")
	}

	if _, err := rightB.Write([]byte("late")); err == nil {
		t.Fatalf("expected right peer write to fail after bridge shutdown")
	}
}
