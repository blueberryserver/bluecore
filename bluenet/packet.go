package bluenet

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
)

// Packet ...
type Packet interface {
	Serialize() []byte
}

// Protocol ...
type Protocol interface {
	ReadPacket(conn *net.TCPConn) (Packet, error)
}

// Message ...
type Message interface {
	Execute(*Session, []byte, uint32) error
}

// BluePacket ...
type BluePacket struct {
	_buff []byte
}

// Serialize ...
func (packet *BluePacket) Serialize() []byte {
	return packet._buff
}

// GetLength ...
func (packet *BluePacket) GetLength() uint32 {
	return binary.BigEndian.Uint32(packet._buff[:4])
}

// GetBody ...
func (packet *BluePacket) GetBody() []byte {
	return packet._buff[4:]
}

// NewBluePacket ...
func NewBluePacket(buff []byte) *BluePacket {
	packet := &BluePacket{}
	packet._buff = buff
	return packet
}

// BlueProtocol ...
type BlueProtocol struct {
}

// ReadPacket ...
func (protocol *BlueProtocol) ReadPacket(conn *net.TCPConn) (Packet, error) {

	var (
		lengthBytes = make([]byte, 4)
		length      uint32
	)

	if _, err := conn.Read(lengthBytes); err != nil {
		log.Println("Read length err: ", err)
		return nil, err
	}

	if length = binary.BigEndian.Uint32(lengthBytes); length > 1024 {
		return nil, errors.New("the size of packet is larger than th limit")
	}

	buff := make([]byte, length)
	copy(buff[0:4], lengthBytes)

	if _, err := conn.Read(buff[4:]); err != nil {
		return nil, err
	}
	return NewBluePacket(buff), nil
}
