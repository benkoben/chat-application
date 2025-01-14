// build a simple echo server
package server

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "7007"
)

type BrokerSvc interface {
	Start()
	Stop()
	Publish([]byte)
	Subscribe() chan []byte
	Unsubscribe(chan []byte)
}

type server struct {
	host   string
	port   string
	broker BrokerSvc
	quit   chan struct{}
}

type ServerOpts struct {
	Broker     BrokerSvc
	Host       string
	Port       string
	StopSignal chan struct{}
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

	return &server{
		host:   o.Host,
		port:   o.Port,
		broker: o.Broker,
	}, nil
}

func (s server) Start() error {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))
	if err != nil {
		return fmt.Errorf("could not open socket: %s", err)
	}
	fmt.Printf("Listening on %s:%s\n", s.host, s.port)

	defer ln.Close()

	go s.broker.Start()

	go func() {
		for {
            select {
            case <-s.quit:
                return
            default:
			    time.Sleep(time.Second * 20)

                m := Message{
                    Timestamp: time.Now().Format(time.RFC850),
                    Author: "system",
                    Body: "This is a broadcasted message",
                }

                rawMsg, _ := json.Marshal(m)
                fmt.Println(string(rawMsg))
			    s.broker.Publish(rawMsg)
            }
		}
	}()

	for {
		conn, _ := ln.Accept()
		fmt.Println("Accepted new connection")

        // For each accepted connection
        // the server waits for a client handshake and construct a client. 
        client, err := handshake(conn)
        if err != nil {
            fmt.Printf("handshake with %s failed: %s\n", conn.RemoteAddr(), err)
        }

        // When the client successfully is instantiated then subscribe to the broker
        // and start a go routine that handles the connection with the client.
        msgCh  := s.broker.Subscribe()
        go client.handleConnection(&msgCh, &s.quit)
	}
}

func handshake(conn net.Conn) (*client, error) {
        // Here we need to implement a handshake to exchange client information.
        helloMsgBuf := make([]byte, 1<<10)
        n, err := conn.Read(helloMsgBuf)
        if err != nil {
            return nil, fmt.Errorf("could not read the hello message from the client: %s", err)
        }

        var helloMsg Message
        if err := json.Unmarshal(helloMsgBuf[:n], &helloMsg); err != nil {
            return nil, fmt.Errorf("could not unmarshal helloMessage into Message type: %s", err)
        }

        if helloMsg.Type != msgTypeHello {
            return nil, fmt.Errorf("invalid message type, expected hello but received %s", helloMsg.Type)
        }

        c, err := newClient(conn, helloMsg.Author)
        if err != nil {
            return nil, fmt.Errorf("could not construct client type: %s", err)
        }

        return c, nil
}
