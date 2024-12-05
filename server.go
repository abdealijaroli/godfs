package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/abdealijaroli/godfs/pkg/p2p"
)

func Serve() {
	t := p2p.NewTCPTransport(":8080")

	go func() {
		if err := t.ListenAndAccept(); err != nil {
			log.Printf("Transport error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
