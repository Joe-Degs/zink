package zinc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"inet.af/netaddr"
)

// A Peer is a node in the system, it can interact with other peers and send
// things back and forth
type Peer struct {
	Id        Uid            `json:"-"`
	Name      string         `json:"name,omitempty"`
	LocalAddr netaddr.IPPort `json:"-"`
	lstn      *net.UDPConn
	recv      chan Packet
	handlers  map[PacketType]InternalHandlerFunc
}

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

// Returns a peer with a random state, mostly good for testing
func RandomPeer(name string) *Peer {
	return peer(name)
}

// PeerFromSpec returns a peer with the desired state passed to the function
func PeerFromSpec(name string, addr string, uidfunc func() Uid) (*Peer, error) {
	peer := &Peer{
		Name:     name,
		Id:       uidfunc(),
		recv:     make(chan Packet),
		handlers: make(map[PacketType]InternalHandlerFunc),
	}
	peer.initInternalHandlers()
	var err error
	if peer.LocalAddr, err = netaddr.ParseIPPort(addr); err != nil {
		return peer, err
	}

	if err = peer.getListenerFromAddr(); err != nil {
		return peer, err
	}
	return peer, nil
}

// peer returns a new peer with mostly random information. this function is
// useful for generating peers for testing purposes. Peers to be used to
// transmit data must use generate peers with more specific data with the
// `PeerFromSpec` function.
func peer(name string) (p *Peer) {
	p = &Peer{
		Id:   RandomUid(),
		Name: name,
	}

	var err error

	// peer with random info, so listen on all the local interfaces except
	// multicast. If this function is not succesful in creating the udp listener
	// Peer.LocalAddr will still be nil
	if p.lstn, err = net.ListenUDP("udp", nil); err != nil {
		ZErrorf("%v", err)
		return p
	}

	addr := p.lstn.LocalAddr()
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		p.LocalAddr, ok = netaddr.FromStdAddr(udpAddr.IP, udpAddr.Port, udpAddr.Zone)
		if ok {
			return p
		}
	} else {
		if p.LocalAddr, err = netaddr.ParseIPPort(addr.String()); err != nil {
			ZErrorf("%v", err)
			return p
		}
	}
	return
}

// Custom json marshaller for the peer type
func (p *Peer) MarshalJSON() ([]byte, error) {
	type PeerInfo Peer
	return json.Marshal(&struct {
		Id   string `json:"id"`
		Addr string `json:"addr,omitempty"`
		*PeerInfo
	}{
		Id:       p.Id.String(),
		Addr:     p.LocalAddr.String(),
		PeerInfo: (*PeerInfo)(p),
	})
}

// Custom json unmarshaller for the peer type
func (p *Peer) UnmarshalJSON(data []byte) error {
	type PeerInfo Peer
	pi := struct {
		Id   string `json:"id"`
		Addr string `json:"addr,omitempty"`
		*PeerInfo
	}{
		PeerInfo: (*PeerInfo)(p),
	}

	var err error
	if err = json.Unmarshal(data, &pi); err != nil {
		return err
	}

	// parse uid
	if p.Id, err = ParseUid(pi.Id); err != nil {
		return err
	}

	// when the Peer.LocalAddr is left empty, it becomes "invalid IPPort" after
	// the unmarshaling. So we check that it is actually an ipport before we
	// try to parse into an ipport. Might want to change it later to something
	// a little more robust
	if pi.Addr != "invalid IPPort" {
		if p.LocalAddr, err = netaddr.ParseIPPort(pi.Addr); err != nil {
			return err
		}
	}
	return nil
}

func (p Peer) String() string {
	str := &strings.Builder{}
	str.WriteString(p.Id.String())
	if p.Name != "" {
		str.WriteString(" " + p.Name)
	}
	if p.LocalAddr.IsValid() {
		str.WriteString(" " + p.LocalAddr.String())
	}
	return str.String()
}

// MarshalText implements the encoding.TextMarshaler inteface. It returns the
// info of a peer in text format, which is useful when you want to write it to
// a file for storage The text format of a peer is just
// 3 space separated strings. In the case that the peer's ip address is `nil`,
// only the name and the uid of the peer will be marshalled to text
func (p *Peer) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements the encoding.TextUmarshaler interface.
func (p *Peer) UnmarshalText(text []byte) error {
	// text contains three space separated strings marshalled to bytes.
	// The strings represent the `id`, `name` and `addr` of a peer.
	// the id field is compulsory, name and addr are optional.
	str := strings.Split(string(text), " ")
	var err error
	if p.Id, err = ParseUid(str[0]); err != nil {
		return err
	}

	// when the name field of a peer is empty, and the address is present, the
	// address field becomes the second field in the text string
	if len(str) > 1 {
		if strings.Contains(str[1], ":") {
			return p.parseIPPort(str[1])
		} else {
			p.Name = str[1]
			if len(str) == 3 {
				return p.parseIPPort(str[2])
			}
		}
	}
	return nil
}

func (p *Peer) parseIPPort(addr string) (err error) {
	if p.LocalAddr, err = netaddr.ParseIPPort(addr); err != nil {
		return err
	}
	return
}

// getListenerFromAddr opens a new listening socket on local interface with
// address Peer.LocalAddr. This function will create a listener listening on all
// the local interfaces if Peer.LocalAddr is nil
func (p *Peer) getListenerFromAddr() error {
	udpAddr := p.LocalAddr.UDPAddr()
	var err error
	if p.lstn, err = net.ListenUDP(udpAddr.Network(), udpAddr); err != nil {
		return err
	}
	return err
}

// Listener returns the udp listener of the peer. This function should be called
// only locally, if you want a remote connection to a peer, you have to call its
// `Connect` method
func (p *Peer) Listener() (*net.UDPConn, error) {
	// if there is a local address without a listener, start a new listener that
	// can be used later
	if p.LocalAddr.IsValid() && p.lstn == nil {
		if err := p.getListenerFromAddr(); err != nil {
			return nil, err
		}
		return p.lstn, nil
	}
	return p.lstn, nil
}

// request implements a zinc package Packet and it represents any packet comming
// into the wire. It store state information about the data and its sender to
// aid in subsequent processing of the data.
type request struct {
	addr *net.UDPAddr
	typ  PacketType
	data []byte
}

func (r request) Addr() *net.UDPAddr                   { return r.addr }
func (r request) Data() []byte                         { return r.data }
func (r request) Type() PacketType                     { return r.typ }
func (r *request) setRemoteEndPoint(addr *net.UDPAddr) { r.addr = addr }

// MarshalPacket returns packet as a slice of bytes that can be send over wire.
func MarshalPacket(p Packet) ([]byte, error) {
	packet := make([]byte, 1+len(p.Data()))
	packet[0] = byte(p.Type())
	if n := copy(packet[1:], p.Data()); n != len(p.Data()) {
		return nil, errors.New("marshalpacket: incomplete copy of packet data")
	}
	return packet, nil
}

// Unmarshal turns an Packet of bytes into a Packet but without the remote
// endpoint from which the packet came from
func UnmarshalPacket(buf []byte) (Packet, error) {
	p := &request{
		typ:  PacketType(buf[0]),
		data: buf[1:],
	}
	return p, nil
}

// Send marshals packet to byte slice and sends it to a remote endpoint
func (p Peer) Send(packet Packet) error {
	byts, err := MarshalPacket(packet)
	if err != nil {
		return fmt.Errorf("Send(Packet): %w", err)
	}
	if n, err := p.lstn.WriteToUDP(byts, packet.Addr()); err != nil {
		return fmt.Errorf("could not send packet: %w", err)
	} else if n != len(packet.Data()) {
		return fmt.Errorf("could not send full packet to remote endpoint")
	}
	return nil
}

// Send recieves a packet, marshals the packet to bytes and sends it to the
// specified peer through its connection.
func (p Peer) SendPacket(packet Packet, addr *net.UDPAddr) error {
	// if packet is of the `request` type, change the remote endpoint
	// of the Packet and then send it
	if pack, ok := packet.(*request); ok {
		pack.setRemoteEndPoint(addr)
		return p.Send(pack)
	}

	// concrete type of packet is not a pointer to request.
	byts, err := MarshalPacket(packet)
	if err != nil {
		return fmt.Errorf("Send(Packet): %w", err)
	}
	if n, err := p.lstn.WriteToUDP(byts, packet.Addr()); err != nil {
		return fmt.Errorf("could not send packet: %w", err)
	} else if n != len(packet.Data()) {
		return fmt.Errorf("could not send full packet to remote endpoint")
	}
	return nil
}

// StartRequestReciever just listens on p.LocalAddr for data from any peer.
// If the data is a request type it shoots it off to the appropriate handler
// for that request to be handled.
func (p *Peer) StartRequestReciever() error {
	// the first thing is to start listening on some port and address
	if p.lstn == nil {
		if !p.LocalAddr.IsValid() {
			return errors.New(p.LocalAddr.String())
		}
		if err := p.getListenerFromAddr(); err != nil {
			return err
		}
	}

	// channels for recieving errors and os signals through
	signals := make(chan os.Signal, 1)
	cherr := make(chan error)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// we start a goroutine whose job is to recieve arbitrary data from the
	// connection and send the data through a channel to another goroutine
	// which will make use of the data.
	go p.processRequests()
	go func() {
		ZPrintf("%s listening on %s", p.Id, p.LocalAddr.String())
		for {
			buf := make([]byte, 512)
			n, raddr, err := p.lstn.ReadFromUDP(buf)
			if err != nil {
				cherr <- err
			}
			ZPrintf("recieved %d bytes from %s", n, raddr.String())
			p.recv <- request{
				addr: raddr,
				typ:  PacketType(buf[0]),
				data: buf[1:n],
			}
		}
	}()

	for {
		select {
		case err := <-cherr:
			ZErrorf("%v", err)
		case sig := <-signals:
			ZPrintf("%s", sig)
			if err := p.lstn.Close(); err != nil {
				ZErrorf("could not close peer: %v", err)
				os.Exit(1)
			}
			os.Exit(1)
		}
	}
	return nil
}

func (p *Peer) processRequests() {
	zlog.Println("starting initial request processor...")
	for req := range p.recv {
		ZPrintf("Recieved %s request from %s", req.Type().String(), req.Addr().String())
		ZPrintf("data recieved: %s", req.Data())

		if f, ok := p.handlers[PacketType(req.Type())]; ok {
			go f(req)
		} else {
			// do not have a handler for the request so send it
			ZErrorf("no registered handler for packet type %s", req.Type().String())
			go func() {
				err := p.Send(&request{
					addr: req.Addr(),
					typ:  Error,
					data: []byte("server cannot handle request type"),
				})
				if err != nil {
					ZErrorf("sending error response failed: %s", err.Error())
				}
			}()
		}
	}
}
