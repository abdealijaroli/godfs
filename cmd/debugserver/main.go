package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/abdealijaroli/godfs/internal/file"
	"github.com/abdealijaroli/godfs/internal/node"
	"github.com/abdealijaroli/godfs/pkg/p2p"
)

type DebugServer struct {
	dht         *node.DHT
	fileManager *file.FileManager
}

func NewDebugServer(dht *node.DHT, fileManager *file.FileManager) *DebugServer {
	return &DebugServer{dht: dht, fileManager: fileManager}
}

func (s *DebugServer) Start(port string) error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/nodes", s.handleNodes)
	http.HandleFunc("/api/data", s.handleData)
	http.HandleFunc("/api/ring", s.handleRing)
	http.HandleFunc("/api/chunks", s.handleChunks)
	http.HandleFunc("/api/upload", s.handleUpload)

	log.Printf("Starting server on port %s", port)
	return http.ListenAndServe(port, nil) // fix tls config
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
	chunkData := make(map[string][]node.DataEntry)
	for key, value := range s.dht.GetAllData() {
		dataEntry, ok := value.(node.DataEntry)
		if !ok {
			http.Error(w, "invalid data entry type", http.StatusInternalServerError)
			return
		}
		chunkData[key] = append(chunkData[key], dataEntry)
	}
	json.NewEncoder(w).Encode(chunkData)
}

func (s *DebugServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	encryptionKey := []byte("1234567890123456")
	filePath := filepath.Join("storage", "dummy.txt")

	err := os.MkdirAll("storage", os.ModePerm)
	if err != nil {
		log.Printf("Failed to create storage directory: %v", err)
		http.Error(w, "Failed to create storage directory", http.StatusInternalServerError)
		return
	}

	dummyContent := []byte("This is a dummy file for testing chunking and distribution.")
	err = os.WriteFile(filePath, dummyContent, 0644)
	if err != nil {
		log.Printf("Failed to create dummy file: %v", err)
		http.Error(w, "Failed to create dummy file", http.StatusInternalServerError)
		return
	}

	err = s.fileManager.UploadEncryptedFile(filePath, encryptionKey)
	if err != nil {
		log.Printf("Failed to upload dummy file: %v", err)
		http.Error(w, "Failed to upload dummy file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Dummy file uploaded successfully")
}

func main() {
	port := flag.String("port", "8000", "Port to run the server on")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatalf("Failed to load TLS certificates: %v", err)
	}

	caCert, err := os.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	dht := node.NewDHT("localhost:"+*port, tlsConfig)
	fileManager := file.NewFileManager(1024, dht, "storage")
	debugServer := NewDebugServer(dht, fileManager)

	dht.AddNode("localhost:8001")
	dht.AddNode("localhost:8002")
	dht.AddNode("localhost:8003")
	dht.AddNode("localhost:8004")
	dht.AddNode("localhost:8005")

	transport := p2p.NewTCPTransport("localhost:"+*port, tlsConfig)
	go func() {
		if err := transport.ListenAndAccept(); err != nil {
			log.Fatalf("Failed to start transport: %v", err)
		}
	}()

	if err := debugServer.Start(":" + *port); err != nil {
		panic(err)
	}
}
