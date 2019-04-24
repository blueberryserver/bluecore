package bluemq

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-zeromq/zmq4"
)

// BMqPub ...
type BMqPub struct {
	_socket    zmq4.Socket
	_exitChan  chan struct{}
	_waitGroup *sync.WaitGroup
	_jobChan   chan interface{}
}

// NewPub ...
func NewPub(queueSize int) *BMqPub {
	return &BMqPub{
		_socket:    zmq4.NewPub(context.Background()),
		_exitChan:  make(chan struct{}),               // exit channel
		_waitGroup: &sync.WaitGroup{},                 // goroutine wait greoup
		_jobChan:   make(chan interface{}, queueSize), // job queue channel
	}
}

// Close ...
func (pub *BMqPub) Close() {
	pub._socket.Close()
}

// Start ...( "tcp://*5563")
func (pub *BMqPub) Start(addr string) error {

	for {
		if err := pub._socket.Listen(addr); err == nil {
			break
		}

		select {
		case <-time.After(10 * time.Second):
			return errors.New("Socket Listen Timeout")
		}
	}

	pub._waitGroup.Add(1)
	go pub.tick()
	return nil
}

// Stop ..
func (pub *BMqPub) Stop() {
	close(pub._exitChan)
	pub._waitGroup.Wait()
}

// Send ...
func (pub *BMqPub) Send(key, msg string) {
	job := func() error {
		log.Println("Send ", key, "-> ", msg)

		msgMq := zmq4.NewMsgFrom(
			[]byte(key),
			[]byte(msg),
		)
		if err := pub._socket.Send(msgMq); err != nil {
			//log.Println(err)
			return err
		}
		return nil
	}
	pub._jobChan <- job
}

func (pub *BMqPub) tick() {

	log.Println("zmq pub start tick")

	defer func() {
		recover()
		pub._waitGroup.Done()
		log.Println("zmq pub tick defer")
	}()

	for {
		// check exit channel
		select {
		case <-pub._exitChan:
			return
		case job := <-pub._jobChan:
			err := job.(func() error)()
			if err != nil {
				log.Println(err)
			}
		default:
		}
	}
}
