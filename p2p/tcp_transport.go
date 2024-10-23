package p2p

import (
	"net"
)

type TCPTransport struct {
	listenAddr string
	listener   net.Listener

	// peers map[string]
}