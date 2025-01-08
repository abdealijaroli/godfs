package discovery

import (
	"sync"
)

type Peer struct {
	Address string
	Active  bool
}

type PeerDiscovery struct {
	Peers map[string]*Peer
	Lock  sync.RWMutex
}

func NewPeerDiscovery() *PeerDiscovery {
	return &PeerDiscovery{Peers: make(map[string]*Peer)}
}

func (pd *PeerDiscovery) AddPeer(address string) {
	pd.Lock.Lock()
	defer pd.Lock.Unlock()

	if _, exists := pd.Peers[address]; !exists {
		pd.Peers[address] = &Peer{Address: address, Active: true}
	}
}

func (pd *PeerDiscovery) GetPeers() []string {
	pd.Lock.RLock()
	defer pd.Lock.RUnlock()

	var peerList []string
	for addr, peer := range pd.Peers {
		if peer.Active {
			peerList = append(peerList, addr)
		}
	}
	return peerList
}
