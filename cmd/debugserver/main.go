package main

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/abdealijaroli/godfs/internal/node"
)

type DebugServer struct {
	dht *node.DHT
}

func NewDebugServer(dht *node.DHT) *DebugServer {
	return &DebugServer{dht: dht}
}

func (s *DebugServer) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/nodes", s.handleNodes)
	// http.HandleFunc("/api/data", s.handleData)
	http.HandleFunc("/api/ring", s.handleRing)

	return http.ListenAndServe(":8080", nil)
}

func (s *DebugServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	tmpl.Execute(w, nil)
}

func (s *DebugServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.dht.ListNodes())
}

// func (s *DebugServer) handleData(w http.ResponseWriter, r *http.Request) {
//     json.NewEncoder(w).Encode(s.dht.GetAllData())
// }

func (s *DebugServer) handleRing(w http.ResponseWriter, r *http.Request) {
	nodes := s.dht.ListNodes()
	ringData := make([]uint32, len(nodes))
	for i, n := range nodes {
		ringData[i] = node.Hash(n)
	}
	json.NewEncoder(w).Encode(ringData)
}

func main() {
	dht := node.NewDHT("localhost:8000")
	debugServer := NewDebugServer(dht)

	// Add some nodes for testing
	dht.AddNode("localhost:8001")
	dht.AddNode("localhost:8002")
	dht.AddNode("localhost:8003")
	dht.AddNode("localhost:8004")
	dht.AddNode("localhost:8005")

	if err := debugServer.Start(); err != nil {
		panic(err)
	}
}
