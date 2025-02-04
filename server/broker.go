package server

import "log"

type Broker struct {
	publishChannel     chan *Message
	unsubscribeChannel chan chan *Message
	subscribeChannel   chan chan *Message
	quit               chan struct{}
}

func NewBroker(StopSignal chan struct{}) *Broker {
	return &Broker{
		publishChannel:     make(chan *Message, 1),
		unsubscribeChannel: make(chan chan *Message, 100),
		subscribeChannel:   make(chan chan *Message, 100),
		quit:               make(chan struct{}),
	}
}

func (b *Broker) Start() {
	// Save all channels to a subscribers map
	subscribers := map[chan *Message]struct{}{}
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

func (b *Broker) Subscribe() *chan *Message {
	msgCh := make(chan *Message, 10)
	b.subscribeChannel <- msgCh
	return &msgCh
}

func (b *Broker) Unsubscribe(msgCh *chan *Message) {
	b.unsubscribeChannel <- *msgCh
}

func (b *Broker) Publish(msg *Message) {
	b.publishChannel <- msg
}
