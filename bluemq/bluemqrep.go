package bluemq

import (
	"context"
	"log"
	"sync"

	"github.com/go-zeromq/zmq4"
)

// BMqRep ...
type BMqRep struct {
	_socket    zmq4.Socket
	_exitChan  chan struct{}
	_waitGroup *sync.WaitGroup
	_repChan   chan zmq4.Msg
}

// NewRep ...
func NewRep() *BMqRep {
	return &BMqRep{
		_socket:    zmq4.NewRep(context.Background()),
		_exitChan:  make(chan struct{}),     // exit channel
		_waitGroup: &sync.WaitGroup{},       // goroutine wait greoup
		_repChan:   make(chan zmq4.Msg, 10), // subscribed messagegs
	}
}

// Close ...
func (rep *BMqRep) Close() {
	rep._socket.Close()
}

// Stop ..
func (rep *BMqRep) Stop() {
	close(rep._exitChan)
}

// Start ...( "tcp://*:5559")
func (rep *BMqRep) Start(addr string) error {
	if err := rep._socket.Listen(addr); err != nil {
		log.Println(err)
		return err
	}
	rep._waitGroup.Add(1)
	go rep.tick()
	return nil
}

// Send ...
func (rep *BMqRep) Send(msg string) error {

	msgMq := zmq4.NewMsgString(msg)
	if err := rep._socket.Send(msgMq); err != nil {
		log.Println("Send ", err)
		return err
	}
	return nil
}

func (rep *BMqRep) tick() {

	log.Println("zmq rep start tick")

	defer func() {
		recover()
		rep._waitGroup.Done()
		log.Println("zmq pub tick defer")
	}()

	for {
		// check exit channel
		select {
		case <-rep._exitChan:
			return
		default:
		}

		msg, err := rep._socket.Recv()
		if err != nil {
			log.Println("tick ", err)
			continue
		}
		rep._repChan <- msg

		log.Println("recv->", string(msg.Frames[0]))

		rep.Send("OK")
	}
}

// RepChan ...
func (rep *BMqRep) RepChan() chan zmq4.Msg {
	return rep._repChan
}
