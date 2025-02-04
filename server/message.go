package server

import (
	"fmt"
    "time"
    "encoding/json"
)

// The following message types are used to distinguish the purpose of transmissions between
// a client and the server
const (
	msgTypeHello   messageType = iota
	msgTypeMessage messageType = iota
	msgTypeBye     messageType = iota
)

type messageType int

func (m messageType) String() string {
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

// returns a raw byte slice of an Hello type message
func newRawHelloMsg() []byte {
    m := Message{
        Timestamp: time.Now().Format(time.RFC850),
        Type: msgTypeHello,
    }
    rawMsg, _ := json.Marshal(m)
    return rawMsg
}

func unmarshalMessage(data []byte) (*Message, error) {
   var msg Message
   err := json.Unmarshal(data, &msg) 
   if err != nil {
       return nil, fmt.Errorf("could not unmarshal incoming message: ", err)
   }

   return &msg, nil
}

// Checks if a message if of a given type, returns true of msg matches t
func isTypeFromRaw(raw []byte, t messageType) bool {
    var m Message
    if err := json.Unmarshal(raw, &m); err != nil {
        fmt.Println("could not unmarshal raw:", err)
        return false
    }

    if m.Type != t {
        return false
    }
    return true
}

func isHello(in []byte) (bool, *Message) {
    var m Message
    if err := json.Unmarshal(in, &m); err != nil {
        return false, nil
    }

    if m.Type != msgTypeHello {
        return false, nil
    }

    return true, &m
}

