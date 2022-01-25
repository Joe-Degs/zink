package zinc

func (p *Peer) initInternalHandlers() {
	p.handlers[Ping] = p.pingRequestHandler
}

func (p Peer) pingRequestHandler(packet Packet) {
	err := p.Send(&request{
		typ:  packet.Type(),
		addr: packet.Addr(),
		data: []byte("ping handler unimplemented"),
	})
	if err != nil {
		ZErrorf("failed to respond to ping request: %s", err.Error())
	}
}
