package zinc

import (
	"net"
)

// A Packet is any type that contains data that can be put into the wire sent
// to another peer and the peer will and can intepret the data. The Packet
// contains the state of the peer to send to, the data to send and the type
// of data that is in transit. The Packet is small and cannot be used to
// transmit large amounts of data, but can be used to initiate the process to
// send and recieve large amounts of data.
type Packet interface {
	Addr() *net.UDPAddr
	Data() []byte
	Type() PacketType
}

//go:generate stringer -type=PacketType
// PacketType represents the type of data that flows through the wire. This will
// help to decode bytes recieved through the connection
type PacketType uint8

// InternalHandlerFunc is an adapter to allow the definition of ordinary
// functions that act on requests coming into the buffer.
type InternalHandlerFunc func(Packet)

const (
	Error PacketType = iota
	Ping
	Pong
	PeerInfoRequest
	PeerInfoResponse
)

// requestWrapper implements a zinc package Packet and it represents any packet comming
// into the wire. It store state information about the data and its sender to
// aid in subsequent processing of the data.
type requestWrapper struct {
	addr *net.UDPAddr
	typ  PacketType
	data []byte
}

func (r requestWrapper) Addr() *net.UDPAddr                   { return r.addr }
func (r requestWrapper) Data() []byte                         { return r.data }
func (r requestWrapper) Type() PacketType                     { return r.typ }
func (r *requestWrapper) setRemoteEndPoint(addr *net.UDPAddr) { r.addr = addr }

// MarshalPacket returns packet as a slice of bytes that can be send over wire.
func MarshalPacket(p Packet) ([]byte, error) {
	packet := make([]byte, 1+len(p.Data()))
	packet[0] = byte(p.Type())
	copy(packet[1:], p.Data())
	return packet, nil
}

// Unmarshal turns an Packet of bytes into a Packet but without the remote
// endpoint from which the packet came from
func UnmarshalPacket(buf []byte) (Packet, error) {
	p := &requestWrapper{
		typ:  PacketType(buf[0]),
		data: buf[1:],
	}
	return p, nil
}
