package node_test

import (
	"crypto/tls"
	"crypto/x509"
	"os"
	"testing"

	"github.com/abdealijaroli/godfs/internal/node"
)

func loadTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.key")
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("../certs/ca.crt")
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return tlsConfig, nil
}

func TestDHT(t *testing.T) {
	_, err := loadTLSConfig()
	if err != nil {
		t.Fatalf("Failed to load TLS config: %v", err)
	}

	dht := node.NewDHT("node1")

	dht.AddNode("node2")
	dht.AddNode("node3")

	nodes := dht.ListNodes()
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}

	dht.Put("file1", "chunk1_location", 3)
	value, err := dht.Get("file1")
	if err != nil || value != "chunk1_location" {
		t.Fatalf("Expected chunk1_location, got %s", value)
	}

	err = dht.Remove("file1")
	if err != nil {
		t.Fatalf("Failed to remove key: %s", err)
	}
}

func TestDHTReplication(t *testing.T) {
	_, err := loadTLSConfig()
	if err != nil {
		t.Fatalf("Failed to load TLS config: %v", err)
	}

	dht := node.NewDHT("node1")

	dht.AddNode("node2")
	dht.AddNode("node3")

	dht.Put("file1", "chunk1_location", 3)

	// Simulate replication
	err = dht.Replicate("file1", "chunk1_location")
	if err != nil {
		t.Fatalf("Failed to replicate key: %s", err)
	}

	value, err := dht.Get("file1")
	if err != nil || value != "chunk1_location" {
		t.Fatalf("Expected chunk1_location, got %s", value)
	}
}
