package main

import (
	"os"

	"github.com/Joe-Degs/zinc"
	"github.com/Joe-Degs/zinc/cmd/zinkctl/cluster"
	"github.com/Joe-Degs/zinc/cmd/zinkctl/peer"
	"github.com/mitchellh/cli"
)

const version = "0.0.1"

// zinkctl [global optins] [peer|cluster] <subcommands> <options>

func main() {
	c := cli.NewCLI("zinkctl", version)
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"peer": func() (cli.Command, error) {
			return &peer.Peer{}, nil
		},
		"cluster": func() (cli.Command, error) {
			return &cluster.Cluster{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		zinc.ZErrorf("%v", err)
	}
	os.Exit(exitStatus)
}
