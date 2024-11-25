package main

import (
	"log"

	"github.com/abdealijaroli/godfs/pkg/p2p"
)

func main() {
	t := p2p.NewTCPTransport(":8080")

	err := t.ListenAndAccept()
	if err != nil {
		log.Fatal(err)
	}
}
