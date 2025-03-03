package GFileMux

import (
	"context"
	"io"
	"time"
)

// UploadFileOptions holds the configuration for uploading a file.
type UploadFileOptions struct {
	FileName string            `json:"file_name,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`

	// Bucket specifies the storage bucket to upload the file to.
	// If not provided, the default bucket will be used.
	Bucket string `json:"bucket,omitempty"`
}

// UploadedFileMetadata contains metadata about a file after it has been uploaded.
type UploadedFileMetadata struct {
	FolderDestination string `json:"folder_destination,omitempty"`
	Key               string `json:"key,omitempty"`
	Size              int64  `json:"size,omitempty"`
}

// PathOptions holds options for generating the file's path.
type PathOptions struct {
	Bucket string `json:"bucket,omitempty"`

	Key string `json:"key,omitempty"`

	ExpirationTime time.Duration `json:"expiration_time,omitempty"`

	// IsSecure indicates if the path should be secured and time-limited.
	IsSecure bool `json:"is_secure,omitempty"`
}

// Storage defines the interface for interacting with file storage systems.
type Storage interface {
	// Upload uploads a file from the provided reader and returns metadata about the uploaded file.
	Upload(ctx context.Context, reader io.Reader, options *UploadFileOptions) (*UploadedFileMetadata, error)

	// Path generates a path for the given file based on the provided options.
	Path(ctx context.Context, options PathOptions) (string, error)

	// Closer interface to close any resources after use.
	io.Closer
}
