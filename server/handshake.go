package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	connectionTimeout = 10 * time.Second
)

/*
Handshake is used to read the first initial message sent from a client.
It is a blocking function until the first received message is read from conn.

If the message is valid handshake will let the client know the server has accepted the client
and constructs a client type which is returned
*/
func handshake(ctx context.Context, conn net.Conn) (*client, error) {
	// Here we need to implement a handshake to exchange client information.

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
