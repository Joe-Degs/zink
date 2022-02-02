package zinc

import (
	"context"
	"io"

	"github.com/Joe-Degs/zinc/internal/config"
	"github.com/google/uuid"
)

type Cluster struct {
	*Peer
	Members map[string]*Node
}

func NewCluster(config *config.Config) (*Cluster, error) {
	cluster := &Cluster{
		Peer: &Peer{
			Name: config.Name,
		},
		Members: make(map[string]*Node),
	}

	if config.Id != "" {
		id, err := uuid.Parse(config.Id)
		if err != nil {
			return cluster, err
		}
		cluster.Id = id
	} else {
		cluster.Id = uuid.New()
	}

	if err := cluster.Peer.setAddr(config.Addr); err != nil {
		return cluster, err
	}

	for _, peer := range config.Peers {
		var id uuid.UUID
		if peer.Id != "" {
			var err error
			id, err = uuid.Parse(peer.Id)
			if err != nil {
				return cluster, err
			}
		} else {
			id = uuid.New()
		}
		p, err := PeerFromSpec(peer.Name, peer.Addr, id)
		if err != nil {
			continue
		}

		cluster.Members[id.String()] = NewNode(p)
	}

	return cluster, nil
}

func (c *Cluster) StartServer(cl chan<- io.Closer) (context.CancelFunc, error) {
	c.initInternalHandlers()
	return c.Peer.StartServer(cl)
}

func (c *Cluster) FindById(id string) *Node {
	return nil
}

func (c *Cluster) Broadcast(b []byte) error {
	return nil
}
