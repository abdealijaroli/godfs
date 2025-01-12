package main

import (
    "log"
    "time"

    "github.com/abdealijaroli/godfs/config"
    "github.com/abdealijaroli/godfs/pkg/p2p"
)

func Dial() {
    tlsConfig, err := config.LoadTLSConfig("certs/client.crt", "certs/client.key", "certs/ca.crt")
    if err != nil {
        log.Fatal("Failed to load TLS config:", err)
    }

    transport := p2p.NewTCPTransport(":8081", tlsConfig)

    go func() {
        if err := transport.ListenAndAccept(); err != nil {
            log.Fatal(err)
        }
    }()

    time.Sleep(2 * time.Second)

    log.Println("Dialing server at localhost:8080")
    peer, err := transport.Dial("localhost:8080")
    if err != nil {
        log.Fatal("Failed to connect to server:", err)
    }

    err = peer.Send(p2p.Message{
        Type:    "greeting",
        Payload: []byte("Hello from the client node!"),
    })
    if err != nil {
        log.Fatal("Failed to send message:", err)
    }

    log.Println("Message sent successfully!")

    select {}
}
