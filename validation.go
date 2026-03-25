package GFileMux

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidateMimeType returns a FileValidatorFunc that checks if a file's MIME type
// matches one of the allowed MIME types. The comparison is case-insensitive.
//
// Example:
//
//	GFileMux.ValidateMimeType("image/jpeg", "image/png")
func ValidateMimeType(validMimeTypes ...string) FileValidatorFunc {
	// Normalise allowed types once at construction time.
	lower := make([]string, len(validMimeTypes))
	for i, m := range validMimeTypes {
		lower[i] = strings.ToLower(strings.TrimSpace(m))
	}

	return func(file File) error {
		fileMime := strings.ToLower(strings.TrimSpace(file.MimeType))
		for _, allowed := range lower {
			if fileMime == allowed {
				return nil
			}
		}
		return &ValidationError{
			Field:   file.FieldName,
			Message: fmt.Sprintf("unsupported MIME type %q; allowed: %s", file.MimeType, strings.Join(validMimeTypes, ", ")),
		}
	}
}

// ValidateFileExtension returns a FileValidatorFunc that checks whether the
// uploaded file's original name has one of the allowed extensions.
// Extensions are matched case-insensitively and must include the leading dot
// (e.g. ".pdf", ".jpg").
//
// Example:
//
//	GFileMux.ValidateFileExtension(".jpg", ".jpeg", ".png")
func ValidateFileExtension(allowedExts ...string) FileValidatorFunc {
	lower := make([]string, len(allowedExts))
	for i, e := range allowedExts {
		lower[i] = strings.ToLower(strings.TrimSpace(e))
	}

	return func(file File) error {
		ext := strings.ToLower(filepath.Ext(file.OriginalName))
		for _, allowed := range lower {
			if ext == allowed {
				return nil
			}
		}
		return &ValidationError{
			Field:   file.FieldName,
			Message: fmt.Sprintf("file extension %q is not allowed; allowed: %s", ext, strings.Join(allowedExts, ", ")),
		}
	}
}

// ValidateMinFileSize returns a FileValidatorFunc that rejects files smaller
// than minBytes bytes. Useful for preventing zero-byte or near-empty uploads.
//
// Example:
//
//	GFileMux.ValidateMinFileSize(1024) // at least 1 KB
func ValidateMinFileSize(minBytes int64) FileValidatorFunc {
	return func(file File) error {
		if file.Size < minBytes {
			return &ValidationError{
				Field:   file.FieldName,
				Message: fmt.Sprintf("file is too small: got %d bytes, minimum is %d bytes", file.Size, minBytes),
			}
		}
		return nil
	}
}

// ChainValidators returns a FileValidatorFunc that applies multiple validation
// functions sequentially. The first error encountered is immediately returned.
//
// Example:
//
//	GFileMux.ChainValidators(
//	    GFileMux.ValidateMimeType("image/jpeg"),
//	    GFileMux.ValidateFileExtension(".jpg"),
//	    GFileMux.ValidateMinFileSize(1024),
//	)
func ChainValidators(validators ...FileValidatorFunc) FileValidatorFunc {
	return func(file File) error {
		for _, v := range validators {
			if err := v(file); err != nil {
				return err
			}
		}
		return nil
	}
}
