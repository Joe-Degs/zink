package opts

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jessevdk/go-flags"
)

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

func HandleShutdown(cl <-chan io.Closer) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		sign := <-signals
		sig, ok := sign.(syscall.Signal)
		if !ok {
			log.Fatal(sign.String() + "not a posix signal")
		}
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// end the process
			log.Printf("recieved signal '%s'", sign.String())
			donec := make(chan bool)
			go func() {
				c := <-cl
				if err := c.Close(); err != nil {
					log.Fatalf("error shutting down: %s", err.Error())
				}
				donec <- true
			}()
			select {
			case <-donec:
				os.Exit(0)
			case <-time.After(3 * time.Second):
				log.Fatal("timeout shutting down")
			}
		case syscall.SIGHUP:
			// restart
			log.Println("restarting process...")
			continue
		default:
			// unexpected something
			log.Fatalf("unexpected signal '%s'", sign.String())
		}
	}
}
