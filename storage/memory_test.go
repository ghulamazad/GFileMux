package storage

import (
	"bytes"
	"context"
	"sync"
	"testing"

	GFileMux "github.com/ghulamazad/GFileMux"
)

func TestMemoryStorage_Upload(t *testing.T) {
	ms := NewMemoryStorage()
	content := []byte("hello, memory")

	meta, err := ms.Upload(context.Background(), bytes.NewReader(content), &GFileMux.UploadFileOptions{
		FileName: "test.txt",
		Bucket:   "mybucket",
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

func TestMemoryStorage_Get(t *testing.T) {
	ms := NewMemoryStorage()
	content := []byte("stored content")

	ms.Upload(context.Background(), bytes.NewReader(content), &GFileMux.UploadFileOptions{
		FileName: "file.txt",
		Bucket:   "b",
	})

	data, err := ms.Get("b", "file.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Errorf("expected %q, got %q", content, data)
	}
}

func TestMemoryStorage_Get_NotFound(t *testing.T) {
	ms := NewMemoryStorage()
	_, err := ms.Get("b", "missing.txt")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	ms := NewMemoryStorage()
	ms.Upload(context.Background(), bytes.NewReader([]byte("data")), &GFileMux.UploadFileOptions{
		FileName: "del.txt",
		Bucket:   "b",
	})

	if err := ms.Delete(context.Background(), "b", "del.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := ms.Get("b", "del.txt"); err == nil {
		t.Fatal("expected file to be deleted")
	}
}

func TestMemoryStorage_Delete_NonExistent(t *testing.T) {
	ms := NewMemoryStorage()
	if err := ms.Delete(context.Background(), "b", "ghost.txt"); err == nil {
		t.Fatal("expected error when deleting non-existent file")
	}
}

func TestMemoryStorage_ConcurrentUploads(t *testing.T) {
	ms := NewMemoryStorage()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			name := "file" + string(rune('0'+i%10)) + ".txt"
			ms.Upload(context.Background(), bytes.NewReader([]byte("data")), &GFileMux.UploadFileOptions{
				FileName: name,
				Bucket:   "bucket",
			})
		}()
	}
	wg.Wait()
}
