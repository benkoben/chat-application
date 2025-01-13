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

        // For each new connection a new go routine is started that subscribes to
        // the server broker.
		go func() {
			fmt.Println("Starting a new worker")
			msgCh  := s.broker.Subscribe()
            buffer := make([]byte, 1<<10)
			for {
				select {
				case msg := <-msgCh:
					conn.Write(msg)
                    case <-s.quit:
					return
                default:
                    // By default just read the connection
                    // whenever something is recieved just shuffle it
                    // into the messageChannel. From there the broker will take
                    // care of distributing it.
                    n, err := conn.Read(buffer)
                    if err != nil {
                        fmt.Println("received a message but could not read it")
                    }
                    msgCh<-buffer[:n]
				}
			}
		}()
	}
}
