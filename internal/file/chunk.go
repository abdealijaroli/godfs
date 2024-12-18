package file

import (
	"crypto/sha256"
	"io"
	"os"
)

type Chunk struct {
	ID       string
	Data     []byte
	Checksum string
}

func SplitFile(filepath string, chunkSize int) ([]Chunk, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var chunks []Chunk
	buffer := make([]byte, chunkSize)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		data := buffer[:n]
		checksum := sha256.Sum256(data)
		chunk := Chunk{
			ID:       string(checksum[:8]),
			Data:     data,
			Checksum: string(checksum[:]),
		}
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

func VerifyChecksum(chunk Chunk) bool {
	computed := sha256.Sum256(chunk.Data)
	return string(computed[:]) == chunk.Checksum
}
