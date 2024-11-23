package main

import (
	"log"

	"github.com/abdealijaroli/godfs/p2p"
)

func main() {
	// create a new TCP transport
	t := p2p.NewTCPTransport(":8080")

	// start listeningn for incoming connections
	err := t.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}
	for {
	}
}
