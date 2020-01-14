package bluenet

import (
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// NetClient ...
type NetClient struct {
	_session        *Session
	_handler        map[uint32]interface{}
	_connected      int32
	_connectHandler interface{}
	_recvHandler    interface{}
	_exitChan       chan struct{}
	_waitGroup      *sync.WaitGroup
	_protocol       Protocol
}

// NewNetClient ...
func NewNetClient(connHandler interface{}, recvHandler interface{}, protocol Protocol) *NetClient {
	return &NetClient{
		_handler:        make(map[uint32]interface{}),
		_connectHandler: connHandler,
		_recvHandler:    recvHandler,
		_exitChan:       make(chan struct{}), // exit channel
		_waitGroup:      &sync.WaitGroup{},   // goroutine wait greoup
		_protocol:       protocol,
	}
}

// AddHandler ...
func (client *NetClient) AddHandler(msgID uint32, handler interface{}) {
	client._handler[msgID] = handler
}

// Connect ...
func (client *NetClient) Connect(n string, addr string, timeout time.Duration) error {
	resolver, err := net.ResolveTCPAddr("tcp4", addr)
	conn, err := net.DialTCP(n, nil, resolver)
	if err != nil {
		log.Println(err)
		return err
	}

	client._session = NewSession(0, conn)
	client.handlerConnect(client._session)
	return nil
}

// IsConnected ...
func (client *NetClient) IsConnected() bool {
	if client._session == nil || atomic.LoadInt32(&client._connected) == 0 {
		return false
	}
	return true
}

func (client *NetClient) handlerConnect(session *Session) {
	log.Println("Connection server")

	client._waitGroup.Add(1)
	go client.handlerRecv(session)

	client._waitGroup.Add(1)
	go client.handlerSend(session)

	client._waitGroup.Add(1)
	go client.handlerLoop(session)
}

func (client *NetClient) handlerRecv(session *Session) {
	//log.Println("Recv")
	defer func() {
		//log.Println("def Recv")
		recover()
		session.Close()
		client._waitGroup.Done()
	}()

	for {

		select {
		case <-client._exitChan:
			return

		case <-session._closeChan:
			return
		default:
		}
		packet, err := client._protocol.ReadPacket(session._conn)
		if err != nil {
			t := reflect.ValueOf(client).Type()
			log.Println(t, "Recv err: ", err)
			return
		}

		session._recvPacketChan <- packet
	}
}

func (client *NetClient) handlerSend(session *Session) {
	//log.Println("Send")
	defer func() {
		//log.Println("def Send")
		recover()
		session.Close()
		client._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-client._exitChan:
			return

		case <-session._closeChan:
			return

		case packet := <-session._sendPacketChan:
			if session.IsClosed() {
				return
			}
			if _, err := session.Write(packet.Serialize()); err != nil {
				t := reflect.ValueOf(client).Type()
				log.Println(t, "Send err: ", err)
				return
			}
		}
	}
}

func (client *NetClient) handlerLoop(session *Session) {
	//log.Println("Loop")
	defer func() {
		//log.Println("def Loop")
		recover()
		session.Close()
		client._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-client._exitChan:
			return
		case <-session._closeChan:
			return

		case packet := <-session._recvPacketChan:
			if session.IsClosed() {
				log.Println("Loop closed socket")
				return
			}

			// excute message
			client.excutePacket(session, packet)
		}
	}
}

func (client *NetClient) excutePacket(session *Session, packet Packet) {

	length := packet.GetLength()
	msgID := packet.GetMSGId()
	body := packet.GetBody()

	log.Println("execute packet length: ", length)
	log.Println("execute packet msgid: ", msgID)

	handler := client._handler[msgID].(Message)
	if handler == nil {
		t := reflect.ValueOf(client).Type()
		log.Println(t, " handler not found")
		return
	}

	handler.Execute(session, body, length-4)
}

// SendMsg ...
func (client *NetClient) SendMsg(msgID uint32, body []byte, length uint32) (err error) {
	packet, err := client._protocol.WritePacket(length, msgID, body)
	if err != nil {
		t := reflect.ValueOf(client).Type()
		log.Println(t, "Recv err: ", err)
		return
	}

	return client._session.SendPacket(packet)
}

// Close ...
func (client *NetClient) Close() {
	close(client._exitChan)
	client._waitGroup.Wait()
}
