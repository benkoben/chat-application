package lib

type ConnectionMessageType string

// Constants that are used to represent the different types of messages that can be sent over the connection
const (
	Hello ConnectionMessageType = "HELLO"
	Bye   ConnectionMessageType = "BYE"
)
