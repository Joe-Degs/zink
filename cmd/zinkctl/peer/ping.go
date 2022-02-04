package peer

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"text/tabwriter"

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
	conn, err := netutil.Connect(addr)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	defer conn.Close()

	if err := sendPingPacket(conn); err != nil {
		zinc.ZErrorf("could not write ping packet: %v", err)
	}

	buffer := pool.GetBufferSized(100)
	defer pool.PutBuffer(buffer)
	buf := buffer.Bytes()
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	packet, err := zinc.UnmarshalPacket(buf[:n])
	if err != nil {
		return err
	}

	switch packet.Type() {
	case zinc.Error:
		return fmt.Errorf(string(packet.Data()))
	case zinc.Pong:
		fmt.Println("recieved pong packet")
	case zinc.PeerInfo:
		var rePeer zinc.Peer
		if err := json.Unmarshal(packet.Data(), &rePeer); err != nil {
			log.Fatal(err)
		}

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 2, '\t', 0)
		fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "Name", "ID", "Address", "Status")
		fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "----", "--", "-------", "------")
		fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", rePeer.Name, rePeer.Id.String(),
			rePeer.LocalAddr.String(), "ACTIVE")
		fmt.Fprintln(w, "")
		w.Flush()
	default:
		return fmt.Errorf("unknown response from peer")
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
