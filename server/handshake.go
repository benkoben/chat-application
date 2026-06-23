package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

const (
	helloMsgType = "HELLO"
)

// TODO
// We want to change how the handshake is handled. What we are missing now is
// 1. No session management
// 2. the handshake is not reliable because of the size is not fixed length (the size of the message is not fixed because it can differentiate between clients)
// New Design:
// 1. Client sends a HELLO message type only
// 2. Server verifies this (fixed length)
// 3  Server sends HELLO-ACK to the client
// 4. Server generates a session id and sends it back to the client
// 5. Client sends information about itself to the server
// 6. Each step is suseptible to timeouts
// Considerations
// 1. Because the steps are mixed between variable and fixed size length we need clear framing -> light weight state machine?

// handshake performs the initial handshake protocol with a new client connection by exchanging hello messages
// and returns a client instance with the client's author name upon successful completion.
func handshake(_ context.Context, conn net.Conn) (*client, error) {
	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 5)); err != nil {
		return nil, fmt.Errorf("could not set read deadline: %s", err)
	}

	helloMsgBuf := make([]byte, 1<<10) // TODO: This should not be hardcoded into 1024 bytes
	n, err := conn.Read(helloMsgBuf)
	if err != nil {
		return nil, fmt.Errorf("handshake: %s", err)
	}

	log.Printf("Received hello message from client: %s", helloMsgBuf)
	ok, clientMsg := isHello(helloMsgBuf[:n])
	if !ok {
		return nil, fmt.Errorf("received invalid message type")
	}

	// Prepare a response hello message
	// This will let the client know the server has accepted the client
	helloMsg := newRawHelloMsg()
	if _, err := io.WriteString(conn, string(helloMsg)); err != nil {
		return nil, fmt.Errorf("could not respond to client: %s", err)
	}

	if err := conn.SetReadDeadline(time.Time{}); err != nil {
		return nil, fmt.Errorf("could not clear read deadline: %s", err)
	} // clear read deadline when exiting
	return new(newClient(clientMsg.Author)), nil
}
