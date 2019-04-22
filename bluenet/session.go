package bluenet

import (
	"encoding/binary"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// Session is network session unit
type Session struct {
	_id             int32
	_conn           *net.TCPConn
	_closed         int32
	_closeOnce      sync.Once
	_closeChan      chan struct{}
	_sendPacketChan chan Packet
	_recvPacketChan chan Packet
}

// NewSession ...
func NewSession(id int32, conn *net.TCPConn) *Session {
	return &Session{
		_closed:         0,
		_id:             id,
		_conn:           conn,
		_closeChan:      make(chan struct{}),
		_sendPacketChan: make(chan Packet, 10),
		_recvPacketChan: make(chan Packet, 10),
	}
}

// GetID ..
func (session *Session) GetID() int32 {
	return session._id
}

// Read buffer
func (session *Session) Read(data []byte) (int, error) {
	return session._conn.Read(data)
}

// Read buffer
func (session *Session) Write(data []byte) (int, error) {
	return session._conn.Write(data)
}

// Close socket
func (session *Session) Close() {
	session._closeOnce.Do(func() {
		atomic.StoreInt32(&session._closed, 1)
		close(session._closeChan)
		close(session._sendPacketChan)
		close(session._recvPacketChan)
		session._conn.Close()
	})
}

// IsClosed ..
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session._closed) == 1
}

// sendPacket ...
func (session *Session) sendPacket(packet Packet) (err error) {

	if session.IsClosed() {
		return errors.New("use of closed network connection")
	}

	defer func() {
		if e := recover(); e != nil {
			err = errors.New("use of closed network connection")
		}
	}()

	select {
	case session._sendPacketChan <- packet:
		return nil

	case <-session._closeChan:
		return errors.New("use of closed network connection")

	default:
		return errors.New("write packet was blocking")
	}
}

// SendMsg ...
func (session *Session) SendMsg(msgID uint32, body []byte, length uint32) (err error) {
	length += 4 + 4
	var buff = make([]byte, length)

	binary.BigEndian.PutUint32(buff[0:], length)
	binary.BigEndian.PutUint32(buff[4:], msgID)
	copy(buff[8:], body)

	packet := NewBluePacket(buff)
	return session.sendPacket(packet)
}
