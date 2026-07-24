// build a simple echo server
package server

import (
	"chat-application/lib"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "7007"
)

type doneCh chan struct{}
type messageCh chan *lib.Message
type errCh chan error

type server struct {
	host   string
	port   string
	broker *Broker
	quit   chan struct{}
}

type ServerOpts struct {
	Broker *Broker
	Host   string
	Port   string
}

func NewServer(o ServerOpts) (*server, error) {
	if o.Host == "" {
		o.Host = defaultHost
	}

	if o.Port == "" {
		o.Port = defaultPort
	}

	if o.Broker == nil {
		return nil, fmt.Errorf("broker cannot be nil")
	}

	quit := make(chan struct{}, 1)

	return &server{
		host:   o.Host,
		port:   o.Port,
		broker: o.Broker,
		quit:   quit,
	}, nil
}

func (s server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))
	if err != nil {
		return fmt.Errorf("could not open socket: %s", err)
	}
	log.Printf("Listening on %s:%s\n", s.host, s.port)

	defer ln.Close()

	go s.broker.Start()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ACCEPT FAILED - ADDR: %s", conn.RemoteAddr())
			continue
		}
		log.Print("Accepted new connection")
		go s.handleConnection(conn)
	}
}

func (s server) handleConnection(connection net.Conn) {
	defer connection.Close()

	// Deadline
	ctx := context.Background()

	// The handshake establishes a formal connection between the client and the server
	// It is used to authenticate the client and learn information about the client.
	// timeout is set to 30 seconds.
	client, err := handshake(ctx, connection)
	if err != nil {
		log.Printf("Handshake failed - ADDR: %s, ERR: %s", connection.RemoteAddr(), err)
		return
	}
	log.Printf("Handshake success - ADDR: %s CLIENT: %s", connection.RemoteAddr(), client.name)

	brokerMsgCh := s.broker.Subscribe()

	// Create a new client connection. This will set new read and write deadlines
	clientConnection, err := newClientConnection(connection)
	if err != nil {
		log.Printf("Could not create client connection - ADDR: %s CLIENT: %s", connection.RemoteAddr(), client.name)
		return
	}

	workersCtx, workerCancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}

	// Read from the connection and write to msgCh
	msgCh, rxErrs, receiverDone := clientConnection.receive(workersCtx)
	// The broker will write to the brokerMsgCh
	s.broker.writer(workersCtx, msgCh)
	// The transmitter will read from the broker channel and write to the connection
	txErrs := clientConnection.transmit(workersCtx, brokerMsgCh, client)

	// Capture signals from the workers
	wg.Go(func() {
		for {
			select {
			case <-receiverDone:
				// If the receiver is done, close the connection
				if err := clientConnection.Close(ctx); err != nil {
					log.Printf("Connection failed to close properly - ADDR: %s CLIENT: %s, MSG: %s",
						connection.RemoteAddr(),
						client.name,
						err,
					)
				} else {
					log.Printf("Connection terminated - ADDR: %s CLIENT: %s", connection.RemoteAddr(), client.name)
				}
				// Signal cancellation to the workers
				workerCancel()
				return
			case <-workersCtx.Done():
				// Clean up this goroutine if the workers are done
				return
			case err, ok := <-txErrs:
				if ok {
					log.Printf("Write connection error - ADDR: %s, CLIENT: %s, MSG: %s", connection.RemoteAddr(), client.name, err)
				}
			case err, ok := <-rxErrs:
				if ok {
					log.Printf("Read connection error - ADDR: %s, CLIENT: %s, MSG: %s", connection.RemoteAddr(), client.name, err)
				}
			}
		}
	})

	wg.Wait()
}
