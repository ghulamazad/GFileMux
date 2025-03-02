package utils

import (
	"io"
	"os"
)

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
