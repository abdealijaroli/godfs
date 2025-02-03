package p2p

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/abdealijaroli/godfs/internal/discovery"
	"github.com/abdealijaroli/godfs/pkg/protocol"
)

// TCPPeer
type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func (p *TCPPeer) Send(msg Message) error {
	encoder := json.NewEncoder(p.conn)
	p.outbound = true
	return encoder.Encode(msg)
}

func (p *TCPPeer) Receive() (Message, error) {
	decoder := json.NewDecoder(bufio.NewReader(p.conn))
	p.outbound = false
	var msg Message
	err := decoder.Decode(&msg)
	return msg, err
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

// TCPTransport
type TCPTransport struct {
	address  string
	listener net.Listener
	peers    map[string]Peer
	lock     sync.Mutex
	config   *tls.Config
}

func NewTCPTransport(address string, config *tls.Config) *TCPTransport {
	return &TCPTransport{
		address: address,
		peers:   make(map[string]Peer),
		config:  config,
	}
}

func (t *TCPTransport) Dial(address string) (Peer, error) {
	log.Printf("Attempting to dial %s", address)

	conn, err := tls.Dial("tcp", address, t.config)
	if err != nil {
		log.Printf("TLS dial error: %v", err)
		return nil, err
	}

	peer := &TCPPeer{conn: conn}

	err = protocol.PerformHandshake(conn, address)
	if err != nil {
		log.Printf("Handshake failed with %s: %v", address, err)
		conn.Close()
		return nil, err
	}

	log.Printf("Connected successfully to %s", address)
	t.lock.Lock()
	t.peers[address] = peer
	t.lock.Unlock()

	return peer, nil
}

func (t *TCPTransport) ListenAndAccept() error {
	listener, err := tls.Listen("tcp", t.address, t.config)
	if err != nil {
		return err
	}
	t.listener = listener

	fmt.Println("Listening on", t.address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		peer := &TCPPeer{
			conn: conn,
		}
		t.lock.Lock()
		t.peers[conn.RemoteAddr().String()] = peer
		t.lock.Unlock()

		go t.handleConnection(peer)
	}
}

func (t *TCPTransport) Broadcast(msg Message) error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, peer := range t.peers {
		err := peer.Send(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *TCPTransport) Close() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	for _, peer := range t.peers {
		peer.Close()
	}
	return t.listener.Close()
}

func (t *TCPTransport) handleConnection(peer Peer) {
	for {
		msg, err := peer.Receive()
		if err != nil {
			fmt.Println("Error receiving message:", err)
			return
		}
		fmt.Printf("Received message: %s\n", string(msg.Payload))
	}
}

func (t *TCPTransport) ConnectToPeers(peerDiscovery *discovery.PeerDiscovery) error {
	peers := peerDiscovery.GetPeers()
	for _, peerAddr := range peers {
		_, err := t.Dial(peerAddr)
		if err != nil {
			log.Println("Error connecting to peer:", err)
			continue
		}
		log.Println("Connected to peer:", peerAddr)
	}
	return nil
}
