package peer

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Joe-Degs/zinc"
)

// ping subcommand of the peer command LOL!
type ping struct{}

func (l ping) Help() string {
	help := `
Usage: zinkctl [global options] peer ping <options>
 
 help

Options:
-p --port:			port peer server should pingen on
-n --name:			name of a zinc peer
-i --id:            id of a zinc peer
-r --from-registry: read peer spec from registry
	`
	return strings.TrimSpace(help)
}

func (l ping) Execute(args []string) error {
	l.help(args)

	conn, _, err := peerConnect("localhost:11112")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	// fashion a nice ping request for the zinc peer
	packet := []byte{0xf, 0xf}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	byts := make(chan []byte)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		for {
			// send request
			if _, err := conn.Write(packet); err != nil {
				zinc.ZErrorf("could not write ping packet: %v", err)
				continue
			}

			// wait for response
			buf := make([]byte, 256)
			n, err := conn.Read(buf)
			if err != nil {
				zinc.ZErrorf("could not read ping response: %s", err.Error())
				continue
			}
			byts <- buf[:n]
		}
	}()

	// count the number of pong response from the peer. If you get three, the
	// peer is active and kicking. And you can stop pinging
	var pongCount int

	for {
		select {
		case sig := <-signals:
			printErr(fmt.Sprintf("%s signal recieved", sig.String()))
		case b := <-byts:
			switch zinc.PacketType(b[0]) {
			case zinc.Pong:
				if pongCount > 3 {
					// request peer info and stop sending ping requests.
					zinc.ZPrintf("peer is active")
					os.Exit(1)
				}
				zinc.ZPrintf("ping request recieved")
				pongCount += 1
			case zinc.Error:
				// recieved an error response, decode, print and stop
				printErr(fmt.Errorf("error pinging peer: %s", b[1:]))
			default:
				// packet type unknown
				zinc.ZErrorf("recieved unknown packet from peer")
			}
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func (l ping) help(args []string) {
	for _, v := range args {
		if v == "help" {
			fmt.Println(l.Help())
			os.Exit(0)
		}
	}
}

func (l ping) Synopsis() string {
	return "ping a zinc Peer Server"
}

var pi ping

func init() {
	peerParser.AddCommand("ping", s.Synopsis(), s.Help(), &pi)
}
