package main

import (
	"log"

	"github.com/abdealijaroli/godfs/pkg/p2p"
)

func main() {
	t := p2p.NewTCPTransport(":8080")

	go func() {
		if err := t.ListenAndAccept(); err != nil {
			log.Printf("Transport error: %v", err)
		}
	}()

	select {}
}
