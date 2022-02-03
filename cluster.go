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

func NewCluster(config *config.ClusterConfig) (*Cluster, error) {
	cluster := &Cluster{
		Members: make(map[string]*Node),
	}

	if config == nil {
		if err := cluster.initDefault(); err != nil {
			return cluster, err
		}
		return cluster, nil
	}

	if err := cluster.init(config); err != nil {
		return cluster, nil
	}

	return cluster, nil
}

// initialize the cluster with the values from the loaded config file
func (c *Cluster) init(config *config.ClusterConfig) error {
	if config.Id != "" {
		id, err := uuid.Parse(config.Id)
		if err != nil {
			return err
		}
		c.Id = id
	} else {
		c.Id = uuid.New()
	}

	// add address and open connection
	if err := c.Peer.setAddr(config.Addr); err != nil {
		return err
	}

	for _, peer := range config.Peers {
		var id uuid.UUID
		if peer.Id != "" {
			var err error
			id, err = uuid.Parse(peer.Id)
			if err != nil {
				return err
			}
		} else {
			id = uuid.New()
		}
		p, err := PeerFromSpec(peer.Name, peer.Addr, id)
		if err != nil {
			//TODO(joe):
			// what to do with the error?
			continue
		}

		c.Members[id.String()] = NewNode(p)
	}
	return nil
}

// use the default config files that come with the project
func (c *Cluster) initDefault() error {
	return nil
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
