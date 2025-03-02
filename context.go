package GFileMux

import (
	"context"
	"net/http"

	GFileMuxErrors "github.com/ghulamazad/GFileMux/internal/errors"
)

// fileContextKey is the key type used to store files in context.
type fileContextKey string

// Define the key used for storing files in context.
const fileKey fileContextKey = "files"

type Files map[string][]File

// addFilesToContext stores the provided files in the context under the key `fileKey`.
// If files already exist in the context, the new ones are appended.
func addFilesToContext(ctx context.Context, files Files) context.Context {
	// Retrieve the existing files from the context, if any.
	existingFiles, _ := ctx.Value(fileKey).(Files)
	if existingFiles == nil {
		// If no existing files, initialize an empty map.
		existingFiles = make(Files)
	}

	// Iterate over the provided files and append them to the corresponding field names.
	for _, fileSlice := range files {
		if len(fileSlice) > 0 {
			fieldName := fileSlice[0].FieldName
			existingFiles[fieldName] = append(existingFiles[fieldName], fileSlice...)
		}
	}

	// Return a new context with the updated files.
	return context.WithValue(ctx, fileKey, existingFiles)
}

// getFilesFromContext retrieves the files stored in the context, or returns an empty map if none exist.
func getFilesFromContext(ctx context.Context) Files {
	if files, ok := ctx.Value(fileKey).(Files); ok {
		return files
	}
	return Files{}
}

// GetUploadedFilesFromContext retrieves all uploaded files from the request's context.
func GetUploadedFilesFromContext(r *http.Request) (Files, error) {
	files := getFilesFromContext(r.Context())
	if len(files) == 0 {
		return nil, GFileMuxErrors.ErrFileNotUploaded
	}
	return files, nil
}

// GetFilesByFieldFromContext retrieves files uploaded under a specific form field (key).
func GetFilesByFieldFromContext(r *http.Request, key string) ([]File, error) {
	files := getFilesFromContext(r.Context())
	if len(files) == 0 {
		return nil, GFileMuxErrors.ErrFieldFilesMissing
	}
	return files[key], nil
}
