package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	conn     net.Conn
	outbound bool
}

func (p *TCPPeer) Send(msg Message) error {
	encoder := json.NewEncoder(p.conn)
	return encoder.Encode(msg)
}

func (p *TCPPeer) Receive() (Message, error) {
	decoder := json.NewDecoder(bufio.NewReader(p.conn))
	var msg Message
	err := decoder.Decode(&msg)
	return msg, err
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransport struct {
	address  string
	listener net.Listener
	peers    map[string]Peer
	lock     sync.Mutex
}

func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{
		address: address,
		peers:   make(map[string]Peer),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.address)
	if err != nil {
		return err
	}

	fmt.Printf("Listening on %s\n", t.address)
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v\n", err)
			continue
		}

		peer := &TCPPeer{conn: conn}
		t.lock.Lock()
		t.peers[conn.RemoteAddr().String()] = peer
		t.lock.Unlock()

		go t.handleConnection(peer)
	}
}

func (t *TCPTransport) Dial(address string) (Peer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	peer := &TCPPeer{conn: conn, outbound: true}
	t.lock.Lock()
	t.peers[address] = peer
	t.lock.Unlock()

	return peer, nil
}

func (t *TCPTransport) handleConnection(peer *TCPPeer) {
	defer func() {
		t.lock.Lock()
		delete(t.peers, peer.conn.RemoteAddr().String())
		t.lock.Unlock()
		peer.Close()
	}()

	for {
		msg, err := peer.Receive()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			return
		}

		log.Printf("Received message from %s: %+v",
			peer.conn.RemoteAddr().String(), msg)
	}
}
