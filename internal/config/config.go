package config

import (
	"encoding/json"
	"net"
	"os"

	"github.com/Joe-Degs/zinc/internal/netutil"
	"inet.af/netaddr"
)

var (
	sampleClusterConfig = `{
	"name": "joe",
	"addr": "localhost:6009",
	"peers": [{
		"name": "kofi",
		"addr": "localhost:7000",
	},{
		"name": "messi",
		"addr": "localhost:30011",
	}, {
		"name": "oskee",
		"addr": "localhost:40011"
	}]
}`

	samplePeerConfig = `{
	"name": "oskee",
	"addr": "localhost:40011"
}`
)

type PeerConfig struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
	Id   string `json:"id"`
}

type ClusterConfig struct {
	*PeerConfig
	Peers []*PeerConfig
}

func (c PeerConfig) GetConnAndIP() (conn *net.UDPConn, addr *netaddr.IPPort, err error) {
	if c.Addr == "" {
		return netutil.ConnAndAddr("localhost:0")
	}
	return netutil.ConnAndAddr(c.Addr)
}

func (c *ClusterConfig) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *ClusterConfig) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, c); err != nil {
		return err
	}
	return nil
}

func (p *PeerConfig) MarshalJSON() ([]byte, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (p *PeerConfig) UnmarshalJSON(b []byte) error {
	if err := json.Unmarshal(b, p); err != nil {
		return err
	}
	return nil
}

func ClusterConfigFromJSON(b []byte) (*ClusterConfig, error) {
	c := &ClusterConfig{}
	if err := c.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return c, nil
}

func PeerConfigFromJSON(b []byte) (*PeerConfig, error) {
	p := &PeerConfig{}
	if err := p.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return p, nil
}

func ClusterConfigFromFile(filename string) (*ClusterConfig, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ClusterConfigFromJSON(file)
}

func DefaultClusterConfig() (*ClusterConfig, error) {
	return ClusterConfigFromJSON([]byte(sampleClusterConfig))
}

func DefaultPeerConfig() (*PeerConfig, error) {
	return PeerConfigFromJSON([]byte(samplePeerConfig))
}
