package file

import (
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

func (fm *FileManager) UploadEncryptedFile(filepath string, encryptionKey []byte) error {
	chunks, err := SplitFile(filepath, fm.ChunkSize)
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

func (fm *FileManager) DownloadFile(outputPath string, chunkIDs []string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	for _, chunkID := range chunkIDs {
		chunkPath, err := fm.DHT.Get(chunkID)
		if err != nil {
			return fmt.Errorf("chunk %s not found in DHT", chunkID)
		}

		data, err := os.ReadFile(chunkPath)
		if err != nil {
			return err
		}

		_, err = outputFile.Write(data)
		if err != nil {
			return err
		}
	}

	fmt.Println("File downloaded and assembled successfully.")
	return nil
}
