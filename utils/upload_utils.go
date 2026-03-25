package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// ReaderToSeeker copies r into a temporary file and returns it seeked to the start.
// The caller is responsible for closing the returned file.
func ReaderToSeeker(r io.Reader) (io.ReadSeeker, error) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "upload-")
	if err != nil {
		return nil, err
	}

	// Ensure the temporary file is cleaned up if an error occurs or when done
	defer func() {
		if err != nil {
			_ = tmpfile.Close()
			_ = os.Remove(tmpfile.Name())
		}
	}()

	// Copy the content of the reader into the temporary file
	_, err = io.Copy(tmpfile, r)
	if err != nil {
		return nil, err
	}

	// Seek to the beginning of the file
	_, err = tmpfile.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	// Return the temporary file as a ReadSeeker
	return tmpfile, nil
}

// ComputeSHA256 reads from rs, computes its SHA-256 digest, seeks back to the
// start, and returns the hex-encoded hash string. It does not consume the reader
// permanently — the seeker is reset so the same data can be uploaded afterward.
func ComputeSHA256(rs io.ReadSeeker) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, rs); err != nil {
		return "", err
	}
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

