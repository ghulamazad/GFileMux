package GFileMux

import (
	"testing"
)

func TestValidateMimeType_Allowed(t *testing.T) {
	validator := ValidateMimeType("image/jpeg", "image/png")
	file := File{FieldName: "photo", MimeType: "image/jpeg"}
	if err := validator(file); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestValidateMimeType_CaseInsensitive(t *testing.T) {
	validator := ValidateMimeType("Image/JPEG")
	file := File{FieldName: "photo", MimeType: "image/jpeg"}
	if err := validator(file); err != nil {
		t.Fatalf("expected nil for case-insensitive match, got %v", err)
	}
}

func TestValidateMimeType_Rejected(t *testing.T) {
	validator := ValidateMimeType("image/jpeg", "image/png")
	file := File{FieldName: "doc", MimeType: "application/pdf"}
	err := validator(file)
	if err == nil {
		t.Fatal("expected error for disallowed MIME type")
	}
	var ve *ValidationError
	if !isValidationError(err, &ve) {
		t.Fatalf("expected *ValidationError, got %T: %v", err, err)
	}
}

func TestValidateFileExtension_Allowed(t *testing.T) {
	validator := ValidateFileExtension(".jpg", ".png")
	file := File{FieldName: "photo", OriginalName: "photo.JPG"}
	if err := validator(file); err != nil {
		t.Fatalf("expected nil for .JPG (case-insensitive), got %v", err)
	}
}

func TestValidateFileExtension_Rejected(t *testing.T) {
	validator := ValidateFileExtension(".jpg", ".png")
	file := File{FieldName: "doc", OriginalName: "report.pdf"}
	if err := validator(file); err == nil {
		t.Fatal("expected error for .pdf extension")
	}
}

func TestValidateMinFileSize_Allowed(t *testing.T) {
	validator := ValidateMinFileSize(100)
	file := File{FieldName: "doc", Size: 200}
	if err := validator(file); err != nil {
		t.Fatalf("expected nil for size 200 >= 100, got %v", err)
	}
}

func TestValidateMinFileSize_Rejected(t *testing.T) {
	validator := ValidateMinFileSize(1024)
	file := File{FieldName: "doc", Size: 10}
	if err := validator(file); err == nil {
		t.Fatal("expected error for size 10 < 1024")
	}
}

func TestChainValidators_AllPass(t *testing.T) {
	chain := ChainValidators(
		ValidateMimeType("image/jpeg"),
		ValidateMinFileSize(1),
	)
	file := File{FieldName: "img", MimeType: "image/jpeg", Size: 100}
	if err := chain(file); err != nil {
		t.Fatalf("expected nil from chain, got %v", err)
	}
}

func TestChainValidators_FirstFails(t *testing.T) {
	chain := ChainValidators(
		ValidateMimeType("image/jpeg"),
		ValidateMinFileSize(1),
	)
	file := File{FieldName: "img", MimeType: "application/pdf", Size: 100}
	if err := chain(file); err == nil {
		t.Fatal("expected error when first validator fails")
	}
}

// isValidationError is a helper to check the error type without importing errors in tests.
func isValidationError(err error, target **ValidationError) bool {
	if ve, ok := err.(*ValidationError); ok {
		*target = ve
		return true
	}
	return false
}
