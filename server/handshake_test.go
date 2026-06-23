package server

import (
	"context"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

// paddedHelloBytes returns a hello message marshaled to exactly 1024 bytes,
// matching the fixed read size used by handshake.
func paddedHelloBytes(author string) []byte {
	m := Message{
		Author:    author,
		Timestamp: time.Now().Format(time.RFC850),
		Type:      msgTypeHello,
	}

	for {
		raw, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}

		switch {
		case len(raw) == 1<<10:
			return raw
		case len(raw) > 1<<10:
			panic("hello message exceeds 1024 bytes")
		default:
			m.Body += "x"
		}
	}
}

func TestHandshake_success(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	const author = "alice"

	go func() {
		if _, err := clientConn.Write(paddedHelloBytes(author)); err != nil {
			t.Errorf("client write hello: %v", err)
			return
		}

		buf := make([]byte, 1<<10)
		n, err := clientConn.Read(buf)
		if err != nil {
			t.Errorf("client read response: %v", err)
			return
		}

		ok, _ := isHello(buf[:n])
		if !ok {
			t.Errorf("expected hello response, got %q", buf[:n])
		}
	}()

	client, err := handshake(context.Background(), serverConn)
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}

	if client.name != author {
		t.Errorf("client name = %q, want %q", client.name, author)
	}
}

func TestHandshake_readFailure(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	clientConn.Close()

	_, err := handshake(context.Background(), serverConn)
	if err == nil {
		t.Fatal("expected handshake error")
	}

	if !strings.Contains(err.Error(), "could not read the hello message from the client") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandshake_invalidMessageType(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		m := Message{
			Author:    "bob",
			Timestamp: time.Now().Format(time.RFC850),
			Type:      msgTypeMessage,
		}

		for {
			raw, err := json.Marshal(m)
			if err != nil {
				t.Errorf("marshal message: %v", err)
				return
			}

			switch {
			case len(raw) == 1<<10:
				if _, err := clientConn.Write(raw); err != nil {
					t.Errorf("client write: %v", err)
				}
				return
			case len(raw) > 1<<10:
				t.Fatal("message exceeds 1024 bytes")
			default:
				m.Body += "x"
			}
		}
	}()

	_, err := handshake(context.Background(), serverConn)
	if err == nil {
		t.Fatal("expected handshake error")
	}

	if !strings.Contains(err.Error(), "expected message type hello") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandshake_malformedJSON(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		payload := make([]byte, 1<<10)
		for i := range payload {
			payload[i] = 'x'
		}
		if _, err := clientConn.Write(payload); err != nil {
			t.Errorf("client write: %v", err)
		}
	}()

	_, err := handshake(context.Background(), serverConn)
	if err == nil {
		t.Fatal("expected handshake error")
	}

	if !strings.Contains(err.Error(), "expected message type hello") {
		t.Errorf("unexpected error: %v", err)
	}
}
