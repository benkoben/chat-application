
package server

import (
	"fmt"
)

// The following message types are used to distinguish the purpose of transmissions between
// a client and the server
const (
    msgTypeHello messageType   = iota 
    msgTypeMessage messageType = iota 
    msgTypeBye messageType     = iota 
)

type messageType int

func (m messageType)String()string{
    var mType string
    switch {
        case m == 0:
            mType = "hello"
        case m == 1:
            mType = "message"
        case m == 2:
            mType = "bye"
    }

    return mType
}

type Message struct {
	Author    string      `json:"author"`
	Timestamp string      `json:"timestamp"`
	Body      string      `json:"body"`
    Type      messageType `json:"type"`
}

// Formats a message into a readable format. Omits the type field in the return value
func (m Message) String() string {
    return fmt.Sprintf("%s - %s: %s\n", m.Timestamp, m.Author, m.Body)
}
