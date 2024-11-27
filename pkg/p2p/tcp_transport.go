package p2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

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

func (t *TCPTransport) Dial(address string) (Peer, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	peer := &TCPPeer{
		conn: conn,
	}
	t.lock.Lock()
	t.peers[address] = peer
	t.lock.Unlock()
	return peer, nil
}

func (t *TCPTransport) ListenAndAccept() error {
	listener, err := net.Listen("tcp", t.address)
	if err != nil {
		return err
	}
	t.listener = listener

	fmt.Println("Listening on", t.address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
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
