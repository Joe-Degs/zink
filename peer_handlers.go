package zinc

func (p *Peer) initInternalHandlers() {
	ZPrintf("starting default internal request handlers...")
	p.handlers[Ping] = p.pingRequestHandler
}

func (p Peer) pingRequestHandler(packet Packet) {
	err := p.Send(&requestWrapper{
		typ:  Error,
		addr: packet.Addr(),
		data: []byte("ping handler unimplemented"),
	})
	if err != nil {
		ZErrorf("failed to respond to ping request: %s", err.Error())
	}
}
