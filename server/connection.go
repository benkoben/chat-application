package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type clientConnection struct {
	connection net.Conn
	mshCh      chan *Message
	client     *client
}

const handshakeTimeout = 10 * time.Second

func newClientConnection(connection net.Conn) *clientConnection {
	return &clientConnection{
		connection: connection,
		mshCh:      make(chan *Message),
	}
}

func (c *clientConnection) close(ctx context.Context) error {
	if err := c.connection.Close(); err != nil {
		return err
	}
	return nil
}

func (c *clientConnection) transmit(ctx context.Context, msgCh *chan *Message, client *client) errCh {
	errCh := make(errCh)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-c.mshCh:
				if msg.Author != client.name {
					rawMsg, err := json.Marshal(msg)
					if err != nil {
						log.Print("could not write to connection", err)
					}

					// TODO:  Should I have a timeout here?
					if _, err := c.connection.Write(rawMsg); err != nil {
						errCh <- fmt.Errorf("could not write to connection: %s", err)
					}
				}
			}
		}
	}()

	return errCh
}

func (c *clientConnection) receive(ctx context.Context, in *chan *Message) (*messageCh, *doneCh) {
	doneCh := make(doneCh)
	out := make(messageCh)

	go func() {
		for {
			select {
			// Cancellation
			case <-ctx.Done():
				return
			default:
				chunk := make([]byte, 1<<10) // TODO: This should not be hardcoded into 1024 bytes
				reader := bufio.NewReader(c.connection)
				n, err := reader.Read(chunk)

				if err != nil {

					if err == io.EOF {
						// Signal completion
						doneCh <- struct{}{}
						return
					}

					continue
				}

				if n > 0 {
					data := chunk[:n]

					if isTypeFromRaw(data, msgTypeBye) {
						// If the client has gracefully sent a BYE message then
						// cleanup and return
						doneCh <- struct{}{}
						return

						// Completion
					}

					// Continue otherwise
					var m Message
					if err := json.Unmarshal(data, &m); err != nil {
						log.Print("could not unmarshal received message: ", err)
						continue
					}
					out <- &m
				}
			}
		}
	}()

	return &out, &doneCh
}
