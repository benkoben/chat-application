package server

// A client is a representation of someone connected to the server

type client struct {
	// Name of the client that is connected
	name string
}

func newClient(name string) client {
	return client{
		name: name,
	}
}
