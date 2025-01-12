package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/abdealijaroli/godfs/config"
    "github.com/abdealijaroli/godfs/pkg/p2p"
)

func Serve() {
    tlsConfig, err := config.LoadTLSConfig("certs/server.crt", "certs/server.key", "certs/ca.crt")
    if err != nil {
        log.Fatal("Failed to load TLS config:", err)
    }

    t := p2p.NewTCPTransport(":8080", tlsConfig)

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
