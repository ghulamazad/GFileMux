package GFileMux

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockStorage is a mock implementation of the Storage interface for testing.
type MockStorage struct {
	uploadedFiles map[string]*UploadedFileMetadata
}

func (ms *MockStorage) Upload(ctx context.Context, reader io.Reader, options *UploadFileOptions) (*UploadedFileMetadata, error) {
	if ms.uploadedFiles == nil {
		ms.uploadedFiles = make(map[string]*UploadedFileMetadata)
	}
	ms.uploadedFiles[options.FileName] = &UploadedFileMetadata{
		FolderDestination: options.Bucket,
		Size:              12345,
		Key:               options.FileName,
	}
	return ms.uploadedFiles[options.FileName], nil
}

func (ms *MockStorage) Path(ctx context.Context, options PathOptions) (string, error) {
	return "mock/path/" + options.Key, nil
}

func (ms *MockStorage) Close() error {
	return nil
}

func TestUpload(t *testing.T) {
	mockStorage := &MockStorage{}
	handler, err := New(
		WithStorage(mockStorage),
		WithMaxFileSize(10<<20),
		WithFileValidatorFunc(DefaultFileValidator),
		WithFileNameGeneratorFunc(DefaultFileNameGeneratorFunc),
		WithUploadErrorHandlerFunc(DefaultUploadErrorHandlerFunc),
	)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create a new HTTP request with a multipart form
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file1", "testfile.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write([]byte("This is a test file"))
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Create a test handler to verify the upload
	testHandler := handler.Upload("test_bucket", "file1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := GetUploadedFilesFromContext(r)
		if err != nil {
			t.Fatalf("Failed to get uploaded files from context: %v", err)
		}

		if len(files["file1"]) != 1 {
			t.Fatalf("Expected 1 file, got %d", len(files["file1"]))
		}

		if files["file1"][0].OriginalName != "testfile.txt" {
			t.Fatalf("Expected file name 'testfile.txt', got '%s'", files["file1"][0].OriginalName)
		}
	}))

	// Serve the HTTP request
	testHandler.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
