package protocol

import (
	"encoding/json"
	"net"
	"fmt"
	"log"
)

type Handshake struct {
	NodeID string
}

func PerformHandshake(conn net.Conn, nodeID string) error {
	hs := Handshake{NodeID: nodeID}

	// Send handshake
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(hs); err != nil {
		return fmt.Errorf("failed to send handshake: %v", err)
	}

	// Read response
	decoder := json.NewDecoder(conn)
	var response Handshake
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("failed to read handshake response: %v", err)
	}

	log.Printf("Handshake successful with node: %s", response.NodeID)
	return nil
}
