package peer

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Joe-Degs/zinc/cmd/zinkctl/opts"
	"github.com/jessevdk/go-flags"
)

type Peer struct{}

func (Peer) Help() string {
	return strings.TrimSpace(`
Usage: zinkctl [global options] peer

 start, stop, list, delete peers
		`)
}

var options opts.Opts
var peerParser = opts.Parser(&options)

func (Peer) Run(args []string) int {
	_, err := peerParser.ParseArgs(args)
	if err != nil {
		if f, ok := err.(*flags.Error); ok {
			printErr(f.Message)
		}
		printErr(err)
	}
	return 0
}

func printErr(err interface{}) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func peerConnect(addr string) (*net.UDPConn, *net.UDPAddr, error) {
	var laddr *net.UDPAddr
	if addr != "" {
		var err error
		laddr, err = net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, nil, fmt.Errorf("error resolving local address: %w", err)
		}
	}
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("localhost:%s", options.Port))
	if err != nil {
		return nil, nil, fmt.Errorf("error resolving peer address: %w", err)
	}
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting peer address: %w", err)
	}
	return conn, raddr, nil
}

func (Peer) Synopsis() string {
	return "Peer management"
}
