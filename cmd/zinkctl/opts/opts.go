package opts

import "github.com/jessevdk/go-flags"

type Opts struct {
	Verbose bool   `short:"v" long:"verbose" description:"Verbose output"`
	Help    bool   `short:"h" long:"help" description:"Display help for command"`
	Name    string `short:"n" long:"name" description:"Peer or cluster name" optional:"yes"`
	Id      string `short:"i" long:"id" description:"unique id of peer or cluster" optional:"yes"`
	Port    string `short:"p" long:"port" default:"6009" description:"Address of peer server"`
}

func Parser(opts *Opts) *flags.Parser {
	return flags.NewParser(opts, flags.HelpFlag|flags.PassDoubleDash|flags.IgnoreUnknown)
}
