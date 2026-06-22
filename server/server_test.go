package server

import (
	"errors"
	"testing"

	"github.com/injoyai/ios/v2"
	"github.com/injoyai/ios/v2/client"
)

func TestServerClientsReturnsSnapshot(t *testing.T) {
	s := &Server{
		client: map[string]*client.Client{
			"a": client.New(nil),
			"b": client.New(nil),
		},
	}

	seq := s.Clients()

	s.clientMu.Lock()
	delete(s.client, "a")
	delete(s.client, "b")
	s.clientMu.Unlock()

	got := map[string]bool{}
	for key, cli := range seq {
		if cli == nil {
			t.Fatalf("expected client for key %s", key)
		}
		got[key] = true
	}

	if len(got) != 2 || !got["a"] || !got["b"] {
		t.Fatalf("expected snapshot of original clients, got %#v", got)
	}
}

func TestServerCloseClientDoesNotDeleteReplacedClient(t *testing.T) {
	s := &Server{client: make(map[string]*client.Client)}
	oldClient := client.New(nil)
	newClient := client.New(nil)
	_ = errors.New("close target")

	s.client["same"] = newClient

	s.clientMu.Lock()
	if s.client["same"] == oldClient {
		delete(s.client, "same")
	}
	got := s.client["same"]
	s.clientMu.Unlock()
	if got != newClient {
		t.Fatalf("expected replaced client to remain, got %#v", got)
	}
}

func TestServerOnChangeKeyReplacesMapping(t *testing.T) {
	s := &Server{client: make(map[string]*client.Client)}
	cli := client.New(nil)
	cli.SetKey("new")
	oldCli := client.New(nil)
	oldCli.SetKey("new")

	s.client["old"] = cli
	s.client["new"] = oldCli

	s.onChangeKey(cli, "old")

	s.clientMu.RLock()
	defer s.clientMu.RUnlock()
	if _, ok := s.client["old"]; ok {
		t.Fatalf("expected old key to be removed")
	}
	if got := s.client["new"]; got != cli {
		t.Fatalf("expected new key to point to current client, got %#v", got)
	}
}

type staticListener struct {
	accept func() (ios.ReadWriteCloser, string, error)
}

func (l staticListener) Accept() (ios.ReadWriteCloser, string, error) {
	return l.accept()
}

func (l staticListener) Close() error { return nil }
func (l staticListener) Addr() string { return "test" }
