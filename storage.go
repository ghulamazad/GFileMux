package GFileMux

import (
	"context"
	"io"
	"time"
)

type UploadFileOptions struct {
	FileName string            `json:"file_name,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`

	// If not provided, the default bucket will be used
	Bucket string `json:"bucket,omitempty"`
}

type UploadedFileMetadata struct {
	FolderDestination string `json:"folder_destination,omitempty"`
	Key               string `json:"key,omitempty"`
	Size              int64  `json:"size,omitempty"`
}

type PathOptions struct {
	Bucket         string        `json:"bucket,omitempty"`
	Key            string        `json:"key,omitempty"`
	ExpirationTime time.Duration `json:"expiration_time,omitempty"` // Only effective if IsSecure is true
	IsSecure       bool          `json:"is_secure,omitempty"`
}

type Storage interface {
	Upload(context.Context, io.Reader, *UploadFileOptions) (*UploadedFileMetadata, error)
	Path(context.Context, PathOptions) (string, error)
	io.Closer
}
