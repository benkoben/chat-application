// build a simple echo server
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
    "bufio"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "7007"
)

type server struct {
	host   string
	port   string
	broker *Broker
	quit   chan struct{}
    connections map[client]net.Conn
}

// When a new connection is received
// then add the connection to a channel

// Have a go routine modify the server.connections

type ServerOpts struct {
	Broker     *Broker
	Host       string
	Port       string
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
        quit: quit,
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
					Author:    "system",
					Body:      "This is a broadcasted message",
				}

				s.broker.Publish(&m)
			}
		}
	}()

	for {
        fmt.Println("Waiting for new connection")
        conn, err := ln.Accept()
        if err != nil {
            fmt.Printf("failed to accept connection: %s\n", err)
            continue
        }
        fmt.Println("Accepted new connection")
        go s.handleConnection(conn)
    }
}

func (s server)handleConnection(connection net.Conn){
    // TODO, we should set a max timeout value on the connection
    defer connection.Close()

    client, err := handshake(connection) 
    if err != nil {
        fmt.Printf("could not shake hands: %s", err)
        return 
    }
	fmt.Println("Handshake complete")

	msgCh := s.broker.Subscribe()
    var wg sync.WaitGroup
    clientDisconnect := make(chan struct{})

    wg.Add(1)
    // Writer worker
    go func(){
        fmt.Println("writer for client", client.name, "up")
        defer wg.Done()
        defer fmt.Println("writer for client", client.name, "exiting...")
	    for {
	    	select {
	    	case msg := <-*msgCh:
                rawMsg, err := json.Marshal(msg)
                if err != nil {
                    fmt.Println("could not write to connection", err)
                }
	    		connection.Write(rawMsg)
	    	case <-clientDisconnect:
	    		return
	    	}
	    }
    }()

    wg.Add(1)
    // Reader worker
    go func(){
        fmt.Println("reader for client", client.name, "up")
        
        defer wg.Done()
        defer fmt.Println("reader for client", client.name, "exiting...")

        reader := bufio.NewReader(connection)
        for {
            select {
                case <- clientDisconnect:
                    return
                default:
                    data, err := io.ReadAll(reader)
                    if err != nil {
                        fmt.Println("could not read from client", err)
                        continue
                    }

                    if isTypeFromRaw(data, msgTypeBye){
                        fmt.Println("received bye message from client", client.name)
                        clientDisconnect <- struct{}{}
                        return
                    }
                    
                    if data != nil {
                        var m *Message
                        if err := json.Unmarshal(data, m); err != nil {
                            fmt.Println("could not unmarshal received message:", err)
                            continue
                        }
                        *msgCh <- m
                    }
            }
        }
    }()

    wg.Wait()
    fmt.Println("both workers have finished, exiting the connection handler...")
}

// Handshake is used to read the first initial message sent from a client.
// It is a blocking function untill the first recevied message is read from conn.
//
// If the message is valid handshake will let the client know the server has accepted the client
// and constructs a client type which is returned
func handshake(conn net.Conn) (*client, error) {
	// Here we need to implement a handshake to exchange client information.
	helloMsgBuf := make([]byte, 1<<10)
    n, err := conn.Read(helloMsgBuf)
    if err != nil {
    	return  nil, fmt.Errorf("could not read the hello message from the client: %s", err)
    }
    
    ok, clientMsg := isHello(helloMsgBuf[:n])
    if !ok{
        return nil, fmt.Errorf("could not perform handshake, expected message type hello but recieved %s", clientMsg.Type)
    }
    
    // Prepare a response hello message
    // This will let the client know the server has accepted the client
    helloMsg := newRawHelloMsg()
    if _, err := io.WriteString(conn, string(helloMsg)); err != nil {
        return nil, fmt.Errorf("could not respond to client: %s", err)
    }

    fmt.Println(clientMsg) 
    client := newClient(clientMsg.Author) 
    return &client,nil
}
