package server

// A wrapper around net.Conn adding additional
// data and methods for handling individual client connections
type client struct {
	// Name of the client that is connected
	name string
}

func newClient(name string) client { 
	return client{
        name: name,
	}
}

