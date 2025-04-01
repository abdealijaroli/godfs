package main

import (
	// "crypto/tls"
	// "crypto/x509"
	"encoding/json"
	"flag"

	// "fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	// "path/filepath"

	// "github.com/abdealijaroli/godfs/config"
	"github.com/abdealijaroli/godfs/internal/file"
	"github.com/abdealijaroli/godfs/internal/node"
	"github.com/abdealijaroli/godfs/pkg/p2p"
)

type DebugServer struct {
	dht         *node.DHT
	fileManager *file.FileManager
}

// For dev purposes, we are not using TLS
func NewDebugServer(dht *node.DHT, fileManager *file.FileManager) *DebugServer {
	return &DebugServer{dht: dht, fileManager: fileManager}
}

func (s *DebugServer) Handler() http.Handler {
	mux := http.NewServeMux()

	// Static files
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// API endpoints
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/api/nodes", s.handleNodes)
	mux.HandleFunc("/api/data", s.handleData)
	mux.HandleFunc("/api/ring", s.handleRing)
	mux.HandleFunc("/api/chunks", s.handleChunks)
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/health", s.handleHealth)

	return mux
}

// func (s *DebugServer) Start(port string) error {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", s.handleDashboard)
// 	mux.HandleFunc("/api/nodes", s.handleNodes)
// 	mux.HandleFunc("/api/data", s.handleData)
// 	mux.HandleFunc("/api/ring", s.handleRing)
// 	mux.HandleFunc("/api/chunks", s.handleChunks)
// 	mux.HandleFunc("/api/upload", s.handleUpload)

// 	log.Printf("Starting server on port %s", port)

// 	tlsConfig, err := config.LoadTLSConfig("certs/server.crt", "certs/server.key", "certs/ca.crt")
// 	if err != nil {
// 		return err
// 	}

// 	server := &http.Server{
// 		Addr:      port,
// 		Handler:   mux,
// 		TLSConfig: tlsConfig,
// 	}
// 	return server.ListenAndServeTLS("certs/server.crt", "certs/server.key")
// }

func (s *DebugServer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/dashboard.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		return
	}
}

func (s *DebugServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Failed to parse form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file: %v", err)
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temp file
	tempFile, err := os.CreateTemp("", "upload-*.tmp")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy file data
	if _, err := io.Copy(tempFile, file); err != nil {
		log.Printf("Failed to copy file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Upload to DHT
	if err := s.fileManager.UploadFile(tempFile.Name(), header.Filename); err != nil {
		log.Printf("Failed to upload file: %v", err)
		http.Error(w, "Failed to distribute file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "File uploaded and distributed successfully",
		"filename": header.Filename,
	})
}

func (s *DebugServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	port := flag.String("port", "8000", "Port to run the server on")
	flag.Parse()

	// Create DHT without TLS for Dev mode
	dht := node.NewDHT("localhost:" + *port)
	fileManager := file.NewFileManager(1024, dht, "storage")
	debugServer := NewDebugServer(dht, fileManager)

	// Add nodes for Dev mode
	dht.AddNode("localhost:8443")
	dht.AddNode("localhost:8444")
	dht.AddNode("localhost:8445")
	dht.AddNode("localhost:8446")
	dht.AddNode("localhost:8447")

	// Create transport without TLS for Dev mode
	transport := p2p.NewTCPTransport("localhost:" + *port)
	go func() {
		if err := transport.ListenAndAccept(); err != nil {
			log.Fatalf("Failed to start transport: %v", err)
		}
	}()

	// Start HTTP server without TLS for Dev mode
	log.Printf("Starting HTTP server on port %s", *port)
	if err := http.ListenAndServe(":"+*port, debugServer.Handler()); err != nil {
		panic(err)
	}
}

// func main() {
// 	port := flag.String("port", "8000", "Port to run the server on")
// 	flag.Parse()

// 	cert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
// 	if err != nil {
// 		log.Fatalf("Failed to load TLS certificates: %v", err)
// 	}

// 	caCert, err := os.ReadFile("certs/ca.crt")
// 	if err != nil {
// 		log.Fatalf("Failed to load CA certificate: %v", err)
// 	}

// 	caCertPool := x509.NewCertPool()
// 	caCertPool.AppendCertsFromPEM(caCert)

// 	tlsConfig := &tls.Config{
// 		Certificates: []tls.Certificate{cert},
// 		RootCAs:      caCertPool,
// 	}

// 	dht := node.NewDHT("localhost:"+*port, tlsConfig)
// 	fileManager := file.NewFileManager(1024, dht, "storage")
// 	debugServer := NewDebugServer(dht, fileManager)

// 	dht.AddNode("localhost:8443")
// 	dht.AddNode("localhost:8444")
// 	dht.AddNode("localhost:8445")
// 	dht.AddNode("localhost:8446")
// 	dht.AddNode("localhost:8447")

// 	transport := p2p.NewTCPTransport("localhost:"+*port, tlsConfig)
// 	go func() {
// 		if err := transport.ListenAndAccept(); err != nil {
// 			log.Fatalf("Failed to start transport: %v", err)
// 		}
// 	}()

// 	if err := debugServer.Start(":" + *port); err != nil {
// 		panic(err)
// 	}
// }
