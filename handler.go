package GFileMux

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/ghulamazad/GFileMux/utils"
	"golang.org/x/sync/errgroup"
)

// GFileMux is the core handler struct holding all upload configuration.
type GFileMux struct {
	// storage defines where uploaded files are persisted.
	storage Storage

	// maxSize is the maximum allowed size for the entire multipart body in bytes.
	maxSize int64

	// maxFiles is the maximum number of files allowed per form field. 0 = unlimited.
	maxFiles int

	// ignoreNonExistentKeys, when true, silently skips form fields that are absent.
	ignoreNonExistentKeys bool

	// allowedBuckets is an optional whitelist of permitted bucket names.
	// An empty slice means all buckets are allowed.
	allowedBuckets []string

	// computeChecksum controls whether SHA-256 is computed for each uploaded file.
	computeChecksum bool

	// fileValidator validates each file before it is stored.
	fileValidator FileValidatorFunc

	// fileNameGenerator generates a storage filename from the original name.
	fileNameGenerator FileNameGeneratorFunc

	// uploadErrorHandler builds the HTTP error response for upload failures.
	uploadErrorHandler UploadErrorHandlerFunc

	// logger is an optional structured logger. nil means no logging.
	logger *slog.Logger
}

// GFileMuxOption is a function that configures a GFileMux instance.
type GFileMuxOption func(*GFileMux)

// New creates a new GFileMux handler with the supplied options.
// A storage backend must be provided via WithStorage; all other options are optional.
func New(options ...GFileMuxOption) (*GFileMux, error) {
	handler := &GFileMux{}

	for _, opt := range options {
		opt(handler)
	}

	if handler.maxSize <= 0 {
		handler.maxSize = DefaultMaxFileUploadSize
	}
	if handler.fileValidator == nil {
		handler.fileValidator = DefaultFileValidator
	}
	if handler.fileNameGenerator == nil {
		handler.fileNameGenerator = DefaultFileNameGeneratorFunc
	}
	if handler.uploadErrorHandler == nil {
		handler.uploadErrorHandler = DefaultUploadErrorHandlerFunc
	}
	if handler.storage == nil {
		return nil, errors.New("a storage backend must be provided via WithStorage")
	}

	return handler, nil
}

// Storage returns the configured storage backend.
func (gfm *GFileMux) Storage() Storage {
	return gfm.storage
}

// isBucketAllowed returns true when the bucket is in the allowedBuckets list,
// or when no whitelist has been configured.
func (gfm *GFileMux) isBucketAllowed(bucket string) bool {
	if len(gfm.allowedBuckets) == 0 {
		return true
	}
	for _, b := range gfm.allowedBuckets {
		if b == bucket {
			return true
		}
	}
	return false
}

// log emits a structured log line when a logger is configured.
func (gfm *GFileMux) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if gfm.logger != nil {
		gfm.logger.Log(ctx, level, msg, args...)
	}
}

// UploadOptions struct encapsulates per-call upload options.
type UploadOptions struct {
	Bucket string
	Keys   []string
}

// Option configures an UploadOptions value.
type Option func(*UploadOptions)

// Upload returns an HTTP middleware that parses a multipart form, uploads the
// files found under each of the provided keys to the configured storage backend,
// and stores their metadata in the request context for use by the next handler.
//
// The race condition that previously existed (concurrent writes to a plain map)
// is eliminated here by using sync.Map: each goroutine writes exclusively to its
// own key, so there is zero lock contention while still being race-detector-clean.
func (gfm *GFileMux) Upload(bucket string, keys ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Guard: validate bucket against allowedBuckets whitelist.
			if !gfm.isBucketAllowed(bucket) {
				gfm.uploadErrorHandler(fmt.Errorf("bucket %q is not allowed", bucket)).ServeHTTP(w, r)
				return
			}

			// Enforce total body size limit before parsing.
			r.Body = http.MaxBytesReader(w, r.Body, gfm.maxSize)
			if err := r.ParseMultipartForm(gfm.maxSize); err != nil {
				if strings.Contains(err.Error(), "request body too large") {
					gfm.uploadErrorHandler(&SizeError{Size: gfm.maxSize, MaxSize: gfm.maxSize}).ServeHTTP(w, r)
					return
				}
				gfm.uploadErrorHandler(err).ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			gfm.log(ctx, slog.LevelInfo, "upload started", "bucket", bucket, "fields", keys)

			// Use sync.Map so each goroutine can write its own key concurrently
			// without any mutex — zero contention, race-detector clean.
			var sm sync.Map
			var wg errgroup.Group

			for _, key := range keys {
				key := key // capture for closure

				wg.Go(func() error {
					fileHeaders, ok := r.MultipartForm.File[key]
					if !ok {
						if gfm.ignoreNonExistentKeys {
							return nil
						}
						return fmt.Errorf("no files found for field %q in the request", key)
					}

					// Enforce per-field file count limit.
					if gfm.maxFiles > 0 && len(fileHeaders) > gfm.maxFiles {
						return &MaxFilesError{Field: key, Got: len(fileHeaders), MaxFiles: gfm.maxFiles}
					}

					localFiles := make([]File, 0, len(fileHeaders))

					for _, header := range fileHeaders {
						f, err := header.Open()
						if err != nil {
							return fmt.Errorf("could not open file for field %q: %w", key, err)
						}
						defer f.Close()

						uploadedFileName := gfm.fileNameGenerator(header.Filename)

						// Detect MIME type from the first 512 bytes.
						mimeType, err := utils.FetchContentType(f)
						if err != nil {
							return fmt.Errorf("could not detect MIME type for field %q: %w", key, err)
						}

						fileData := File{
							FieldName:        key,
							OriginalName:     header.Filename,
							UploadedFileName: uploadedFileName,
							MimeType:         mimeType,
							Size:             header.Size,
						}

						// Run user-configured validators before touching storage.
						if err := gfm.fileValidator(fileData); err != nil {
							return fmt.Errorf("validation failed for field %q: %w", key, err)
						}

						// Optionally compute SHA-256 before upload (reader is seeked back afterward).
						if gfm.computeChecksum {
							checksum, err := utils.ComputeSHA256(f)
							if err != nil {
								return fmt.Errorf("could not compute checksum for field %q: %w", key, err)
							}
							fileData.ChecksumSHA256 = checksum
						}

						// Upload to the configured storage backend.
						metadata, err := gfm.storage.Upload(ctx, f, &UploadFileOptions{
							FileName: uploadedFileName,
							Bucket:   bucket,
						})
						if err != nil {
							return fmt.Errorf("storage upload failed for field %q: %w", key, err)
						}

						fileData.Size = metadata.Size
						fileData.FolderDestination = metadata.FolderDestination
						fileData.StorageKey = metadata.Key

						localFiles = append(localFiles, fileData)
					}

					// Each goroutine owns one unique key — zero contention with sync.Map.
					sm.Store(key, localFiles)
					return nil
				})
			}

			if err := wg.Wait(); err != nil {
				gfm.log(ctx, slog.LevelError, "upload failed", "error", err)
				gfm.uploadErrorHandler(err).ServeHTTP(w, r)
				return
			}

			// Collect results from sync.Map back into a plain Files map (single-threaded).
			uploadedFiles := make(Files, len(keys))
			sm.Range(func(k, v any) bool {
				uploadedFiles[k.(string)] = v.([]File)
				return true
			})

			gfm.log(ctx, slog.LevelInfo, "upload completed",
				"bucket", bucket,
				"total_files", uploadedFiles.Count(),
			)

			r = r.WithContext(addFilesToContext(r.Context(), uploadedFiles))
			next.ServeHTTP(w, r)
		})
	}
}

// UploadSingle is a convenience wrapper around Upload that enforces exactly one
// file for the given field. If the request contains more than one file for that
// field, the middleware returns an error before touching storage.
func (gfm *GFileMux) UploadSingle(bucket, key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		inner := gfm.Upload(bucket, key)
		return inner(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			files, err := GetUploadedFilesFromContext(r)
			if err != nil {
				gfm.uploadErrorHandler(err).ServeHTTP(w, r)
				return
			}
			if len(files[key]) > 1 {
				gfm.uploadErrorHandler(
					&MaxFilesError{Field: key, Got: len(files[key]), MaxFiles: 1},
				).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
