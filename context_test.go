package GFileMux

import (
	"net/http/httptest"
	"testing"

	"net/http"
)

func TestGetUploadedFilesFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := GetUploadedFilesFromContext(req)
	if err == nil {
		t.Fatal("expected error when no files in context")
	}
}

func TestGetFilesByFieldFromContext_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	_, err := GetFilesByFieldFromContext(req, "field")
	if err == nil {
		t.Fatal("expected error when no files in context")
	}
}

func TestFiles_All(t *testing.T) {
	f := Files{
		"images": {
			{OriginalName: "a.jpg"},
			{OriginalName: "b.jpg"},
		},
		"docs": {
			{OriginalName: "c.pdf"},
		},
	}
	all := f.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 files from All(), got %d", len(all))
	}
}

func TestFiles_Count(t *testing.T) {
	f := Files{
		"images": {{OriginalName: "a.jpg"}, {OriginalName: "b.jpg"}},
		"docs":   {{OriginalName: "c.pdf"}},
	}
	if got := f.Count(); got != 3 {
		t.Fatalf("expected Count()=3, got %d", got)
	}
}

func TestFiles_CountEmpty(t *testing.T) {
	f := Files{}
	if got := f.Count(); got != 0 {
		t.Fatalf("expected Count()=0 for empty Files, got %d", got)
	}
}

func TestAddFilesToContext_Accumulates(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)

	first := Files{"a": {{OriginalName: "one.txt"}}}
	ctx := addFilesToContext(req.Context(), first)

	second := Files{"b": {{OriginalName: "two.txt"}}}
	ctx = addFilesToContext(ctx, second)

	all := getFilesFromContext(ctx)
	if all.Count() != 2 {
		t.Fatalf("expected 2 files after two addFilesToContext calls, got %d", all.Count())
	}
}
