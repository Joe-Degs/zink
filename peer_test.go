package zinc

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"inet.af/netaddr"
)

func getIPPort(addr string) *netaddr.IPPort {
	laddr, err := netaddr.ParseIPPort(addr)
	if err != nil {
		zlog.Panicln(err)
	}
	return &laddr
}

func TestPeerJSON(t *testing.T) {
	fn := func(t *testing.T, b []byte, p Peer) {
		peer := &Peer{}
		if err := peer.UnmarshalJSON(b); err != nil {
			t.Fatalf("failed to unmarshal json: %v", err)
		}

		if diff := cmp.Diff(p, *peer, cmpopts.IgnoreUnexported(netaddr.IPPort{}, Peer{})); diff != "" {
			t.Fatalf("Json marshaling and unmarshaling anomaly: (-want +got):\n%s", diff)
		}

		if p.LocalAddr.IsValid() && p.LocalAddr.IP() != peer.LocalAddr.IP() {
			t.Fatalf("IPPort mismatch: want %s; got %s", p.LocalAddr.IP(), p.LocalAddr.IP())
		}
	}

	testCases := []struct {
		name string
		peer Peer
	}{
		{
			name: "peer with all 3 fields set",
			peer: Peer{
				Name:      "jsonPeer1",
				Id:        RandomUid(),
				LocalAddr: getIPPort("192.168.43.101:6969"),
			},
		}, {
			name: "peer without name",
			peer: Peer{
				Id:        RandomUid(),
				LocalAddr: getIPPort("[::]:8869"),
			},
		}, {
			name: "peer without local address",
			peer: Peer{
				Id:   RandomUid(),
				Name: "jsonPeer2",
			},
		}, {
			name: "peer without local address and name",
			peer: Peer{
				Id: RandomUid(),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(&tc.peer)
			if err != nil {
				t.Fatalf("failed to marshal peer to json: %v", err)
			}

			fn(t, b, tc.peer)
		})
	}

}

func TestPeerText(t *testing.T) {
	fn := func(t *testing.T, b []byte, p Peer) {
		peer := &Peer{}
		if err := peer.UnmarshalText(b); err != nil {
			t.Fatalf("failed to unmarshal text: %v", err)
		}

		diff := cmp.Diff(p, *peer, cmpopts.IgnoreUnexported(netaddr.IPPort{}, Peer{}))
		if diff != "" {
			t.Fatalf("Text marshaling and unmarshaling anomaly: (-want +got):\n%s", diff)
		}

		if p.LocalAddr.IsValid() && p.LocalAddr.IP() != peer.LocalAddr.IP() {
			t.Fatalf("IPPort mismatch: want %s; got %s", p.LocalAddr.IP(), p.LocalAddr.IP())
		}
	}

	testCases := []struct {
		name string
		peer Peer
	}{
		{
			name: "peer with all 3 fields set",
			peer: Peer{
				Name:      "textPeer1",
				Id:        RandomUid(),
				LocalAddr: getIPPort("192.168.43.101:6969"),
			},
		}, {
			name: "text with just name and id",
			peer: Peer{
				Name: "textPeer1",
				Id:   RandomUid(),
			},
		}, {
			name: "peer without name and id",
			peer: Peer{
				Id: RandomUid(),
			},
		}, {
			name: "peer with just id and ipport",
			peer: Peer{
				Id:        RandomUid(),
				LocalAddr: getIPPort("[1b20:485b:12a5:024c:551e:e040:04e0:f9c0]:6969"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.peer.MarshalText()
			if err != nil {
				t.Fatalf("failed to marshal peer to text: %v", err)
			}

			fn(t, b, tc.peer)
		})
	}

}
