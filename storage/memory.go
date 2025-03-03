package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ghulamazad/GFileMux"
)

// MemoryStorage is an in-memory storage client.
type MemoryStorage struct {
}

// NewMemoryStorage initializes a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// Upload uploads a file to memory storage.
func (ms *MemoryStorage) Upload(ctx context.Context, r io.Reader, options *GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	// Ensure the file name is not empty
	if len(strings.TrimSpace(options.FileName)) == 0 {
		return nil, fmt.Errorf("file name is required")
	}

	// Read the file contents into memory
	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		return nil, err
	}

	// Return metadata for the uploaded file
	return &GFileMux.UploadedFileMetadata{
		FolderDestination: "memory", // All files are stored in memory
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Path generates a path for accessing the file in memory storage.
// Since it's in memory, we don't have a URL, so we return a file path-like string.
func (ms *MemoryStorage) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	return fmt.Sprintf("memory://%s/%s", "memory", options.Key), nil
}

// Close closes the memory storage (no-op for in-memory storage).
func (ms *MemoryStorage) Close() error {
	// No cleanup required for memory storage.
	return nil
}
