// build a simple echo server
package server

import (
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
type messageCh chan *Message
type errCh chan error

type server struct {
	host        string
	port        string
	broker      *Broker
	quit        chan struct{}
	connections map[client]net.Conn
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
	handshakeCtx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()

	client, err := handshake(handshakeCtx, connection)
	if err != nil {
		log.Printf("Handshake failed - ADDR: %s", connection.RemoteAddr())
		return
	}
	log.Printf("Handshake success - ADDR: %s CLIENT: %s", connection.RemoteAddr(), client.name)

	brokerMsgCh := s.broker.Subscribe()
	clientConnection := newClientConnection(connection)

	workersCtx, workerCancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}

	// Read from the connection and write to msgCh
	msgCh, rxErrs, done := clientConnection.receive(workersCtx)
	// The broker will write to the brokerMsgCh
	s.broker.writer(workersCtx, msgCh)
	// The transmitter will read from the broker channel and write to the connection
	txErrs := clientConnection.transmit(workersCtx, brokerMsgCh, client)

	wg.Go(func() {
		for {
			select {
			case <-done:
				if err := clientConnection.Close(ctx); err != nil {
					log.Printf("Connection failed to close properly - ADDR: %s CLIENT: %s, MSG: %s",
						connection.RemoteAddr(),
						client.name,
						err,
					)
				}
				log.Printf("Connection terminated - ADDR: %s CLIENT: %s", connection.RemoteAddr(), client.name)

				// Signal cancellation to the workers
				workerCancel()
				return
			case <-workersCtx.Done():
				// Clean up this goroutine
				return
			case err := <-txErrs:
				log.Printf("Write connection error - ADDR: %s, CLIENT: %s, MSG: %s", connection.RemoteAddr(), client.name, err)
			case err := <-rxErrs:
				log.Printf("Read connection error - ADDR: %s, CLIENT: %s, MSG: %s", connection.RemoteAddr(), client.name, err)
			}
		}
	})

	wg.Wait()
}
