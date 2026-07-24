package lib

type ConnectionMessageType string

// Constants that are used to represent the different types of messages that can be sent over the connection
// Each message is 4 bytes
const (
	Hello ConnectionMessageType = "HLLO"

	Ack ConnectionMessageType = "ACK\000"

	Bye ConnectionMessageType = "BYE\000"
)

func (ct ConnectionMessageType) Bytes() []byte {
	return []byte(ct)
}

type ClientDiscoveryMessage struct {
	// We probably want the client to send back the sessionId
	SessionId []byte `json:"sessionId"`
	Name      string `json:"name"`
}
