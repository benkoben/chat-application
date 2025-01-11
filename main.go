package main

import (
	"chat-application/server"
	"fmt"
)

var (
	globalQuit = make(chan struct{})
)

func quit() {
	globalQuit <- struct{}{}
}

func main() {
	defer quit()
	b := server.NewBroker(globalQuit)

	opts := server.ServerOpts{
		Host:   "0.0.0.0",
		Port:   "7007",
		Broker: b,
	}

	s, err := server.NewServer(opts)
	if err != nil {
		panic(fmt.Sprintf("could not create new server: %s", err))
	}

	s.Start()
}
