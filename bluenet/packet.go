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
	GetLength() uint32
	GetMSGId() uint32
	GetBody() []byte
}

const packetLenghtSize = 4
const packetLengthLimit = 65535

// Protocol ...
type Protocol interface {
	ReadPacket(conn *net.TCPConn) (Packet, error)
	WritePacket(length, msgid uint32, body []byte) (Packet, error)
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

// GetMSGId ...
func (packet *BluePacket) GetMSGId() uint32 {
	return binary.BigEndian.Uint32(packet._buff[4:8])
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
		lengthBytes = make([]byte, packetLenghtSize)
		length      uint32
	)

	// read packet length
	if readLength, err := conn.Read(lengthBytes); err != nil {
		log.Println("Read length err: ", err, " read: ", readLength)
		return nil, err
	}

	// length bigendian
	if length = binary.BigEndian.Uint32(lengthBytes); length > packetLengthLimit {
		return nil, errors.New("the size of packet is larger than th limit")
	}

	buff := make([]byte, length)
	copy(buff[0:packetLenghtSize], lengthBytes)

	// read packet body
	if readLength, err := conn.Read(buff[packetLenghtSize:]); err != nil {
		log.Println("Read length err: ", err, " read: ", readLength)
		return nil, err
	}
	return NewBluePacket(buff), nil
}

// WritePacket ...
func (protocol *BlueProtocol) WritePacket(length, msgid uint32, body []byte) (Packet, error) {
	var buff = make([]byte, length)
	binary.BigEndian.PutUint32(buff[0:], length)
	binary.BigEndian.PutUint32(buff[4:], msgid)
	copy(buff[8:], body)

	return NewBluePacket(buff), nil
}
