package zinc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Joe-Degs/zinc/internal/netutil"
	"github.com/Joe-Degs/zinc/internal/pool"
	"inet.af/netaddr"
)

// A Peer is a node in the system, it can interact with other peers and send
// things back and forth
type Peer struct {
	Id        Uid             `json:"-"`
	Name      string          `json:"name,omitempty"`
	LocalAddr *netaddr.IPPort `json:"-"`
	lstn      *net.UDPConn
	recv      chan Packet
	handlers  map[PacketType]InternalHandlerFunc
}

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

	var err error
	if peer.LocalAddr, err = netutil.IPPortFromAddr(addr); err != nil {
		return peer, err
	}

	if err = peer.setPeerListener(); err != nil {
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
	if p.lstn, err = netutil.ListenOnLocalRandomPort(); err != nil {
		ZErrorf("%v", err)
		return p
	}
	if p.LocalAddr, err = netutil.IPPortFromAddr(p.lstn.LocalAddr().String()); err != nil {
		return p
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
		if p.LocalAddr, err = netutil.IPPortFromAddr(pi.Addr); err != nil {
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
			return p.setIPPort(str[1])
		} else {
			p.Name = str[1]
			if len(str) == 3 {
				return p.setIPPort(str[2])
			}
		}
	}
	return nil
}

func (p *Peer) setIPPort(addr string) (err error) {
	if p.LocalAddr, err = netutil.IPPortFromAddr(addr); err != nil {
		return err
	}
	return
}

// setPeerListener opens a new listening socket on local interface with
// address Peer.LocalAddr. This function will create a listener listening on all
// the local interfaces if Peer.LocalAddr is nil
func (p *Peer) setPeerListener() error {
	var err error
	if p.lstn, err = netutil.Listen(p.LocalAddr.String()); err != nil {
		return err
	}
	return err
}

// Send transmits a packet containing the remote address it is being sent to.
func (p Peer) Send(packet Packet) error { return p.SendToAddr(packet, nil) }

// SendToAddr sends packet to their remote endpoints. If the packet already
// contains its remote endpoint it just sends it, else it unmarshals the
// packet and sends it to its remote address
func (p Peer) SendToAddr(packet Packet, addr *net.UDPAddr) error {
	write := func(b []byte, addr *net.UDPAddr) error {
		if n, err := p.lstn.WriteToUDP(b, addr); err != nil {
			return fmt.Errorf("could not send packet: %w", err)
		} else if n < len(packet.Data()) {
			return fmt.Errorf("could not send all data, got: %d, sent: %d", len(packet.Data()), n)
		}
		return nil
	}
	b, err := MarshalPacket(packet)
	if err != nil {
		return fmt.Errorf("Send(Packet): %w", err)
	}
	if addr != nil {
		return write(b, addr)
	} else if packet.Addr() != nil {
		return write(b, packet.Addr())
	}
	return fmt.Errorf("specify remote endpoint to send packet")
}

// StartRequestReciever starts the goroutines for recieving new packets and
// determining what to do with the packets.
func (p *Peer) StartRequestReciever(cl chan<- io.Closer) (context.CancelFunc, error) {
	if p.lstn == nil {
		if !p.LocalAddr.IsValid() {
			return nil, fmt.Errorf("StartRequestReciever: %s", p.LocalAddr.String())
		}
		if err := p.setPeerListener(); err != nil {
			return nil, err
		}
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		<-sig
		cl <- p.lstn
	}()

	p.initInternalHandlers()
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan Packet)
	go p.processRequests(ctx, ch)
	go func(ctx context.Context) {
		ZPrintf("%s listening on %s", p.Id, p.LocalAddr.String())
		for {
			select {
			case <-ctx.Done():
				return
			default:
				buffer := pool.GetBufferSized(500)
				buf := buffer.Bytes()
				n, raddr, err := p.lstn.ReadFromUDP(buf)
				if err != nil {
					ZErrorf("StartRequestReceiver: %v", err)
					continue
				}
				ch <- requestWrapper{
					addr: raddr,
					typ:  PacketType(buf[0]),
					data: append(make([]byte, 0, n), buf[:n]...),
				}
				pool.PutBuffer(buffer)
			}
		}
	}(ctx)

	return cancel, nil
}

// processRequests is run as a goroutine to process a newly recieved packet
// and determine where the packet is destined for.
func (p *Peer) processRequests(ctx context.Context, ch <-chan Packet) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ch:
			ZPrintf("Recieved '%s' request from %s", req.Type().String(), req.Addr().String())

			if f, ok := p.handlers[req.Type()]; ok {
				go f(req)
			} else {
				ZErrorf("no registered handler for packet type %s", req.Type().String())
				go func() {
					err := p.Send(ErrrorWithAddr(UnknownPacketType, req.Addr()))
					if err != nil {
						ZErrorf("sending error response failed: %s", err.Error())
					}
				}()
			}
		}
	}
}
