package cluster

import (
	"fmt"
	"os"
	"strings"

	"github.com/Joe-Degs/zinc/cmd/zinkctl/opts"
	"github.com/jessevdk/go-flags"
)

type Cluster struct{}

func (Cluster) Help() string {
	return strings.TrimSpace(`
Usage: zinkctl [global options] cluster

 start, stop, list, delete peers in clusters
		`)
}

var options opts.Opts
var peerParser = opts.Parser(&options)

func (Cluster) Run(args []string) int {
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

func (Cluster) Synopsis() string {
	return "Cluster management"
}
