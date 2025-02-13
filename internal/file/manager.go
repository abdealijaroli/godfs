package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/abdealijaroli/godfs/internal/crypto"
	"github.com/abdealijaroli/godfs/internal/node"
)

type FileManager struct {
	ChunkSize  int
	DHT        *node.DHT
	StorageDir string
}

func NewFileManager(chunkSize int, dht *node.DHT, storageDir string) *FileManager {
	return &FileManager{
		ChunkSize:  chunkSize,
		DHT:        dht,
		StorageDir: storageDir,
	}
}

func (fm *FileManager) UploadFile(filePath string, filename string) error {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file: %v", err)
	}

	// Split into chunks
	chunks := make([][]byte, 0)
	chunkSize := fm.ChunkSize
	maxChunks := 5 // Limit to 5 chunks for development purposes
	for i := 0; i < len(data) && i < chunkSize*maxChunks; i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, data[i:end])
	}

	// Store chunks in DHT
	chunkRefs := make([]string, len(chunks))
	for i, chunk := range chunks {
		chunkID := fmt.Sprintf("%s-chunk-%d", filename, i)
		if err := fm.DHT.Store(chunkID, chunk); err != nil {
			return fmt.Errorf("store chunk %d: %v", i, err)
		}
		chunkRefs[i] = chunkID
	}

	// Store file metadata
	metadata := &FileMetadata{
		Filename:  filename,
		ChunkRefs: chunkRefs,
	}

	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %v", err)
	}

	if err := fm.DHT.Store(filename, metadataBytes); err != nil {
		return fmt.Errorf("store metadata: %v", err)
	}

	return nil
}

type FileMetadata struct {
	Filename  string   `json:"filename"`
	ChunkRefs []string `json:"chunk_refs"`
}

func (fm *FileManager) UploadEncryptedFile(filePath string, encryptionKey []byte) error {
	chunks, err := SplitFile(filePath, fm.ChunkSize)
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		encryptedData, err := crypto.Encrypt(chunk.Data, encryptionKey)
		if err != nil {
			return err
		}

		chunkPath := path.Join(fm.StorageDir, chunk.ID)
		err = os.WriteFile(chunkPath, encryptedData, 0644)
		if err != nil {
			return err
		}

		err = fm.DHT.PutConsistent(chunk.ID, chunkPath, 3)
		if err != nil {
			return err
		}
	}

	fmt.Println("File encrypted and uploaded successfully.")
	return nil
}

func (fm *FileManager) DownloadFile(outputPath string, chunkIDs []string, encryptionKey []byte) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, chunkID := range chunkIDs {
		chunkPath, err := fm.DHT.Get(chunkID)
		if err != nil {
			return err
		}

		encryptedData, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}

		decryptedData, err := crypto.Decrypt(encryptedData, encryptionKey)
		if err != nil {
			return err
		}

		_, err = outputFile.Write(decryptedData)
		if err != nil {
			return err
		}
	}

	fmt.Println("File downloaded and decrypted successfully.")
	return nil
}
