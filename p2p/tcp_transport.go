package p2p

import (
	"net"
	"sync"
)

type TCPTransport struct {
	listenAddr string
	listener   net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: listenAddr,
		peers:      make(map[net.Addr]Peer),
	}
}

func (t *TCPTransport) Start() error {
	listener, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = listener
	go t.acceptConnections()

	return nil
}

func (t *TCPTransport) Stop() error {
}


// to be cont.