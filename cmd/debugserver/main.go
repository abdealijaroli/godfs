package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/abdealijaroli/godfs/internal/file"
	"github.com/abdealijaroli/godfs/internal/node"
)

type DebugServer struct {
	dht         *node.DHT
	fileManager *file.FileManager
}

func NewDebugServer(dht *node.DHT, fileManager *file.FileManager) *DebugServer {
	return &DebugServer{dht: dht, fileManager: fileManager}
}

func (s *DebugServer) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/nodes", s.handleNodes)
	http.HandleFunc("/api/data", s.handleData)
	http.HandleFunc("/api/ring", s.handleRing)
	http.HandleFunc("/api/chunks", s.handleChunks)
	http.HandleFunc("/api/upload-dummy", s.handleUploadDummy)

	return http.ListenAndServe(":8080", nil)
}

func (s *DebugServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	tmpl.Execute(w, nil)
}

func (s *DebugServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.dht.ListNodes())
}

func (s *DebugServer) handleData(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(s.dht.GetAllData())
}

func (s *DebugServer) handleRing(w http.ResponseWriter, r *http.Request) {
	nodes := s.dht.ListNodes()
	ringData := make([]uint32, len(nodes))
	for i, n := range nodes {
		ringData[i] = node.Hash(n)
	}
	json.NewEncoder(w).Encode(ringData)
}

func (s *DebugServer) handleChunks(w http.ResponseWriter, r *http.Request) {
	chunkData := make(map[string][]string)
	for key, value := range s.dht.GetAllData() {
		chunkData[key] = append(chunkData[key], value.(string))
	}
	json.NewEncoder(w).Encode(chunkData)
}

func (s *DebugServer) handleUploadDummy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	dummyFilePath := "dummy_file.txt"
	dummyContent := []byte("This is a dummy file for testing chunking and distribution.")

	err := os.WriteFile(dummyFilePath, dummyContent, 0644)
	if err != nil {
		http.Error(w, "Failed to create dummy file", http.StatusInternalServerError)
		return
	}

	err = s.fileManager.UploadEncryptedFile(dummyFilePath, []byte("encryption-key"))
	if err != nil {
		http.Error(w, "Failed to upload dummy file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Dummy file uploaded successfully")
}

func main() {
	dht := node.NewDHT("localhost:8000")
	fileManager := file.NewFileManager(1024, dht, "storage")
	debugServer := NewDebugServer(dht, fileManager)

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
