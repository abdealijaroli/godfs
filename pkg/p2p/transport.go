package p2p

// handles communication between nodes in the network
type Transport interface {
	Dial(addr string) (Peer, error)
	ListenAndAccept() error
	Broadcast(msg []byte) error
	Close() error
}

// represents a remote node
type Peer interface {
	Send(msg Message) error
	Receive() (Message, error)
	Close() error
}

// represents a message that can be sent between nodes
type Message struct {
	// ID        string
	Type    string
	Payload []byte
	// Sender    Peer
	// Timestamp int64
}
