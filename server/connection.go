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
	client     *client
}

const handshakeTimeout = 10 * time.Second

func newClientConnection(connection net.Conn) *clientConnection {
	return &clientConnection{
		connection: connection,
	}
}

func (c *clientConnection) Close(ctx context.Context) error {
	if err := c.connection.Close(); err != nil {
		return err
	}
	return nil
}

// transmit starts a goroutine that writes messages from msgCh to the connection, filtering out messages
// authored by the specified client. Messages are marshaled to JSON before transmission. Returns an error channel
// that reports write failures. The goroutine respects context cancellation and stops when ctx is done.
func (c *clientConnection) transmit(ctx context.Context, msgCh chan *Message, client *client) errCh {
	errCh := make(errCh)
	go func() {
		defer close(errCh)
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Closing transmit goroutine")
				return
			case msg := <-msgCh:
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

// receive starts a goroutine that reads messages from the client connection and returns channels for messages
// and completion signals. It reads data in 1KB chunks, unmarshals JSON messages, and sends them to the message
// channel. The method handles context cancellation, EOF conditions, and BYE message types by signaling completion
// through the done channel. Invalid messages are logged and skipped without stopping the receiver.
func (c *clientConnection) receive(ctx context.Context) (*messageCh, errCh, doneCh) {
	done := make(doneCh)
	rxErrs := make(errCh)
	out := make(messageCh)

	go func() {
		// Clean up on exit
		//defer close(out) // TODO: This is currently causing a panic
		defer close(rxErrs)
		defer close(done)

		for {
			select {
			case <-ctx.Done():
				log.Print("Closing receiver goroutine") // TODO: This is not run when context is cancelled
				return
			default:
				chunk := make([]byte, 1<<10) // TODO: This should not be hardcoded into 1024 bytes
				reader := bufio.NewReader(c.connection)
				n, err := reader.Read(chunk)
				if err != nil {

					if err == io.EOF {
						// Signal completion
						done <- struct{}{}
						return
					}

					continue
				}

				if n > 0 {
					data := chunk[:n]

					if isTypeFromRaw(data, msgTypeBye) {
						// If the client has gracefully sent a BYE message, then
						// cleanup and return
						done <- struct{}{}
						return
					}

					// Continue otherwise
					var m Message
					if err := json.Unmarshal(data, &m); err != nil {
						rxErrs <- fmt.Errorf("could not unmarshal received message: %s", err)
						continue
					}
					out <- &m
				}
			}
		}
	}()

	return &out, rxErrs, done
}
