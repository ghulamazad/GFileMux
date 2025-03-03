package errors

import (
	"errors"
	"fmt"
)

var (
	ErrStorageFailure    = errors.New("failed to store the file")
	ErrFileNotUploaded   = errors.New("no files were uploaded in the request")
	ErrFieldFilesMissing = errors.New("no files found for the specified field key")
	ErrStorageRequired   = errors.New("a storage backend must be provided")
	ErrBucketRequired    = errors.New("please provide a valid S3 bucket")
)

func ErrUnsupportedMimeType(mimeType string) error {
	return fmt.Errorf("unsupported MIME type uploaded: %s", mimeType)
}

func ErrFileSizeExceeded(limit int64) error {
	return fmt.Errorf("file size exceeded the limit of %d bytes", limit)
}

func ErrFilesNotFoundInKey(key string) error {
	return fmt.Errorf("files could not be found in key (%s) from the HTTP request", key)
}

func ErrCouldNotOpenFile(key string, err error) error {
	return fmt.Errorf("could not open file for key (%s): %v", key, err)
}

func ErrInvalidMimeType(key string, err error) error {
	return fmt.Errorf("%s has an invalid MIME type: %v", key, err)
}

func ErrValidationFailed(key string, err error) error {
	return fmt.Errorf("validation failed for (%s): %v", key, err)
}

func ErrCouldNotUploadFile(key string, err error) error {
	return fmt.Errorf("could not upload file to storage (%s): %v", key, err)
}

func ErrCouldNotGetBucketLocation(err error) error {
	return fmt.Errorf("failed to get bucket location: %w", err)
}

func ErrCouldNotGeneratePresignedURL(err error) error {
	return fmt.Errorf("failed to generate presigned URL: %w", err)
}
