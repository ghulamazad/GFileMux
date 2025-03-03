package GFileMux

import (
	"fmt"
	"strings"
)

// ValidateMimeType returns a FileValidatorFunc that checks if a file's MIME type
// matches one of the allowed MIME types. The MIME types to be validated against
// are provided as a variadic argument. The comparison is case-insensitive.
//
// It returns nil if the file's MIME type is valid, otherwise it returns an error
// with the message indicating the unsupported MIME type.
func ValidateMimeType(validMimeTypes ...string) FileValidatorFunc {
	return func(file File) error {
		for _, mimeType := range validMimeTypes {
			if strings.EqualFold(strings.ToLower(mimeType), file.MimeType) {
				return nil
			}
		}

		return fmt.Errorf("unsupported MIME type uploaded: %s", file.MimeType)
	}
}

// ChainValidators returns a FileValidatorFunc that applies multiple validation functions
// sequentially. Each validator function is applied in the order it is provided. If
// any validator returns an error, that error is immediately returned, and the remaining
// validators are not executed.
//
// This function is useful for combining different validation rules into a single validation
// process.
func ChainValidators(validators ...FileValidatorFunc) FileValidatorFunc {
	return func(file File) error {
		for _, validator := range validators {
			if err := validator(file); err != nil {
				return err
			}
		}
		return nil
	}
}
