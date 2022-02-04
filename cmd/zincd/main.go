package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/Joe-Degs/zinc"
	"github.com/google/uuid"
)

func main() {
	peer1 := zinc.RandomPeer("node1")
	peer2, err := zinc.PeerFromSpec("node2", "0.0.0.0:60009", uuid.New())
	if err != nil {
		log.Fatal(err)
	}

	peer3 := &zinc.Peer{}
	var pjson = "{\"name\": \"node3\", \"id\": \"" + zinc.RandomUid().String() + "\", \"addr\": \"[::]:60009\"}"
	err = peer3.UnmarshalJSON([]byte(pjson))
	if err != nil {
		log.Fatal(err)
	}

	// output some logs see how it goes
	text, _ := peer1.MarshalText()
	peer1.UnmarshalText(text)
	zinc.ZErrorf("%v", errors.New("Error before everything"))
	zinc.ZPrintf("%s", text)
	zinc.ZErrorf("%v", errors.New("The new error in town"))

	pprint(peer1, peer2, peer3)
}

func pprint(data ...interface{}) {
	for _, datum := range data {
		b, err := json.MarshalIndent(datum, "", "  ")
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Printf("%s\n", b)
	}
}
