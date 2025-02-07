package node

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"hash/fnv"
	"fmt"
	"sync"
	"time"

	"github.com/abdealijaroli/godfs/pkg/p2p"
)

type DHT struct {
	data      map[string]DataEntry
	nodes     []string
	selfNode  string
	lock      sync.RWMutex
	transport *p2p.TCPTransport
}

type DataEntry struct {
	Value     string
	Version   int64
	Timestamp time.Time
}

func NewDHT(selfNode string, tlsConfig *tls.Config) *DHT {
	return &DHT{
		data:      make(map[string]DataEntry),
		nodes:     []string{},
		selfNode:  selfNode,
		transport: p2p.NewTCPTransport(selfNode, tlsConfig),
	}
}

func (d *DHT) AddNode(node string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, n := range d.nodes {
		if n == node {
			return
		}
	}
	d.nodes = append(d.nodes, node)
}

func (d *DHT) ListNodes() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return append([]string{}, d.nodes...)
}

func (d *DHT) GetAllData() map[string]interface{} {
	d.lock.RLock()
	defer d.lock.RUnlock()
	data := make(map[string]interface{})
	for k, v := range d.data {
		data[k] = v
	}
	return data
}

func (d *DHT) Remove(key string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, exists := d.data[key]; !exists {
		return errors.New("key not found")
	}

	delete(d.data, key)
	return nil
}

func (d *DHT) Put(key, value string, version int64) {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.data[key] = DataEntry{
		Value:     value,
		Version:   version,
		Timestamp: time.Now(),
	}
}

func (d *DHT) Get(key string) (string, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	entry, exists := d.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	return entry.Value, nil
}

func (d *DHT) Replicate(key, value string) error {
	for _, node := range d.nodes {
		err := d.sendToNode(node, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DHT) PutConsistent(key, value string, replicationFactor int) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	// Store data locally
	d.data[key] = DataEntry{
		Value:     value,
		Version:   1,
		Timestamp: time.Now(),
	}

	if len(d.nodes) == 0 {
		return errors.New("no available nodes to replicate")
	}

	// Replicate to available nodes
	for i := 0; i < replicationFactor && i < len(d.nodes); i++ {
		node := d.nodes[i%len(d.nodes)]
		err := d.sendToNode(node, key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DHT) sendToNode(node, key, value string) error {
	payload := map[string]string{
		"key":   key,
		"value": value,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := p2p.Message{
		Type:    "dht_store",
		Payload: data,
	}

	const maxRetries = 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		peer, err := d.transport.Dial(node)
		if err != nil {
			lastErr = err
			continue
		}
		defer peer.Close()

		// Send data
		err = peer.Send(msg)
		if err != nil {
			return err
		}

		// Read acknowledgment
		resp, err := peer.Receive()
		if err != nil {
			return err
		}

		if resp.Type != "ack" {
			return errors.New("unexpected response from node")
		}

		return nil
	}
	return fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

func Hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
