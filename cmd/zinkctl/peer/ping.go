package peer

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Joe-Degs/zinc"
	"github.com/Joe-Degs/zinc/internal/netutil"
	"github.com/Joe-Degs/zinc/internal/pool"
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

func sendPingPacket(conn *net.UDPConn) error {
	packet := []byte{byte(zinc.Ping)}
	if _, err := conn.Write(packet); err != nil {
		return err
	}
	return nil
}

func (l ping) Execute(args []string) error {
	l.help(args)

	addr := fmt.Sprintf("localhost:%s", options.Port)
	conn, err := netutil.TryConnect(addr, 1000)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	byts := make(chan []byte)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		for {
			if err := sendPingPacket(conn); err != nil {
				zinc.ZErrorf("could not write ping packet: %v", err)
				continue
			}

			buffer := pool.GetBufferSized(100)
			buf := buffer.Bytes()
			n, err := conn.Read(buf)
			if err != nil {
				zinc.ZErrorf("could not read ping response: %s", err.Error())
				continue
			}
			byts <- append(make([]byte, 0, n), buf[:n]...)
			pool.PutBuffer(buffer)
		}
	}()

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
					os.Exit(0)
				}
				zinc.ZPrintf("ping request recieved")
				pongCount += 1
			case zinc.Error:
				// recieved an error response, decode, print and stop
				printErr(fmt.Errorf("error pinging peer: %s", b[1:]))
			default:
				// packet type unknown
				return fmt.Errorf("recieved unknown packet from peer")
			}
		case <-ctx.Done():
			// fmt.Fprintf(os.Stderr, "timeout: peer at %s not active", addr)
			return fmt.Errorf("timeout: peer at %s not active", addr)
		}
	}
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
