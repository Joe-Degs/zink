package peer

import (
	"fmt"
	"os"
	"strings"
)

// peer subcommand of the list command LOL!
type list struct {
	FromRegistry bool `short:"r" long:"from-registry" description:"load peer spec from registry"`
}

func (l list) Help() string {
	help := `
Usage: zinkctl [global options] peer list <options>
 
 help

Options:
-p --port:			port peer server should listen on
-n --name:			name of a zinc peer
-i --id:            id of a zinc peer
-r --from-registry: read peer spec from registry
	`
	return strings.TrimSpace(help)
}

func (l list) Execute(args []string) error {
	l.help(args)
	return nil
}

func (l list) help(args []string) {
	for _, v := range args {
		if v == "help" {
			fmt.Println(l.Help())
			os.Exit(0)
		}
	}
}

func (l list) Synopsis() string {
	return "list a zinc Peer Server"
}

var l list

func init() {
	peerParser.AddCommand("list", s.Synopsis(), s.Help(), &l)
}
