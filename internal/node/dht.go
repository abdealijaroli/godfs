package node

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"hash/fnv"
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

    // Store the data locally
    d.data[key] = DataEntry{
        Value:     value,
        Version:   1,
        Timestamp: time.Now(),
    }

    // Replicate the data to other nodes
    for i := 0; i < replicationFactor; i++ {
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

	peer, err := d.transport.Dial(node)
	if err != nil {
		return err
	}
	defer peer.Close()

	return peer.Send(msg)
}

func Hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
