package bluemq

import (
	"context"
	"log"
	"sync"

	"github.com/go-zeromq/zmq4"
)

// BMqSub ...
type BMqSub struct {
	_socket    zmq4.Socket
	_exitChan  chan struct{}
	_waitGroup *sync.WaitGroup
	_subChan   chan zmq4.Msg
}

// NewSub ...
func NewSub() *BMqSub {
	return &BMqSub{
		_socket:    zmq4.NewSub(context.Background()),
		_exitChan:  make(chan struct{}),     // exit channel
		_waitGroup: &sync.WaitGroup{},       // goroutine wait group
		_subChan:   make(chan zmq4.Msg, 10), // subscribed messagegs
	}
}

// Close ...
func (sub *BMqSub) Close() {
	sub._socket.Close()
}

// Start ... addr: "tcp://localhost:5563",  key "test1"
func (sub *BMqSub) Start(addr string, key string) error {
	if err := sub._socket.Dial(addr); err != nil {
		log.Println(err)
		return err
	}

	if err := sub._socket.SetOption(zmq4.OptionSubscribe, key); err != nil {
		log.Println(err)
		return err
	}

	sub._waitGroup.Add(1)
	go sub.tick()
	return nil
}

// Stop ...
func (sub *BMqSub) Stop() {
	close(sub._exitChan)
}

func (sub *BMqSub) tick() {

	defer func() {
		recover()
		sub._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-sub._exitChan:
			return
		default:
		}

		msgMq, err := sub._socket.Recv()
		if err != nil {
			log.Println(err)
			return
		}
		sub._subChan <- msgMq

		//log.Println(string(msgMq.Frames[0]), string(msgMq.Frames[1]))
	}
}

// SubChan ...
func (sub *BMqSub) SubChan() chan zmq4.Msg {
	return sub._subChan
}
