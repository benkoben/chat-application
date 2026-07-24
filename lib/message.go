package lib

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Message struct {
	Author    string `json:"author"`
	SessionId []byte `json:"sessionId"`
	Timestamp string `json:"timestamp"`
	Body      string `json:"body"`
}

func (m *Message) Unmarshal(raw []byte) error {
	// Unmarshal the raw bytes into a Message
	if err := json.Unmarshal(raw, m); err != nil {
		return err
	}

	// Sanitize the message body
	m.Body = strings.TrimSuffix(m.Body, "\n")

	return nil
}

// Formats a message into a readable format. Omits the type field in the return value
func (m *Message) String() string {
	return fmt.Sprintf("%s - %s (%s): %s\n", m.Timestamp, m.Author, string(m.SessionId), m.Body)
}

func (m *Message) Bytes() []byte {
	return []byte(m.String())
}
