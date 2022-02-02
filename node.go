package zinc

import "time"

type NodeStatus bool

const (
	ACTIVE   NodeStatus = true
	INACTIVE NodeStatus = false
)

type connection struct {
	// if the connection is open
	active bool

	// time connection was open
	timeOpened time.Time

	// lastime connection was used
	lastUsed time.Time
}

type Node struct {
	*Peer
	Status     NodeStatus `json:"status"`
	connStatus connection
}

func NewNode(p *Peer) *Node {
	return &Node{
		Peer:   p,
		Status: INACTIVE,
	}
}

func (n *Node) Send(b []byte) error {
	return nil
}
