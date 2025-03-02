package utils

import (
	"io"
	"net/http"
	"strings"
)

// FetchContentType detects the MIME type of a file based on its first 512 bytes.
// It reads the initial portion of the file to determine its type, resets the file
// pointer back to the beginning after detection, and returns the MIME type without
// any charset information (e.g., "text/plain" instead of "text/plain; charset=utf-8").
//
// Parameters:
//
//	f (io.ReadSeeker): The file or stream from which content is read. It must support
//	both reading and seeking.
//
// Returns:
//   - A string containing the MIME type (e.g., "text/plain", "image/jpeg").
//   - An error if there is an issue with reading or seeking the file.
func FetchContentType(f io.ReadSeeker) (string, error) {
	// Allocate a buffer to read the first 512 bytes
	buffer := make([]byte, 512)

	// Seek to the beginning of the file
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Read the first 512 bytes
	bytesRead, err := f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Trim the buffer to the actual number of bytes read
	buffer = buffer[:bytesRead]

	// Detect the MIME type based on the first few bytes
	contentType := http.DetectContentType(buffer)

	// Reset the file pointer to the beginning after detection
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Handle potential charset in the MIME type, e.g., "text/plain; charset=utf-8"
	if mimeParts := strings.Split(contentType, ";"); len(mimeParts) > 1 {
		contentType = mimeParts[0] // Keep only the MIME type, not the charset
	}

	return contentType, nil
}
