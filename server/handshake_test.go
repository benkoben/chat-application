package server

import (
	"chat-application/lib"
	"context"
	"encoding/json"
	"net"
	"testing"
)

func TestHandshake_success(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go func() {
		if _, err := lib.WriteFixed(clientConn, lib.Hello.Bytes()); err != nil {
			t.Errorf("client hello write error: %s", err)
		}

		t.Logf("Hello sent, awaiting ACK")

		sessionBuf := make([]byte, 32)
		if _, err := lib.ReadFixed(clientConn, sessionBuf); err != nil {
			t.Errorf("could not retrieve session id: %s", err)
		}

		t.Logf("awaiting ACK from server")
		ackBuf := make([]byte, len(lib.Ack.Bytes()))
		if _, err := lib.ReadFixed(clientConn, ackBuf); err != nil {
			t.Errorf("could not retrieve ACK: %s", err)
		}

		// validate
		if len(ackBuf) != len(lib.Ack.Bytes()) {
			t.Errorf("invalid ACK length: %d", len(ackBuf))
		}

		t.Logf("session received")

		if len(sessionBuf) != 32 {
			t.Errorf("invalid session id length: %d", len(sessionBuf))
		}

		t.Logf("session id length: %d", len(sessionBuf))

		// Send client information
		info := lib.ClientDiscoveryMessage{
			SessionId: sessionBuf,
			Name:      "batman",
		}
		data, err := json.Marshal(info)
		if err != nil {
			t.Errorf("could not marshal info: %s", err)
		}

		// send the client info back to the server
		if _, err := lib.WriteFramed(clientConn, data); err != nil {
			t.Errorf("client info write error: %s", err)
		}

	}()

	t.Logf("awaiting handshake on serverConn")
	client, err := handshake(context.Background(), serverConn)
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}

	if client.name != "batman" {
		t.Errorf("client name = %q, want %q", client.name, "batman")
	}
}

// func TestHandshake_readFailure(t *testing.T) {
// 	serverConn, clientConn := net.Pipe()
// 	defer serverConn.Close()
// 	clientConn.Close()
//
// 	_, err := handshake(context.Background(), serverConn)
// 	if err == nil {
// 		t.Fatal("expected handshake error")
// 	}
//
// 	if !strings.Contains(err.Error(), "could not read the hello message from the client") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }
//
// func TestHandshake_invalidMessageType(t *testing.T) {
// 	serverConn, clientConn := net.Pipe()
// 	defer serverConn.Close()
// 	defer clientConn.Close()
//
// 	go func() {
// 		m := Message{
// 			Author:    "bob",
// 			Timestamp: time.Now().Format(time.RFC850),
// 			Type:      msgTypeMessage,
// 		}
//
// 		for {
// 			raw, err := json.Marshal(m)
// 			if err != nil {
// 				t.Errorf("marshal message: %v", err)
// 				return
// 			}
//
// 			switch {
// 			case len(raw) == 1<<10:
// 				if _, err := clientConn.Write(raw); err != nil {
// 					t.Errorf("client write: %v", err)
// 				}
// 				return
// 			case len(raw) > 1<<10:
// 				t.Fatal("message exceeds 1024 bytes")
// 			default:
// 				m.Body += "x"
// 			}
// 		}
// 	}()
//
// 	_, err := handshake(context.Background(), serverConn)
// 	if err == nil {
// 		t.Fatal("expected handshake error")
// 	}
//
// 	if !strings.Contains(err.Error(), "expected message type hello") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }
//
// func TestHandshake_malformedJSON(t *testing.T) {
// 	serverConn, clientConn := net.Pipe()
// 	defer serverConn.Close()
// 	defer clientConn.Close()
//
// 	go func() {
// 		payload := make([]byte, 1<<10)
// 		for i := range payload {
// 			payload[i] = 'x'
// 		}
// 		if _, err := clientConn.Write(payload); err != nil {
// 			t.Errorf("client write: %v", err)
// 		}
// 	}()
//
// 	_, err := handshake(context.Background(), serverConn)
// 	if err == nil {
// 		t.Fatal("expected handshake error")
// 	}
//
// 	if !strings.Contains(err.Error(), "expected message type hello") {
// 		t.Errorf("unexpected error: %v", err)
// 	}
// }
//
