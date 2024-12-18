package protocol

import (
    "encoding/json"
    "net"
)

type Handshake struct {
    NodeID string
}

func PerformHandshake(conn net.Conn, nodeID string) error {
    encoder := json.NewEncoder(conn)
    decoder := json.NewDecoder(conn)

    err := encoder.Encode(Handshake{NodeID: nodeID})
    if err != nil {
        return err
    }

    var response Handshake
    err = decoder.Decode(&response)
    if err != nil {
        return err
    }

    println("Handshake successful with node:", response.NodeID)
    return nil
}
