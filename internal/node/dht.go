package node

import (
	"errors"
	"hash/fnv"
	"log"
	"sort"
	"sync"
)

type DHT struct {
	data     map[string]string
	nodes    []string
	selfNode string
	lock     sync.RWMutex
}

func NewDHT(selfNode string) *DHT {
	return &DHT{
		data:     make(map[string]string),
		nodes:    []string{},
		selfNode: selfNode,
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

func (d *DHT) RemoveNode(node string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	for i, n := range d.nodes {
		if n == node {
			d.nodes = append(d.nodes[:i], d.nodes[i+1:]...)
			break
		}
	}
}

func (d *DHT) Put(key, value string, replicationFactor int) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data[key] = value

	for i, node := range d.nodes {
		if node == d.selfNode {
			continue
		}

		err := d.sendToNode(node, key, value)
		if err != nil {
			log.Printf("Failed to replicate to node %s: %v", node, err)
		}

		if i+1 >= replicationFactor {
			break
		}
	}
	return nil
}

func (d *DHT) PutConsistent(key, value string, replicationFactor int) error {
	targetNode := d.consistentHash(key)
	if targetNode == "" {
		return errors.New("no nodes available")
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	err := d.sendToNode(targetNode, key, value)
	if err != nil {
		log.Printf("Failed to replicate key %s to node %s: %v", key, targetNode, err)
		return err
	}

	// Also store locally
	d.data[key] = value

	// Replicate to additional nodes based on replication factor
	replicated := 1
	for _, node := range d.nodes {
		if node == targetNode || replicated >= replicationFactor {
			break
		}
		err := d.sendToNode(node, key, value)
		if err != nil {
			log.Printf("Failed to replicate key %s to node %s: %v", key, node, err)
			continue
		}
		replicated++
	}

	return nil
}

func (d *DHT) Get(key string) (string, error) {
	d.lock.RLock()
	defer d.lock.RUnlock()

	value, exists := d.data[key]
	if exists {
		return value, nil
	}

	targetNode := d.consistentHash(key)
	if targetNode != "" {
		return d.queryNode(targetNode, key)
	}

	return "", errors.New("key not found in DHT or replicas")
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

func (d *DHT) ListNodes() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()
	return append([]string{}, d.nodes...)
}

func (d *DHT) sendToNode(node, key, value string) error {
	log.Printf("Replicating key %s to node %s", key, node)
	// todo
	return nil
}

func (d *DHT) queryNode(node, key string) (string, error) {
	log.Printf("Querying key %s from node %s", key, node)
	// todo
	return "", errors.New("not implemented")
}

// consistent hashing
func hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (d *DHT) consistentHash(key string) string {
	if len(d.nodes) == 0 {
		return ""
	}

	keyHash := hash(key)

	sort.Slice(d.nodes, func(i, j int) bool {
		return hash(d.nodes[i]) < hash(d.nodes[j])
	})

	for _, node := range d.nodes {
		if hash(node) >= keyHash {
			return node
		}
	}

	return d.nodes[0]
}
