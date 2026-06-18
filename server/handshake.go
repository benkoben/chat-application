package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

// handshake performs the initial handshake protocol with a new client connection by exchanging hello messages
// and returns a client instance with the client's author name upon successful completion.
func handshake(_ context.Context, conn net.Conn) (*client, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	defer conn.SetReadDeadline(time.Time{}) // clear read deadline when exiting

	helloMsgBuf := make([]byte, 1<<10) // TODO: This should not be hardcoded into 1024 bytes
	n, err := conn.Read(helloMsgBuf)
	if err != nil {
		return nil, fmt.Errorf("could not read the hello message from the client: %s", err)
	}

	ok, clientMsg := isHello(helloMsgBuf[:n])
	if !ok {
		return nil, fmt.Errorf("could not perform handshake, expected message type hello but recieved %s", clientMsg.Type)
	}

	// Prepare a response hello message
	// This will let the client know the server has accepted the client
	helloMsg := newRawHelloMsg()
	if _, err := io.WriteString(conn, string(helloMsg)); err != nil {
		return nil, fmt.Errorf("could not respond to client: %s", err)
	}

	client := newClient(clientMsg.Author)
	return &client, nil
}
