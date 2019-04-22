package bluenet

import (
	"encoding/binary"
	"log"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// init function
func init() {

}

// session map
type _SessionMap map[int32]*Session

// msg handler interface
type _MsgHandler interface {
	execute(*Session, []byte, uint32) error
}

// msg handler
type _MsgHandlerMap map[uint32]_MsgHandler

// NetServer network server
type NetServer struct {
	_net            string
	_addr           string
	_genID          int32
	_sessions       _SessionMap
	_handler        map[uint32]interface{}
	_lock           *sync.Mutex
	_connectHandler interface{}
	_recvHandler    interface{}
	_exitChan       chan struct{}
	_waitGroup      *sync.WaitGroup
	_protocol       Protocol
}

// NewNetServer create new server
func NewNetServer(net string, addr string, connHandler interface{}, recvHandler interface{}, protocol Protocol) *NetServer {
	return &NetServer{
		_net:            net,
		_addr:           addr,
		_genID:          0,
		_sessions:       make(_SessionMap),
		_handler:        make(map[uint32]interface{}),
		_lock:           &sync.Mutex{},
		_connectHandler: connHandler,         // connection callback
		_recvHandler:    recvHandler,         // recv bytes callback
		_exitChan:       make(chan struct{}), // exit channel
		_waitGroup:      &sync.WaitGroup{},   // goroutine wait greoup
		_protocol:       protocol,
	}
}

func (server *NetServer) addSession(session *Session) {
	server._lock.Lock()
	defer server._lock.Unlock()
	server._sessions[session._id] = session
}

// RemoveSession remove session
func (server *NetServer) removeSession(session *Session) {
	server._lock.Lock()
	defer server._lock.Unlock()
	server._sessions[session._id] = nil
}

// AddHandler ...
func (server *NetServer) AddHandler(msgID uint32, handler interface{}) {
	server._handler[msgID] = handler
}

// Start server tcp listen
func (server *NetServer) Start(timeout time.Duration) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", server._addr)
	if err != nil {
		return err
	}

	ln, err := net.ListenTCP(server._net, tcpAddr)
	if err != nil {
		return err
	}

	// goroutine add
	server._waitGroup.Add(1)

	// defer closing
	defer func() {
		ln.Close()
		server._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-server._exitChan:
			return nil
		default:
		}

		// setting time out
		ln.SetDeadline(time.Now().Add(timeout))

		conn, err := ln.AcceptTCP()
		if err != nil {
			continue
		}

		atomic.StoreInt32(&server._genID, server._genID+1)
		session := NewSession(server._genID, conn)
		server.addSession(session)

		server.handlerConnect(session)
	}
}

// Stop server
func (server *NetServer) Stop() {
	close(server._exitChan)
	server._waitGroup.Wait()
}

func (server *NetServer) handlerConnect(session *Session) {
	log.Println("connection session id: ", session._id)

	server._waitGroup.Add(1)
	go server.handlerRecv(session)

	server._waitGroup.Add(1)
	go server.handlerLoop(session)

	server._waitGroup.Add(1)
	go server.handlerSend(session)
}

func (server *NetServer) handlerRecv(session *Session) {
	//log.Println("Recv id: ", session._id)
	defer func() {
		//log.Println("def Recv")
		recover()
		session.Close()
		server.removeSession(session)
		server._waitGroup.Done()
	}()

	for {

		select {
		case <-server._exitChan:
			log.Println("Recv exit channel id: ", session._id)
			return
		case <-session._closeChan:
			log.Println("Recv close channel id: ", session._id)
			return
		default:
		}

		packet, err := server._protocol.ReadPacket(session._conn)
		if err != nil {
			t := reflect.ValueOf(server).Type()
			log.Println(t, " err: ", err)
			return
		}

		//log.Println("Recv chan <- packet")
		session._recvPacketChan <- packet
	}
}

func (server *NetServer) handlerSend(session *Session) {
	//log.Println("Send id: ", session._id)
	defer func() {
		//log.Println("def Send")
		recover()
		session.Close()
		server.removeSession(session)
		server._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-server._exitChan:
			log.Println("Send exit channel id: ", session._id)
			return
		case <-session._closeChan:
			log.Println("Send close channel id: ", session._id)
			return

		case packet := <-session._sendPacketChan:
			if session.IsClosed() {
				log.Println("Closed session id: ", session._id)
				return
			}

			if _, err := session.Write(packet.Serialize()); err != nil {
				t := reflect.ValueOf(server).Type()
				log.Println(t, " err: ", err)
				return
			}
		}
	}
}

func (server *NetServer) handlerLoop(session *Session) {
	//log.Println("Loop id: ", session._id)
	defer func() {
		//log.Println("def Loop")
		recover()
		session.Close()
		server.removeSession(session)
		server._waitGroup.Done()
	}()

	for {
		// check exit channel
		select {
		case <-server._exitChan:
			log.Println("Loop exit channel id: ", session._id)
			return
		case <-session._closeChan:
			log.Println("Loop close channel id: ", session._id)
			return

		case packet := <-session._recvPacketChan:
			if session.IsClosed() {
				log.Println("Loop closed sicket id: ", session._id)
				return
			}

			// excute message
			server.excutePacket(session, packet)
		}
	}
}

func (server *NetServer) excutePacket(session *Session, packet Packet) {
	buff := packet.Serialize()

	length := binary.BigEndian.Uint32(buff[:4])
	msgID := binary.BigEndian.Uint32(buff[4:8])
	body := buff[8:]

	handler := server._handler[msgID].(Message)
	if handler == nil {
		t := reflect.ValueOf(server).Type()
		log.Println(t, " handler not found")
		return
	}

	handler.Execute(session, body, length-8)
}
