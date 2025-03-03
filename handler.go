package GFileMux

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ghulamazad/GFileMux/utils"
	"golang.org/x/sync/errgroup"
)

// Config represents the configuration for the file upload handler.
type GFileMux struct {
	// storage defines where the uploaded files are stored (e.g., local disk, cloud storage).
	storage Storage

	// maxSize sets the maximum allowed size for uploaded files in bytes.
	maxSize int64

	// ignoreNonExistentKeys, when set to true, allows the handler to skip missing keys
	// during file retrieval instead of failing the request.
	ignoreNonExistentKeys bool

	// fileValidator is a function used to validate uploaded files (e.g., file type, size).
	fileValidator FileValidatorFunc

	// fileNameGenerator generates unique names for uploaded files.
	fileNameGenerator FileNameGeneratorFunc

	// uploadErrorHandler handles errors that occur during file upload, typically by
	// customizing the response returned to the client.
	uploadErrorHandler UploadErrorHandlerFunc
}

// GFileMuxOption is a function type that configures the GFileMux instance.
type GFileMuxOption func(*GFileMux)

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
		return nil, errors.New("a storage backend must be provided")
	}

	return handler, nil
}

func (gfm *GFileMux) Storage() Storage {
	return gfm.storage
}

// UploadOptions struct to encapsulate the options
type UploadOptions struct {
	Bucket string
	Keys   []string
}

// Option is a function that configures an UploadOptions
type Option func(*UploadOptions)

// Upload is a HTTP middleware that takes in a list of form fields and the next
// HTTP handler to run after the upload process is completed
func (gfm *GFileMux) Upload(bucket string, keys ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, gfm.maxSize)

			err := r.ParseMultipartForm(gfm.maxSize)
			if err != nil {
				if strings.Contains(err.Error(), "request body too large") {
					gfm.uploadErrorHandler(fmt.Errorf("file size exceeded the limit of %d bytes", gfm.maxSize)).ServeHTTP(w, r)
					return
				}
				gfm.uploadErrorHandler(err).ServeHTTP(w, r)
				return
			}

			ctx, cancel := context.WithCancel(r.Context())
			defer cancel()

			// Create an errgroup with context propagation
			var wg errgroup.Group

			uploadedFiles := make(Files, len(keys))

			// Iterate over each key and process the uploaded files
			for _, key := range keys {
				key := key // capture key for closure

				wg.Go(func() error {
					fileHeaders, ok := r.MultipartForm.File[key]
					if !ok {
						if gfm.ignoreNonExistentKeys {
							return nil
						}
						return fmt.Errorf("files could not be found in key (%s) from the HTTP request", key)
					}

					uploadedFiles[key] = make([]File, 0, len(fileHeaders))

					for _, header := range fileHeaders {
						// Open the file and handle the file metadata
						f, err := header.Open()
						if err != nil {
							return fmt.Errorf("could not open file for key (%s): %v", key, err)
						}
						defer f.Close()

						uploadedFileName := gfm.fileNameGenerator(header.Filename)

						// Fetch MIME type of the uploaded file
						mimeType, err := utils.FetchContentType(f)
						if err != nil {
							return fmt.Errorf("%s has an invalid MIME type: %v", key, err)
						}

						fileSize := header.Size

						// Create a file data struct
						fileData := File{
							FieldName:        key,
							OriginalName:     header.Filename,
							UploadedFileName: uploadedFileName,
							MimeType:         mimeType,
							Size:             fileSize,
						}

						// Validate file data
						if err := gfm.fileValidator(fileData); err != nil {
							return fmt.Errorf("validation failed for (%s): %v", key, err)
						}

						// Upload file to storage
						metadata, err := gfm.storage.Upload(ctx, f, &UploadFileOptions{
							FileName: uploadedFileName,
							Bucket:   bucket,
						})
						if err != nil {
							return fmt.Errorf("could not upload file to storage (%s): %v", key, err)
						}

						// Add metadata to file data
						fileData.Size = metadata.Size
						fileData.FolderDestination = metadata.FolderDestination
						fileData.StorageKey = metadata.Key

						// Append file data to uploaded files map
						uploadedFiles[key] = append(uploadedFiles[key], fileData)
					}

					return nil
				})
			}

			// Wait for all file upload operations to finish
			if err := wg.Wait(); err != nil {
				gfm.uploadErrorHandler(err).ServeHTTP(w, r)
				return
			}

			// Write uploaded files to request context
			r = r.WithContext(addFilesToContext(r.Context(), uploadedFiles))

			// Pass the request to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
