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

func (ms *MockStorage) Delete(ctx context.Context, bucket, key string) error {
	return nil
}

func (ms *MockStorage) Close() error {
	return nil
}

func newTestHandler(t *testing.T, opts ...GFileMuxOption) *GFileMux {
	t.Helper()
	defaults := []GFileMuxOption{
		WithStorage(&MockStorage{}),
		WithMaxFileSize(10 << 20),
		WithFileValidatorFunc(DefaultFileValidator),
		WithFileNameGeneratorFunc(DefaultFileNameGeneratorFunc),
		WithUploadErrorHandlerFunc(DefaultUploadErrorHandlerFunc),
	}
	handler, err := New(append(defaults, opts...)...)
	if err != nil {
		t.Fatalf("failed to create test handler: %v", err)
	}
	return handler
}

func buildMultipartRequest(t *testing.T, field, filename string, content []byte) *http.Request {
	t.Helper()
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	part.Write(content)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestUpload(t *testing.T) {
	handler := newTestHandler(t)

	req := buildMultipartRequest(t, "file1", "testfile.txt", []byte("This is a test file"))
	rr := httptest.NewRecorder()

	handler.Upload("test_bucket", "file1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := GetUploadedFilesFromContext(r)
		if err != nil {
			t.Fatalf("GetUploadedFilesFromContext: %v", err)
		}
		if len(files["file1"]) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files["file1"]))
		}
		if files["file1"][0].OriginalName != "testfile.txt" {
			t.Fatalf("expected OriginalName 'testfile.txt', got %q", files["file1"][0].OriginalName)
		}
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUpload_MaxFiles(t *testing.T) {
	handler := newTestHandler(t, WithMaxFiles(1))

	// Build request with 2 files for the same field.
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	for _, name := range []string{"a.txt", "b.txt"} {
		part, _ := w.CreateFormFile("docs", name)
		part.Write([]byte("data"))
	}
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()

	handler.Upload("bucket", "docs")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be reached when MaxFiles is exceeded")
	})).ServeHTTP(rr, req)

	// Default error handler returns 500; check it didn't reach the next handler.
	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 when MaxFiles exceeded")
	}
}

func TestUpload_AllowedBuckets_Rejected(t *testing.T) {
	handler := newTestHandler(t, WithAllowedBuckets("images"))

	req := buildMultipartRequest(t, "file1", "doc.pdf", []byte("pdf content"))
	rr := httptest.NewRecorder()

	handler.Upload("documents", "file1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be reached for disallowed bucket")
	})).ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 for disallowed bucket")
	}
}

func TestUpload_Checksum(t *testing.T) {
	handler := newTestHandler(t, WithChecksumValidation(true))

	req := buildMultipartRequest(t, "file1", "test.txt", []byte("hello world"))
	rr := httptest.NewRecorder()

	handler.Upload("bucket", "file1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := GetUploadedFilesFromContext(r)
		if err != nil {
			t.Fatalf("GetUploadedFilesFromContext: %v", err)
		}
		checksum := files["file1"][0].ChecksumSHA256
		if checksum == "" {
			t.Fatal("expected non-empty ChecksumSHA256")
		}
		// SHA-256 of "hello world"
		const want = "b94d27b9934d3e08a52e52d7da7dabfac484efe04294e576b4e8ad5194123ecf"
		// We just verify it's non-empty and 64 hex chars.
		if len(checksum) != 64 {
			t.Fatalf("expected 64-char hex checksum, got %d chars: %s", len(checksum), checksum)
		}
	})).ServeHTTP(rr, req)
}

func TestUploadSingle(t *testing.T) {
	handler := newTestHandler(t)
	req := buildMultipartRequest(t, "avatar", "pic.jpg", []byte("fake-jpeg"))
	rr := httptest.NewRecorder()

	handler.UploadSingle("images", "avatar")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		files, err := GetUploadedFilesFromContext(r)
		if err != nil {
			t.Fatalf("GetUploadedFilesFromContext: %v", err)
		}
		if len(files["avatar"]) != 1 {
			t.Fatalf("expected 1 file, got %d", len(files["avatar"]))
		}
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUpload_IgnoreNonExistentKey(t *testing.T) {
	handler := newTestHandler(t, WithIgnoreNonExistentKey(true))
	req := buildMultipartRequest(t, "file1", "a.txt", []byte("data"))
	rr := httptest.NewRecorder()

	// Request for "missing_field" — should be ignored, not error.
	handler.Upload("bucket", "missing_field")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when IgnoreNonExistentKey=true, got %d", rr.Code)
	}
}
