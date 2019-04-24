package bluemq

import (
	"context"
	"log"
	"sync"

	"github.com/go-zeromq/zmq4"
)

// BMqReq ...
type BMqReq struct {
	_socket    zmq4.Socket
	_exitChan  chan struct{}
	_waitGroup *sync.WaitGroup
	_reqChan   chan zmq4.Msg
}

// NewReq ...
func NewReq(queueSize int) *BMqReq {
	return &BMqReq{
		_socket:    zmq4.NewReq(context.Background()),
		_exitChan:  make(chan struct{}),            // exit channel
		_waitGroup: &sync.WaitGroup{},              // goroutine wait greoup
		_reqChan:   make(chan zmq4.Msg, queueSize), // subscribed messagegs
	}
}

// Close ...
func (req *BMqReq) Close() {
	req._socket.Close()
}

// Stop ..
func (req *BMqReq) Stop() {
	close(req._exitChan)
	req._waitGroup.Wait()
}

// Start ...( "tcp://localhost:5559")
func (req *BMqReq) Start(addr string) {
	if err := req._socket.Dial(addr); err != nil {
		log.Println(err)
		return
	}
	req._waitGroup.Add(1)
	go req.tick()
}

// Send ...
func (req *BMqReq) Send(msg string) error {

	msgMq := zmq4.NewMsgString(msg)
	if err := req._socket.Send(msgMq); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (req *BMqReq) tick() {

	log.Println("zmq req start tick")

	defer func() {
		recover()
		req._waitGroup.Done()
		log.Println("zmq req tick defer")
	}()

	for {
		// check exit channel
		select {
		case <-req._exitChan:
			return
		default:
		}

		msg, err := req._socket.Recv()
		if err != nil {
			log.Println(err)
			continue
		}

		req._reqChan <- msg

		//log.Println("recv->", string(msg.Frames[0]))
	}
}

// ReqChan ...
func (req *BMqReq) ReqChan() chan zmq4.Msg {
	return req._reqChan
}
