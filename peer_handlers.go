package zinc

func (p *Peer) initInternalHandlers() {
	ZPrintf("starting default internal request handlers...")
	p.handlers[Ping] = p.pingRequestHandler
}

func (p Peer) pingRequestHandler(packet Packet) {
	err := p.SendToAddr(UnImplementedEndPoint, packet.Addr())
	if err != nil {
		ZErrorf("failed to respond to ping request: %s", err.Error())
	}
}
