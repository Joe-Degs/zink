package peer

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	addr := fmt.Sprintf("127.0.0.1:%s", options.Port)
	zinc.ZPrintf("address to start peer on: %s", addr)
	pier, err := zinc.PeerFromSpec(options.Name, addr, zinc.RandomUid)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Errorf("could not start peer: %w", err))
		os.Exit(1)
	}
	cl := make(chan io.Closer)
	cancel, err := pier.StartRequestReciever(cl)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return fmt.Errorf("could not start peer: %w", err)
	}
	started := true
	go handleShutdown(cl)
	if started {
		select {}
	}
	return nil
}

func handleShutdown(cl <-chan io.Closer) {
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
				log.Fatal("timeout shutting down...")
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
