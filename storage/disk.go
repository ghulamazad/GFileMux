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

type DiskStorage struct {
	Directory string
}

// NewDiskStorage initializes a new DiskStorage instance with the provided directory.
func NewDiskStorage(directory string) (*DiskStorage, error) {
	// Trim any leading/trailing whitespace and check if the directory path is empty.
	directory = strings.TrimSpace(directory)
	if directory == "" {
		return nil, fmt.Errorf("directory path is empty or only whitespace")
	}

	// Check if the provided directory exists.
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory '%s' does not exist", directory)
		}
		return nil, fmt.Errorf("could not access directory '%s': %v", directory, err)
	}

	return &DiskStorage{Directory: directory}, nil
}

// Upload saves a file to the disk from the reader with the provided options.
func (ds *DiskStorage) Upload(ctx context.Context, reader io.Reader, options *GFileMux.UploadFileOptions) (*GFileMux.UploadedFileMetadata, error) {
	// Ensure the options are valid.
	if options == nil || options.FileName == "" {
		return nil, fmt.Errorf("invalid upload options: file name is required")
	}

	// Create the destination file.
	destPath := filepath.Join(ds.Directory, options.FileName)
	file, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("could not create file '%s': %v", destPath, err)
	}
	defer file.Close()

	// Copy the contents of the reader to the file.
	n, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy data to file '%s': %v", destPath, err)
	}

	// Return the metadata of the uploaded file.
	return &GFileMux.UploadedFileMetadata{
		FolderDestination: ds.Directory,
		Size:              n,
		Key:               options.FileName,
	}, nil
}

// Path returns the full path of the file with the given options.
func (ds *DiskStorage) Path(ctx context.Context, options GFileMux.PathOptions) (string, error) {
	if options.Key == "" {
		return "", fmt.Errorf("invalid path options: key is required")
	}
	return filepath.Join(ds.Directory, options.Key), nil
}

// Close performs any necessary cleanup (currently a no-op for DiskStorage).
func (ds *DiskStorage) Close() error {
	// No resources to clean up in this implementation, but the method is still available for future use.
	return nil
}
