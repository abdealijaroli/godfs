package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer is the remote node in a TCP connection.
type TCPPeer struct {
	conn     net.Conn
	outbound bool // true if we dialed the connection, false if we accepted it
}

// TCPTransport is a transport implementation that uses TCP sockets.
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

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	go t.StartAcceptLoop()

	return nil
}

func (t *TCPTransport) StartAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Accept error: ", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	fmt.Println("new incoming conn")
}
