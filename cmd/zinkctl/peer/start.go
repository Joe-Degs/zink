package peer

import (
	"fmt"
	"os"
	"strings"

	"github.com/Joe-Degs/zinc"
)

// peer subcommand of the start command LOL!
type start struct{}

func (s start) Help() string {
	help := `
Usage: zinkctl [global options] peer start <options>
 
 help

Options:
-p --port:			port peer server should listen on
-n --name:			name of a zinc peer
-i --id:            id of a zinc peer`
	return strings.TrimSpace(help)
}

func (s start) Execute(args []string) error {
	s.help(args)
	// start the peer listener
	addr := fmt.Sprintf("127.0.0.1:%s", options.Port)
	zinc.ZPrintf("address to start peer on: %s", addr)
	pier, err := zinc.PeerFromSpec(options.Name, addr, zinc.RandomUid)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("could not get a peer: %w", err))
		os.Exit(1)
	}
	// start waiting for requests
	pier.StartRequestReciever()
	return nil
}

func (s start) help(args []string) {
	for _, v := range args {
		if v == "help" {
			fmt.Println(s.Help())
			os.Exit(0)
		}
	}
}

func (s start) Synopsis() string {
	return "Start a zinc Peer Server"
}

var s start

func init() {
	peerParser.AddCommand("start", s.Synopsis(), s.Help(), &s)
}
