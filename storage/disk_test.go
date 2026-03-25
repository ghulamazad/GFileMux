package storage

import (
	"bytes"
	"context"
	"os"
	"testing"

	GFileMux "github.com/ghulamazad/GFileMux"
)

func TestDiskStorage_AutoCreateDir(t *testing.T) {
	dir := t.TempDir() + "/new/nested/dir"
	_, err := NewDiskStorage(dir)
	if err != nil {
		t.Fatalf("NewDiskStorage should auto-create directories, got: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
}

func TestDiskStorage_Upload(t *testing.T) {
	dir := t.TempDir()
	ds, err := NewDiskStorage(dir)
	if err != nil {
		t.Fatalf("NewDiskStorage: %v", err)
	}

	content := []byte("hello, disk storage")
	meta, err := ds.Upload(context.Background(), bytes.NewReader(content), &GFileMux.UploadFileOptions{
		FileName: "test.txt",
		Bucket:   "",
	})
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if meta.Key != "test.txt" {
		t.Errorf("expected key 'test.txt', got %q", meta.Key)
	}
	if meta.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), meta.Size)
	}
}

func TestDiskStorage_Upload_WithBucket(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDiskStorage(dir)

	_, err := ds.Upload(context.Background(), bytes.NewReader([]byte("data")), &GFileMux.UploadFileOptions{
		FileName: "file.txt",
		Bucket:   "mybucket",
	})
	if err != nil {
		t.Fatalf("Upload with bucket: %v", err)
	}

	// Verify the file landed in the bucket subdirectory.
	bucketPath := dir + "/mybucket/file.txt"
	if _, err := os.Stat(bucketPath); err != nil {
		t.Fatalf("expected file at %s, got: %v", bucketPath, err)
	}
}

func TestDiskStorage_Path(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDiskStorage(dir)

	path, err := ds.Path(context.Background(), GFileMux.PathOptions{Key: "file.txt", Bucket: "bucket"})
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	expected := dir + "/bucket/file.txt"
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestDiskStorage_Delete(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDiskStorage(dir)

	ds.Upload(context.Background(), bytes.NewReader([]byte("data")), &GFileMux.UploadFileOptions{
		FileName: "todelete.txt",
	})

	if err := ds.Delete(context.Background(), "", "todelete.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := os.Stat(dir + "/todelete.txt"); !os.IsNotExist(err) {
		t.Fatal("file should have been deleted")
	}
}

func TestDiskStorage_Delete_NonExistent(t *testing.T) {
	dir := t.TempDir()
	ds, _ := NewDiskStorage(dir)
	err := ds.Delete(context.Background(), "", "ghost.txt")
	if err == nil {
		t.Fatal("expected error when deleting non-existent file")
	}
}
