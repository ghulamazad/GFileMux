package GFileMux

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// FileValidatorFunc validates a File during upload, returning an error if the file is invalid.
type FileValidatorFunc func(f File) error

// UploadErrorHandlerFunc handles upload errors by returning an http.HandlerFunc
// that writes an appropriate response to the client.
type UploadErrorHandlerFunc func(err error) http.HandlerFunc

// FileNameGeneratorFunc generates a storage filename from the original filename.
type FileNameGeneratorFunc func(s string) string

var (
	// DefaultMaxFileUploadSize is the default maximum allowed file size (5 MB).
	DefaultMaxFileUploadSize int64 = 1024 * 1024 * 5

	// DefaultMaxFiles is the default maximum number of files per field (unlimited).
	DefaultMaxFiles int = 0

	// DefaultFileValidator accepts every file without validation.
	DefaultFileValidator FileValidatorFunc = func(file File) error {
		return nil
	}

	// DefaultFileNameGeneratorFunc generates a unique filename using a Unix timestamp prefix.
	DefaultFileNameGeneratorFunc FileNameGeneratorFunc = func(s string) string {
		return fmt.Sprintf("GFileMux-%d-%s", time.Now().Unix(), s)
	}

	// DefaultUploadErrorHandlerFunc returns a JSON error response for upload failures.
	DefaultUploadErrorHandlerFunc UploadErrorHandlerFunc = func(err error) http.HandlerFunc {
		return func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status":"error","message":"GFileMux: File upload failed","error":%q}`, err.Error())
		}
	}
)

// WithStorage sets the storage backend for the GFileMux instance.
func WithStorage(store Storage) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.storage = store
	}
}

// WithMaxFileSize sets the maximum allowed file size in bytes.
//
//	GFileMux.WithMaxFileSize(10 << 20) // 10 MB
func WithMaxFileSize(size int64) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.maxSize = size
	}
}

// WithMaxFiles limits the number of files accepted per form field. When set to
// 0 (the default), there is no limit.
//
//	GFileMux.WithMaxFiles(5) // at most 5 files per field
func WithMaxFiles(n int) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.maxFiles = n
	}
}

// WithFileValidatorFunc sets the file validation function.
//
//	GFileMux.WithFileValidatorFunc(GFileMux.ValidateMimeType("image/jpeg"))
func WithFileValidatorFunc(validator FileValidatorFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.fileValidator = validator
	}
}

// WithFileNameGeneratorFunc sets the function used to generate storage filenames.
//
//	GFileMux.WithFileNameGeneratorFunc(func(orig string) string {
//	    return uuid.NewString() + filepath.Ext(orig)
//	})
func WithFileNameGeneratorFunc(generator FileNameGeneratorFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.fileNameGenerator = generator
	}
}

// WithIgnoreNonExistentKey controls whether missing form fields cause an error.
// When true, fields not present in the multipart form are silently skipped.
func WithIgnoreNonExistentKey(ignore bool) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.ignoreNonExistentKeys = ignore
	}
}

// WithUploadErrorHandlerFunc sets a custom error response handler for upload failures.
//
//	GFileMux.WithUploadErrorHandlerFunc(func(err error) http.HandlerFunc {
//	    return func(w http.ResponseWriter, r *http.Request) {
//	        http.Error(w, err.Error(), http.StatusBadRequest)
//	    }
//	})
func WithUploadErrorHandlerFunc(handler UploadErrorHandlerFunc) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.uploadErrorHandler = handler
	}
}

// WithAllowedBuckets restricts which bucket names may be used with this handler.
// Passing a bucket not in this list causes the Upload middleware to return an error.
// If no buckets are configured, all bucket names are accepted.
//
//	GFileMux.WithAllowedBuckets("avatars", "documents")
func WithAllowedBuckets(buckets ...string) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.allowedBuckets = append(cfg.allowedBuckets, buckets...)
	}
}

// WithLogger attaches a structured logger that GFileMux will use to emit
// lifecycle events (upload started, completed, failed). Pass nil to disable logging.
//
//	GFileMux.WithLogger(slog.Default())
func WithLogger(logger *slog.Logger) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.logger = logger
	}
}

// WithChecksumValidation enables SHA-256 checksum computation for every uploaded
// file. The hex digest is stored in File.ChecksumSHA256 and available in handlers.
func WithChecksumValidation(enable bool) GFileMuxOption {
	return func(cfg *GFileMux) {
		cfg.computeChecksum = enable
	}
}

// WithBucket sets the bucket option for UploadOptions.
func WithBucket(bucket string) Option {
	return func(o *UploadOptions) {
		o.Bucket = bucket
	}
}

// WithKeys sets the keys option for UploadOptions.
func WithKeys(keys ...string) Option {
	return func(o *UploadOptions) {
		o.Keys = keys
	}
}
