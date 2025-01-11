package server

import "fmt"

// We need to add more metadata to a connection
// so that we can distingish each connection

type Broker struct {
	publishChannel     chan []byte
	unsubscribeChannel chan chan []byte
	subscribeChannel   chan chan []byte
	quit               chan struct{}
}

func NewBroker(StopSignal chan struct{}) *Broker {
	return &Broker{
		publishChannel:     make(chan []byte, 1),
		unsubscribeChannel: make(chan chan []byte, 1),
		subscribeChannel:   make(chan chan []byte, 1),
		quit:               make(chan struct{}),
	}
}


// Intended to be run as a Go routine, its dispatches messages to a channel that is shared
// between zero or multiple connections. Each connection saved in subscribers will receive messages
// that are published to the publishChannel.
func (b *Broker) Start() {
	// Save all channels to a subscribers map
	subscribers := map[chan []byte]struct{}{}
	for {
		select {
		case msgCh := <-b.subscribeChannel:
			subscribers[msgCh] = struct{}{}
		case msgCh := <-b.unsubscribeChannel:
			delete(subscribers, msgCh)
		case msg := <-b.publishChannel:
			fmt.Println("Broker received a message")
			for msgCh := range subscribers {
				select {
				case msgCh <- msg:
					fmt.Println("Broker published message to subscriber")
				default:
				}
			}
		case <-b.quit:
			return
		}
	}
}

func (b *Broker) Stop() {
	b.quit <- struct{}{}
}

func (b *Broker) Subscribe() chan []byte {
	msgCh := make(chan []byte)
	b.subscribeChannel <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh chan []byte) {
	b.unsubscribeChannel <- msgCh
}

func (b *Broker) Publish(msg []byte) {
	b.publishChannel <- msg
}
