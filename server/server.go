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
			log.Printf("failed to accept connection: %s\n", err)
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
		log.Printf("could not shake hands: %s", err)
		return
	}
	log.Print(client.name, "connected from ", connection.RemoteAddr())

	brokerMsgCh := s.broker.Subscribe()
	clientConnection := newClientConnection(connection)

	workersCtx, cancel := context.WithCancel(ctx)
	wg := sync.WaitGroup{}
	// Writes to net.Conn if a message from broker is received
	errCh := clientConnection.transmit(workersCtx, brokerMsgCh, client)
	msgCh, doneCh := clientConnection.receive(workersCtx, brokerMsgCh)

	s.broker.writer(workersCtx, msgCh)
	wg.Go(func() {
		for {
			select {
			case <-*doneCh:
				cancel()
			case <-workersCtx.Done():
				return
			case err := <-errCh:
				log.Print("error in client connection: ", err)
			}
		}
	})

	//var wg sync.WaitGroup
	//clientDisconnect := make(chan struct{})
	//
	//wg.Add(1)
	//// Writer worker
	//go func() {
	//	defer wg.Done()
	//	for {
	//		select {
	//		case msg := <-*msgCh:
	//
	//			// Filter out messages than originate from the same client
	//			// in order to prevent echoing.
	//			if msg.Author != client.name {
	//				rawMsg, err := json.Marshal(msg)
	//				if err != nil {
	//					log.Print("could not write to connection", err)
	//				}
	//				connection.Write(rawMsg)
	//			}
	//		case <-clientDisconnect:
	//			// Cancellation
	//			return
	//		}
	//	}
	//}()
	//
	//wg.Add(1)
	//
	//// Reader worker
	//go func() {
	//	defer wg.Done()
	//	for {
	//		select {
	//		// Cancellation
	//		case <-clientDisconnect:
	//			return
	//		default:
	//			chunk := make([]byte, 1<<10)
	//			reader := bufio.NewReader(connection)
	//			n, err := reader.Read(chunk)
	//
	//			if err != nil {
	//
	//				if err == io.EOF {
	//					// If there fore some reason an EOF is encountered
	//					// the connection is terminated. Cleanup and return
	//					log.Print(client.name, " disconnected")
	//					// Completion
	//					clientDisconnect <- struct{}{}
	//					return
	//				}
	//
	//				log.Print("could not read from client", client.name, err)
	//				continue
	//			}
	//
	//			if n > 0 {
	//				data := chunk[:n]
	//
	//				if isTypeFromRaw(data, msgTypeBye) {
	//					// If the client has gracefully sent a BYE message then
	//					// cleanup and return
	//					log.Print(client.name, " disconnected")
	//					clientDisconnect <- struct{}{}
	//					return
	//
	//					// Completion
	//				}
	//
	//				// Continue otherwise
	//				var m Message
	//				if err := json.Unmarshal(data, &m); err != nil {
	//					log.Print("could not unmarshal received message: ", err)
	//					continue
	//				}
	//				s.broker.Publish(&m)
	//			}
	//		}
	//	}
	//}()
	//
	//wg.Wait()
	// Return and cleanup if both reader and writer go routines are done.
}
