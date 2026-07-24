package server

import (
	"chat-application/lib"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	helloMsgType = "HELLO"
)

func handshake(_ context.Context, conn net.Conn) (*client, error) {
	// Set read deadline
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return nil, fmt.Errorf("could not set read deadline: %w", err)
	}

	helloBuf := make([]byte, len(lib.Hello))
	if _, err := lib.ReadFixed(conn, helloBuf); err != nil {
		return nil, fmt.Errorf("handshake: %s", err)
	}

	// Validate the data's contents
	ok := isHello(helloBuf)
	if !ok {
		return nil, fmt.Errorf("expected hello message, got %s", string(helloBuf))
	}

	if _, err := lib.WriteFixed(conn, lib.Ack.Bytes()); err != nil {
		return nil, fmt.Errorf("could not send ACK message: %s", err)
	}

	log.Printf("Sent ACK message to client")

	// Generate Session-ID
	session, err := lib.GenerateSessionId()
	if err != nil {
		return nil, fmt.Errorf("could not generate session id: %w", err)
	}

	if _, err := lib.WriteFixed(conn, session.Id[:]); err != nil {
		return nil, fmt.Errorf("could not send session id to client: %w", err)
	}

	info, err := lib.ReadFramed(conn)
	if err != nil {
		return nil, fmt.Errorf("handshake: %s", err)
	}

	// Unmarshal the received data into a client discovery message
	// so that we can use that to construct a client
	var clientDiscoveryMsg lib.ClientDiscoveryMessage
	if err := json.Unmarshal(info, &clientDiscoveryMsg); err != nil {
		return nil, fmt.Errorf("handshake: %s", err)
	}

	// Reset read deadline (no more reading necessary for the handshake)
	// conn is shared.
	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("could not reset read deadline: %w", err)
	}

	//  construct client and return
	return new(newClient(clientDiscoveryMsg.Name, session.Id)), nil
}

func generateSessionId() ([]byte, error) {
	var sessionId [32]byte
	if _, err := io.ReadFull(rand.Reader, sessionId[:]); err != nil {
		return nil, err
	}
	return sessionId[:], nil
}
