package zinc

import (
	"fmt"
	"net"
	"os"
)

func (p *Peer) initInternalHandlers() {
	ZPrintf("starting default internal request handlers...")
	p.handlers[Ping] = p.pingRequestHandler
}

func makeResponsePacket(typ PacketType, data []byte, addr *net.UDPAddr) Packet {
	return &requestWrapper{typ: typ, data: data, addr: addr}
}

// handle ping requests sent to peer
func (p Peer) pingRequestHandler(packet Packet) {
	// err := p.SendToAddr(UnImplementedEndPoint, packet.Addr())
	data, err := p.MarshalJSON()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	resp := makeResponsePacket(PeerInfo, data, packet.Addr())

	err = p.Send(resp)
	if err != nil {
		ZErrorf("failed to respond to ping request: %s", err.Error())
	}
}
