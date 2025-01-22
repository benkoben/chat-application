// build a simple echo server
package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
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
        // client.handleConnection()
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

    clientMsgCh := s.broker.Subscribe()
    var wg sync.WaitGroup
    // handle writing to connection
    fmt.Println("Starting client writer worker")

    wg.Add(1)
    go func(){
        defer wg.Done()
        for {
            select {
            case msg:=<-*clientMsgCh:
                // Prevent a message from echoing back to the user
                if msg.Author != client.name {
                    jsonRaw, err := json.Marshal(&msg)
                    if err != nil {
                        fmt.Println("could not marshal message received from client channel:", err)
                    }
    
    
                    io.WriteString(connection, string(jsonRaw))
                }
            case <-s.quit:
                return
            }
        }
    }()
    wg.Add(1)
    // handle reading from the connection
    fmt.Println("Starting client reader worker")
    go func(){
        defer wg.Done()
        //TODO: Here we should handle a reader timeout as well.
        readerBuffer := make([]byte, 1<<10)
        for {
            select {
                case <-s.quit:
                    return
                default:
                    // TODO: I think if we handle this with a bufio reader/scanner, we can actually terminate the connection once EOF has been reached.
                    n,  err := connection.Read(readerBuffer)
                    if err != nil {
                        fmt.Println("could to read from the connection", err)
                        continue
                    }
                    msg, err := unmarshalMessage(readerBuffer[:n]) 
                    s.broker.Publish(msg)
            }
        }
    }()

    wg.Wait()
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
