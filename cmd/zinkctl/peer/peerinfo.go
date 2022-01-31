package peer

import (
	"fmt"
	"os"
	"strings"

	"github.com/Joe-Degs/zinc"
	"github.com/Joe-Degs/zinc/internal/netutil"
)

// peerinfo subcommand of the peer command LOL!
// peerinfo sends a peerinfo request and waits for a peerinfo request back from
// the peer.
type peerinfo struct{}

func (l peerinfo) Help() string {
	help := `
Usage: zinkctl [global options] peer peerinfo <options>
 
 help

Options:
-p --port:			port peer server should peerinfoen on
-n --name:			name of a zinc peer
-i --id:            id of a zinc peer
-r --from-registry: read peer spec from registry
	`
	return strings.TrimSpace(help)
}

func (l peerinfo) Execute(args []string) error {
	l.help(args)
	conn, err := netutil.TryConnect("localhost:11121", 1000)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()
	// a peerinfo request packet
	packet := []byte{byte(zinc.PeerInfo), 0xf}
	if _, err := conn.Write(packet); err != nil {
		zinc.ZErrorf("could not write peerinfo packet: %v", err)
		os.Exit(1)
	}
	return nil
}

func (l peerinfo) help(args []string) {
	for _, v := range args {
		if v == "help" {
			fmt.Println(l.Help())
			os.Exit(0)
		}
	}
}

func (l peerinfo) Synopsis() string {
	return "get information on a zinc peer"
}

var pf peerinfo

func init() {
	peerParser.AddCommand("peerinfo", s.Synopsis(), s.Help(), &pf)
}
