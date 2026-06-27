package server

import (
	"chat-application/lib"
	"encoding/json"
	"fmt"
	"time"
)

// The following message types are used to distinguish the purpose of transmissions between
// a client and the server
// TODO: lets move away from using these constants and use the libs instead
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
	Author    string `json:"author"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
	Type      messageType
}

// Formats a message into a readable format. Omits the type field in the return value
func (m Message) String() string {
	return fmt.Sprintf("%s - %s: %s\n", m.Timestamp, m.Author, m.Body)
}

// returns a raw byte slice of a Hello type message
func newRawHelloMsg() []byte {
	m := Message{
		Timestamp: time.Now().Format(time.RFC850),
		Type:      msgTypeHello,
	}
	rawMsg, _ := json.Marshal(m)
	return rawMsg
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

func isHello(in []byte) bool {
	if in == nil {
		return false
	}

	if len(in) < len(lib.Hello) {
		return false
	}

	if string(in) != string(lib.Hello) {
		return false
	}

	return true
}

func isBye(in []byte) bool {
	if in == nil {
		return false
	}

	if len(in) < len(lib.Hello) {
		return false
	}

	if string(in) != string(lib.Bye) {
		return false
	}

	return true
}
