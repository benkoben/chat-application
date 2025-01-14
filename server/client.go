package server

import (
	"fmt"
	"net"
)

// A wrapper around net.Conn adding additional
// data and methods for handling individual client connections
type client struct {
    // Embedding of a net.Conn
    net.Conn

    // Name of the client that is connected
    name string
}

func newClient(conn net.Conn, name string) (*client, error) {
    
    if conn == nil {
        return nil, fmt.Errorf("conn cannot be nil")
    }

    if name == "" {
        return nil, fmt.Errorf("name cannot be empty")
    }

    return &client{
        conn,
        name,
    }, nil
}

// closes a connection
func (c *client) close(){
    fmt.Printf("Closing connection to %s\n", c.Conn.RemoteAddr())
    c.Conn.Close()
}

func (c *client) handleConnection(msgCh *chan []byte, quit *chan struct{}) {
    buffer := make([]byte, 1<<10)
	for {
		select {
		case msg := <-*msgCh:
			c.Conn.Write(msg)
        case <-*quit:
		    return
        default:
            // By default just read the connection
            // whenever something is recieved redirect it
            // into the messageChannel. From there the broker will take
            // care of distributing it.
            n, err := c.Conn.Read(buffer)
            if err != nil {
                fmt.Println("received a message but could read it correctly:", err)
            }
            *msgCh<-buffer[:n]
		}
	}
}
