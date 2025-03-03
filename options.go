package GFileMux

import (
	"fmt"
	"net/http"
	"time"
)

// FileValidator is a type that represents a function used to validate a file during upload.
// It receives a `File` object and returns an error if the file is invalid.
type FileValidatorFunc func(f File) error

// UploadErrorHandler is a custom function type used to handle errors when
// an upload fails. It takes an error and returns an HTTP handler function to
// send the appropriate error response.
type UploadErrorHandlerFunc func(err error) http.HandlerFunc

// FileNameGenerator is a function type that allows you to alter the name of the file
// before it is uploaded and stored. This is useful when you need to ensure filenames
// adhere to a specific format or naming convention.
type FileNameGeneratorFunc func(s string) string

var (
	// DefaultMaxFileUploadSize is the default maximum allowed file size for uploads (5MB).
	DefaultMaxFileUploadSize int64 = 1024 * 1024 * 5

	// DefaultFileValidator allows all files to pass through without validation.
	DefaultFileValidator FileValidatorFunc = func(file File) error {
		return nil
	}

	// DefaultFileNameGeneratorFunc generates a unique file name based on a timestamp and the original file name.
	DefaultFileNameGeneratorFunc FileNameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("GFileMux-%d-%s", time.Now().Unix(), s)
	}

	// DefaultUploadErrorHandlerFunc provides a default error handler for upload failures.
	// It sends a JSON response with a generic error message and the underlying error.
	DefaultUploadErrorHandlerFunc UploadErrorHandlerFunc = func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": "error", "message": "GFileMux: File upload failed", "error": "%s"}`, err.Error())
		}
	}
)

// WithStorage sets the storage backend for the GFileMux instance.
func WithStorage(store Storage) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.storage = store
	}
}

// WithMaxFileSize sets the maximum file size for uploads.
func WithMaxFileSize(size int64) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.maxSize = size
	}
}

// WithValidationFunc sets the file validation function.
func WithValidationFunc(validator FileValidatorFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.fileValidator = validator
	}
}

// WithNameFuncGenerator sets the function to generate file names.
func WithNameFuncGenerator(generator FileNameGeneratorFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.fileNameGenerator = generator
	}
}

// WithIgnoreNonExistentKey sets whether to ignore non-existent keys during file retrieval.
func WithIgnoreNonExistentKey(ignore bool) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.ignoreNonExistentKeys = ignore
	}
}

// WithErrorResponseHandler sets the error response handler for uploads.
func WithErrorResponseHandler(handler UploadErrorHandlerFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.uploadErrorHandler = handler
	}
}

// WithBucket sets the bucket option
func WithBucket(bucket string) Option {
	return func(o *UploadOptions) {
		o.Bucket = bucket
	}
}

// WithKeys sets the keys option
func WithKeys(keys ...string) Option {
	return func(o *UploadOptions) {
		o.Keys = keys
	}
}
