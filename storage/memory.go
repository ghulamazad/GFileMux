package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ghulamazad/GFileMux"
)

// MemoryStorage is a thread-safe, in-memory storage backend.
// Stored files survive the lifetime of the process and are keyed by
// "<bucket>/<filename>". This backend is primarily intended for testing.
type MemoryStorage struct {
	mu    sync.RWMutex
	store map[string][]byte // key → file bytes
}

// NewMemoryStorage initializes a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		store: make(map[string][]byte),
	}
}

// storeKey returns the internal map key for a bucket+filename pair.
func storeKey(bucket, fileName string) string {
	if bucket == "" {
		return fileName
	}
	return bucket + "/" + fileName
}

// Upload reads the file into memory and stores it by bucket+filename key.
func (ms *MemoryStorage) Upload(ctx context.Context, r io.Reader, options *GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	if options == nil || len(strings.TrimSpace(options.FileName)) == 0 {
		return nil, fmt.Errorf("file name is required")
	}

	var buf bytes.Buffer
	n, err := io.Copy(&buf, r)
	if err != nil {
		return nil, &GFileMux.StorageError{Backend: "memory", Op: "Upload", Err: err}
	}

	key := storeKey(options.Bucket, options.FileName)

	ms.mu.Lock()
	ms.store[key] = buf.Bytes()
	ms.mu.Unlock()

	folder := "memory"
	if options.Bucket != "" {
		folder = "memory/" + options.Bucket
	}
	return &GFileMux.UploadedFileMetadata{
		FolderDestination: folder,
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Get returns the raw bytes stored for the given bucket+key pair.
// Returns an error if the file was not found.
func (ms *MemoryStorage) Get(bucket, key string) ([]byte, error) {
	ms.mu.RLock()
	data, ok := ms.store[storeKey(bucket, key)]
	ms.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("file not found: %s", storeKey(bucket, key))
	}
	return data, nil
}

// Path returns a descriptive URI for the stored file (not a real filesystem path).
func (ms *MemoryStorage) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	return fmt.Sprintf("memory://%s/%s", options.Bucket, options.Key), nil
}

// Delete removes the stored file identified by bucket and key.
func (ms *MemoryStorage) Delete(ctx context.Context, bucket, key string) error {
	k := storeKey(bucket, key)
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, ok := ms.store[k]; !ok {
		return &GFileMux.StorageError{
			Backend: "memory",
			Op:      "Delete",
			Err:     fmt.Errorf("file not found: %s", k),
		}
	}
	delete(ms.store, k)
	return nil
}

// Close is a no-op for MemoryStorage.
func (ms *MemoryStorage) Close() error {
	return nil
}
