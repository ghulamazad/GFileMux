package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ghulamazad/GFileMux"
)

// DiskStorage saves uploaded files to the local filesystem.
// The optional Bucket parameter is used as a subdirectory under Directory,
// allowing logical separation of files (e.g. by tenant or file type).
type DiskStorage struct {
	Directory string
}

// NewDiskStorage initializes a new DiskStorage instance. If the directory does
// not exist it is created automatically (including any parent directories).
func NewDiskStorage(directory string) (*DiskStorage, error) {
	directory = strings.TrimSpace(directory)
	if directory == "" {
		return nil, fmt.Errorf("directory path is empty or only whitespace")
	}

	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("could not create directory '%s': %v", directory, err)
	}

	return &DiskStorage{Directory: directory}, nil
}

// bucketDir returns the resolved directory for the given bucket, creating it
// when it does not already exist.
func (ds *DiskStorage) bucketDir(bucket string) (string, error) {
	dir := ds.Directory
	if bucket != "" {
		dir = filepath.Join(ds.Directory, filepath.Clean(bucket))
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("could not create bucket directory '%s': %v", dir, err)
	}
	return dir, nil
}

// Upload saves a file to disk. If a non-empty Bucket is provided in options it
// is used as a subdirectory under the root Directory.
func (ds *DiskStorage) Upload(ctx context.Context, reader io.Reader, options *GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	if options == nil || options.FileName == "" {
		return nil, fmt.Errorf("invalid upload options: file name is required")
	}

	dir, err := ds.bucketDir(options.Bucket)
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(dir, options.FileName)
	file, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("could not create file '%s': %v", destPath, err)
	}
	defer file.Close()

	n, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy data to file '%s': %v", destPath, err)
	}

	return &GFileMux.UploadedFileMetadata{
		FolderDestination: dir,
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Path returns the full filesystem path of a stored file.
func (ds *DiskStorage) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	if options.Key == "" {
		return "", fmt.Errorf("invalid path options: key is required")
	}
	dir := ds.Directory
	if options.Bucket != "" {
		dir = filepath.Join(ds.Directory, filepath.Clean(options.Bucket))
	}
	return filepath.Join(dir, options.Key), nil
}

// Delete removes the file identified by key from the given bucket.
func (ds *DiskStorage) Delete(ctx context.Context, bucket, key string) error {
	if key == "" {
		return fmt.Errorf("key is required")
	}
	dir := ds.Directory
	if bucket != "" {
		dir = filepath.Join(ds.Directory, filepath.Clean(bucket))
	}
	path := filepath.Join(dir, key)
	if err := os.Remove(path); err != nil {
		return &GFileMux.StorageError{Backend: "disk", Op: "Delete", Err: err}
	}
	return nil
}

// Close is a no-op for DiskStorage but satisfies the Storage interface.
func (ds *DiskStorage) Close() error {
	return nil
}
