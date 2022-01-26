package netutil

import (
	"errors"
	"fmt"
	"net"
	"time"

	"inet.af/netaddr"
)

func resolveAddr(addr string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", addr)
}

// TryConnect tries to make to UDP connection to addr regularly until `wait`
// time is over
func TryConnect(addr string, wait time.Duration) (*net.UDPConn, error) {
	done := time.Now().Add(wait)
	for time.Now().Before(done) {
		c, err := Connect(addr)
		if err == nil {
			return c, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("%v unreachable for %v", addr, wait)
}

// Connect connects to a UDP listener on address `addr`
func Connect(addr string) (*net.UDPConn, error) {
	raddr, err := resolveAddr(addr)
	if err != nil {
		return nil, err
	}
	c, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Listen a listening UDP connection
func Listen(addr string) (*net.UDPConn, error) {
	raddr, err := resolveAddr(addr)
	if err != nil {
		return nil, err
	}
	lstn, err := net.ListenUDP("udp", raddr)
	if err != nil {
		return nil, err
	}
	return lstn, nil
}

// Return a UDP connection listening on a random port
func ListenOnLocalRandomPort() (*net.UDPConn, error) {
	return Listen("localhost:0")
}

// IPPortFromAddr tries to return a valid netaddr.IPPort from an address string
func IPPortFromAddr(addr string) (*netaddr.IPPort, error) {
	ipport, err := netaddr.ParseIPPort(addr)
	if err != nil {
		return nil, err
	}
	if !ipport.IsValid() {
		return nil, errors.New("invalid ipport")
	}
	return &ipport, nil
}
