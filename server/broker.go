package server

import "fmt"

// We need to add more metadata to a connection
// so that we can distingish each connection

type Broker struct {
	publishChannel     chan []byte
	unsubscribeChannel chan *Subscriber
	subscribeChannel   chan *Subscriber
	quit               chan struct{}
}

func NewBroker(StopSignal chan struct{}) *Broker {
	return &Broker{
		publishChannel:     make(chan []byte, 1),
		unsubscribeChannel: make(chan *Subscriber, 1),
		subscribeChannel:   make(chan *Subscriber, 1),
		quit:               make(chan struct{}),
	}
}


// Intended to be run as a Go routine, its dispatches messages to a channel that is shared
// between zero or multiple connections. Each connection saved in subscribers will receive messages
// that are published to the publishChannel.
func (b *Broker) Start() {
	// Save all channels to a subscribers map
	subscribers := map[*Subscriber]struct{}{}
	// subscribers := map[chan []byte]string
	for {
		select {
		case subscriber := <-b.subscribeChannel:
			subscribers[subscriber] = struct{}{}
		case subscriber := <-b.unsubscribeChannel:
			delete(subscribers, subscriber)
		case msg := <-b.publishChannel:
			fmt.Println("Broker received a message")
			for subscriber := range subscribers {
				select {
				case subscriber.msgCh <- msg:
					fmt.Println("Broker published message to subscriber")
				default:
				}
			}
		case <-b.quit:
			return
		}
	}
}

type Subscriber struct {
    msgCh chan []byte 
    name string
}

func (b *Broker) Stop() {
	b.quit <- struct{}{}
}

func (b *Broker) Subscribe(name string) chan []byte {
	msgCh := make(chan []byte)
    sub := &Subscriber{
        msgCh: make(chan []byte),
        name: name,
    }

	b.subscribeChannel <- sub
	return msgCh
}

func (b *Broker) Unsubscribe(subscriber *Subscriber) {
	b.unsubscribeChannel <- subscriber
}

func (b *Broker) Publish(msg []byte) {
	b.publishChannel <- msg
}
