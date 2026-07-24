package server

import (
	"chat-application/lib"
	"context"
	"log"
)

type Broker struct {
	publishChannel     chan *lib.Message
	unsubscribeChannel chan chan *lib.Message
	subscribeChannel   chan chan *lib.Message
	quit               chan struct{}
}

func NewBroker(StopSignal chan struct{}) *Broker {
	return &Broker{
		publishChannel:     make(chan *lib.Message, 1),
		unsubscribeChannel: make(chan chan *lib.Message, 100),
		subscribeChannel:   make(chan chan *lib.Message, 100),
		quit:               make(chan struct{}),
	}
}

func (b *Broker) Start() {
	// Save all channels to a subscribers map
	subscribers := map[chan *lib.Message]struct{}{}
	for {
		select {
		case msgCh := <-b.subscribeChannel:
			subscribers[msgCh] = struct{}{}
		case msgCh := <-b.unsubscribeChannel:
			delete(subscribers, msgCh)
		case msg := <-b.publishChannel:
			log.Print("Broker received a message")
			for msgCh := range subscribers {
				select {
				case msgCh <- msg:
					log.Print("Broker published message to subscriber")
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

func (b *Broker) Subscribe() chan *lib.Message {
	msgCh := make(chan *lib.Message, 10)
	b.subscribeChannel <- msgCh
	return msgCh
}

func (b *Broker) Unsubscribe(msgCh *chan *lib.Message) {
	b.unsubscribeChannel <- *msgCh
}

func (b *Broker) Publish(msg *lib.Message) {
	b.publishChannel <- msg
}

// the writer reads messages from the provided channel and publishes them to the broker until context is canceled.
func (b *Broker) writer(ctx context.Context, msgCh messageCh) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Print("Closing broker writer goroutine")
				return
			case msg, ok := <-msgCh:
				if ok {
					b.Publish(msg)
				}
			}
		}
	}()
}
