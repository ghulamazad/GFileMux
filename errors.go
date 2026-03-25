package GFileMux

import "fmt"

// ValidationError is returned when a file fails validation (e.g. wrong MIME type, extension, or size).
// Callers can detect this with errors.As to distinguish validation failures from infrastructure errors.
type ValidationError struct {
	Field   string // form field name
	Message string // human-readable reason
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("GFileMux: validation error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("GFileMux: validation error: %s", e.Message)
}

// SizeError is returned when an uploaded file exceeds the configured size limit.
type SizeError struct {
	Field   string
	Size    int64 // actual size in bytes
	MaxSize int64 // configured limit in bytes
}

func (e *SizeError) Error() string {
	return fmt.Sprintf(
		"GFileMux: file in field %q is too large: got %d bytes, max allowed is %d bytes",
		e.Field, e.Size, e.MaxSize,
	)
}

// MaxFilesError is returned when the number of files in a field exceeds WithMaxFiles.
type MaxFilesError struct {
	Field    string
	Got      int
	MaxFiles int
}

func (e *MaxFilesError) Error() string {
	return fmt.Sprintf(
		"GFileMux: too many files in field %q: got %d, max allowed is %d",
		e.Field, e.Got, e.MaxFiles,
	)
}

// StorageError wraps errors that originate from a storage backend.
type StorageError struct {
	Backend string // e.g. "disk", "memory", "s3"
	Op      string // e.g. "Upload", "Delete", "Path"
	Err     error
}

func (e *StorageError) Error() string {
	return fmt.Sprintf("GFileMux: %s storage error during %s: %v", e.Backend, e.Op, e.Err)
}

func (e *StorageError) Unwrap() error {
	return e.Err
}
