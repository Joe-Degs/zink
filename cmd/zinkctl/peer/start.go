package peer

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Joe-Degs/zinc"
	"github.com/Joe-Degs/zinc/cmd/zinkctl/opts"
	"github.com/Joe-Degs/zinc/internal/config"
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

	// addr := fmt.Sprintf("127.0.0.1:%s", options.Port)
	// zinc.ZPrintf("address to start peer on: %s", addr)
	// pier, err := zinc.PeerFromSpec(options.Name, addr, uuid.New())
	conf, err := config.DefaultPeerConfig()
	if err != nil {
		return fmt.Errorf("error loading configs: %w", err)
	}
	pier, err := zinc.NewPeer(conf)
	if err != nil {
		return fmt.Errorf("could not start peer: %w", err)
	}
	cl := make(chan io.Closer)
	cancel, err := pier.StartServer(cl)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return fmt.Errorf("could not start peer: %w", err)
	}
	started := true
	go opts.HandleShutdown(cl)
	if started {
		select {}
	}
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
