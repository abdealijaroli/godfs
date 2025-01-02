package node_test

import (
	"testing"

	"github.com/abdealijaroli/godfs/internal/node"
)

func TestDHT(t *testing.T) {
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
	dht := node.NewDHT("node1")
	dht.AddNode("node2")
	dht.AddNode("node3")

	err := dht.PutConsistent("file1", "chunk1_data", 2)
	if err != nil {
		t.Fatalf("Failed to store data in DHT: %s", err)
	}

	value, err := dht.Get("file1")
	if err != nil || value != "chunk1_data" {
		t.Fatalf("Expected chunk1_data, got %s", value)
	}
}
