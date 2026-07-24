package server

// A client is a representation of someone connected to the server

type client struct {
	// Name of the client that is connected
	name string

	sessionId []byte
}

func newClient(name string, sessionId []byte) client {
	return client{
		name:      name,
		sessionId: sessionId,
	}
}
