package node

import (
    "errors"
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

func (d *DHT) Put(key, value string) {
    d.lock.Lock()
    defer d.lock.Unlock()
    d.data[key] = value
}

func (d *DHT) Get(key string) (string, error) {
    d.lock.RLock()
    defer d.lock.RUnlock()

    value, exists := d.data[key]
    if !exists {
        return "", errors.New("key not found")
    }
    return value, nil
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
