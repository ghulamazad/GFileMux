# GFileMux

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.23-blue)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/ghulamazad/GFileMux)](https://goreportcard.com/report/github.com/ghulamazad/GFileMux)

**GFileMux** is a fast, lightweight Go package for handling multipart file uploads. Inspired by Multer, it offers flexible storage options, middleware-style handling, and seamless processing with minimal overhead. Compatible with any Go HTTP framework, GFileMux simplifies file uploads for your web apps.

GFileMux is a personal passion project by me ([@ghulamazad](https://github.com/ghulamazad)). I built it to avoid rewriting multipart upload boilerplate and to create a clean, composable, framework-agnostic solution for Go—similar in spirit to Multer for Node.js. It’s developed in my free time.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Storage Backends](#storage-backends)
  - [Disk Storage](#disk-storage)
  - [Memory Storage](#memory-storage)
  - [S3 Storage](#s3-storage)
- [Validation](#validation)
  - [ValidateMimeType](#validatemimetype)
  - [ValidateFileExtension](#validatefileextension)
  - [ValidateMinFileSize](#validateminfilesize)
  - [ChainValidators](#chainvalidators)
- [Options](#options)
  - [WithStorage](#withstorage)
  - [WithMaxFileSize](#withmaxfilesize)
  - [WithMaxFiles](#withmaxfiles)
  - [WithFileValidatorFunc](#withfilevalidatorfunc)
  - [WithFileNameGeneratorFunc](#withfilenamegeneratorfunc)
  - [WithIgnoreNonExistentKey](#withignorenonexistentkey)
  - [WithUploadErrorHandlerFunc](#withuploaderrorhandlerfunc)
  - [WithAllowedBuckets](#withallowedbuckets)
  - [WithLogger](#withlogger)
  - [WithChecksumValidation](#withchecksumvalidation)
- [API Reference](#api-reference)
  - [Upload](#upload)
  - [UploadSingle](#uploadsingle)
  - [File](#file)
  - [Files helpers](#files-helpers)
  - [Storage Interface](#storage-interface)
  - [Error Types](#error-types)
- [Contributing](CONTRIBUTING.md)
- [Changelog](CHANGELOG.md)
- [License](#license)

## Features
✅ **Efficient File Parsing** – Handles `multipart/form-data` seamlessly.  
📂 **Flexible Storage** – Disk, in-memory, and Amazon S3 backends with a clean interface.  
🔍 **Rich Validation** – Filter by MIME type, file extension, and minimum/maximum size.  
🏷 **Custom Naming** – Define unique filename strategies via a pluggable function.  
⚡ **Concurrent Processing** – Processes multiple form fields in parallel using `errgroup` and `sync.Map`.  
🔒 **Bucket Allowlist** – Restrict which storage buckets may be used per handler.  
🔑 **SHA-256 Checksums** – Optionally compute and expose upload integrity hashes.  
📋 **Structured Errors** – Type-safe errors (`ValidationError`, `StorageError`, etc.) for precise error handling.  
📝 **Structured Logging** – Plug in a `log/slog.Logger` for lifecycle events.  
🛠 **Middleware Support** – Works with `net/http` and any compatible router/framework.  

## Installation

To install GFileMux in your Go project, use `go get` pointing to this repository.

To get the **latest release**:
```sh
go get github.com/ghulamazad/GFileMux@latest
```

To lock to a **specific version** (recommended for production):
```sh
go get github.com/ghulamazad/GFileMux@v0.2.0
```

## Quick Start
```go
package main

import (
    "context"
    "fmt"
    "log"
    "log/slog"
    "net/http"
    "strings"

    "github.com/ghulamazad/GFileMux"
    "github.com/ghulamazad/GFileMux/storage"
    "github.com/google/uuid"
)

func main() {
    // Disk storage — directory is auto-created if it does not exist.
    disk, err := storage.NewDiskStorage("./uploads")
    if err != nil {
        log.Fatalf("storage init: %v", err)
    }
    defer disk.Close()

    handler, err := GFileMux.New(
        GFileMux.WithStorage(disk),
        GFileMux.WithMaxFileSize(10<<20),            // 10 MB body limit
        GFileMux.WithMaxFiles(5),                    // max 5 files per field
        GFileMux.WithAllowedBuckets("images"),       // only "images" bucket allowed
        GFileMux.WithChecksumValidation(true),       // compute SHA-256 per file
        GFileMux.WithLogger(slog.Default()),         // structured logging
        GFileMux.WithFileValidatorFunc(
            GFileMux.ChainValidators(
                GFileMux.ValidateMimeType("image/jpeg", "image/png"),
                GFileMux.ValidateFileExtension(".jpg", ".jpeg", ".png"),
                GFileMux.ValidateMinFileSize(1024), // at least 1 KB
            ),
        ),
        GFileMux.WithFileNameGeneratorFunc(func(original string) string {
            ext := original[strings.LastIndex(original, "."):]
            return uuid.NewString() + ext
        }),
    )
    if err != nil {
        log.Fatalf("handler init: %v", err)
    }

    mux := http.NewServeMux()

    mux.Handle("/upload", handler.Upload("images", "photos")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        files, _ := GFileMux.GetUploadedFilesFromContext(r)
        fmt.Fprintf(w, "uploaded %d file(s)\n", files.Count())
        for _, f := range files.All() {
            path, _ := disk.Path(context.Background(), GFileMux.PathOptions{
                Key:    f.StorageKey,
                Bucket: f.FolderDestination,
            })
            fmt.Fprintf(w, "  %s → %s (sha256: %s)\n", f.OriginalName, path, f.ChecksumSHA256)
        }
    })))

    log.Println("listening on :3300")
    log.Fatal(http.ListenAndServe(":3300", mux))
}
```

## Configuration

```go
handler, err := GFileMux.New(
    GFileMux.WithStorage(disk),
    GFileMux.WithMaxFileSize(10<<20),
    GFileMux.WithMaxFiles(3),
    GFileMux.WithFileValidatorFunc(
        GFileMux.ChainValidators(
            GFileMux.ValidateMimeType("image/jpeg", "image/png"),
            GFileMux.ValidateMinFileSize(512),
        ),
    ),
    GFileMux.WithFileNameGeneratorFunc(func(orig string) string {
        return uuid.NewString() + filepath.Ext(orig)
    }),
    GFileMux.WithLogger(slog.Default()),
    GFileMux.WithChecksumValidation(true),
)
```

## Storage Backends

### Disk Storage
Files are stored on the local filesystem. The upload directory (and any bucket subdirectory) is **created automatically** if it does not exist.

```go
disk, err := storage.NewDiskStorage("./uploads")
if err != nil {
    log.Fatal(err)
}
```

Passing a `bucket` to `Upload()` stores files under `<directory>/<bucket>/`:
```go
handler.Upload("avatars", "photo") // → ./uploads/avatars/<filename>
```

Delete a stored file:
```go
err := disk.Delete(ctx, "avatars", "filename.jpg")
```

### Memory Storage
Keeps uploaded files in a thread-safe in-memory map. Primarily useful for testing.

```go
mem := storage.NewMemoryStorage()

// Retrieve raw bytes after upload:
data, err := mem.Get("bucket", "filename.jpg")

// Delete:
err = mem.Delete(ctx, "bucket", "filename.jpg")
```

### S3 Storage
```go
cfg, _ := config.LoadDefaultConfig(context.TODO())

s3Store, err := storage.NewS3FromConfig(cfg, storage.S3Options{
    UsePathStyle: true,
    ACL:          types.ObjectCannedACLPublicRead,
})

// Delete an S3 object:
err = s3Store.Delete(ctx, "my-bucket", "path/to/file.jpg")
```

## Validation

### ValidateMimeType
```go
GFileMux.ValidateMimeType("image/jpeg", "image/png", "application/pdf")
```

### ValidateFileExtension
```go
GFileMux.ValidateFileExtension(".jpg", ".jpeg", ".png")
```
Comparison is case-insensitive (`.JPG` matches `.jpg`).

### ValidateMinFileSize
```go
GFileMux.ValidateMinFileSize(1024) // reject files smaller than 1 KB
```

### ChainValidators
Combine multiple validators — the first failure short-circuits the chain:
```go
GFileMux.ChainValidators(
    GFileMux.ValidateMimeType("image/jpeg"),
    GFileMux.ValidateFileExtension(".jpg"),
    GFileMux.ValidateMinFileSize(512),
    func(f GFileMux.File) error {
        // custom validation logic
        return nil
    },
)
```

## Options

### WithStorage
```go
GFileMux.WithStorage(disk)
```

### WithMaxFileSize
```go
GFileMux.WithMaxFileSize(10 << 20) // 10 MB
```

### WithMaxFiles
Limit the number of files accepted per form field.
```go
GFileMux.WithMaxFiles(5)
```

### WithFileValidatorFunc
```go
GFileMux.WithFileValidatorFunc(GFileMux.ValidateMimeType("image/jpeg"))
```

### WithFileNameGeneratorFunc
```go
GFileMux.WithFileNameGeneratorFunc(func(original string) string {
    return uuid.NewString() + filepath.Ext(original)
})
```

### WithIgnoreNonExistentKey
```go
GFileMux.WithIgnoreNonExistentKey(true) // silently skip missing form fields
```

### WithUploadErrorHandlerFunc
```go
GFileMux.WithUploadErrorHandlerFunc(func(err error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(w, `{"error": %q}`, err.Error())
    }
})
```

### WithAllowedBuckets
Reject uploads to unlisted bucket names before any I/O occurs.
```go
GFileMux.WithAllowedBuckets("avatars", "documents")
```

### WithLogger
Attach a `log/slog` logger for structured lifecycle events.
```go
GFileMux.WithLogger(slog.Default())
// or a custom handler:
GFileMux.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
```

### WithChecksumValidation
When enabled, a SHA-256 hex digest is computed for each file and stored in `File.ChecksumSHA256`.
```go
GFileMux.WithChecksumValidation(true)
```

## API Reference

### Upload
Standard middleware for one or more form fields:
```go
handler.Upload("bucket", "field1", "field2")(nextHandler)
```

### UploadSingle
Convenience middleware that enforces exactly one file per field:
```go
handler.UploadSingle("avatars", "photo")(nextHandler)
```

### File
```go
type File struct {
    FieldName         string `json:"field_name,omitempty"`
    OriginalName      string `json:"original_name,omitempty"`
    UploadedFileName  string `json:"uploaded_file_name,omitempty"`
    FolderDestination string `json:"folder_destination,omitempty"`
    StorageKey        string `json:"storage_key,omitempty"`
    MimeType          string `json:"mime_type,omitempty"`
    Size              int64  `json:"size,omitempty"`
    ChecksumSHA256    string `json:"checksum_sha256,omitempty"`
}
```

### Files helpers
```go
files, _ := GFileMux.GetUploadedFilesFromContext(r)

files.All()     // []File — flat slice across all fields
files.Count()   // int   — total count across all fields

// By field:
byField, _ := GFileMux.GetFilesByFieldFromContext(r, "photos")
```

### Storage Interface
```go
type Storage interface {
    Upload(ctx context.Context, reader io.Reader, options *UploadFileOptions) (*UploadedFileMetadata, error)
    Path(ctx context.Context, options PathOptions) (string, error)
    Delete(ctx context.Context, bucket, key string) error
    io.Closer
}
```

### Error Types
Use `errors.As` to distinguish error categories:

```go
var ve *GFileMux.ValidationError
var se *GFileMux.StorageError
var mfe *GFileMux.MaxFilesError
var sizeErr *GFileMux.SizeError

switch {
case errors.As(err, &ve):
    // field validation failed
case errors.As(err, &mfe):
    // too many files
case errors.As(err, &sizeErr):
    // body too large
case errors.As(err, &se):
    // backend I/O error (se.Backend, se.Op, se.Unwrap())
}
```

## License
This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
