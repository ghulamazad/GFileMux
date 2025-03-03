package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gfilemux "github.com/ghulamazad/GFileMux"
)

type DiskStorage struct {
	Directory string
}

func NewDiskStorage(directory string) (*DiskStorage, error) {
	if len(strings.TrimSpace(directory)) == 0 {
		return nil, fmt.Errorf("directory path is empty or only whitespace: %s", directory)
	}

	if _, err := os.Stat(directory); err != nil {
		return nil, fmt.Errorf("could not access directory '%s': %v", directory, err)
	}

	return &DiskStorage{Directory: directory}, nil
}

func (ds *DiskStorage) Upload(ctx context.Context, reader io.Reader, options *gfilemux.UploadFileOptions) (*gfilemux.UploadedFileMetadata, error) {
	file, err := os.Create(filepath.Join(ds.Directory, options.FileName))
	if err != nil {
		return nil, err
	}

	defer file.Close()

	n, err := io.Copy(file, reader)
	if err != nil {
		return nil, err
	}

	return &gfilemux.UploadedFileMetadata{
		FolderDestination: ds.Directory,
		Size:              n,
		Key:               options.FileName,
	}, err
}

func (ds *DiskStorage) Path(ctx context.Context, options gfilemux.PathOptions) (string, error) {
	return fmt.Sprintf("%s/%s", ds.Directory, options.Key), nil
}

func (ds *DiskStorage) Close() error {
	return nil
}
