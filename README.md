# GFileMux

**GFileMux** is a fast, lightweight Go package for handling multipart file uploads. Inspired by Multer, it offers flexible storage options, middleware-style handling, and seamless processing with minimal overhead. Compatible with any Go HTTP framework, GFileMux simplifies file uploads for your web apps.

## Table of Contents
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Storage Backends](#storage-backends)
  - [Disk Storage](#disk-storage)
  - [Memory Storage](#memory-storage)
  - [S3 Storage](#s3-storage)
- [API Reference](#api-reference)
  - [GFileMux](#gfilemux)
  - [File](#file)
  - [UploadFileOptions](#uploadfileoptions)
  - [UploadedFileMetadata](#uploadedfilemetadata)
  - [PathOptions](#pathoptions)
  - [Storage Interface](#storage-interface)
  - [FileValidatorFunc](#filevalidatorfunc)
  - [UploadErrorHandlerFunc](#uploaderrorhandlerfunc)
  - [FileNameGeneratorFunc](#filenamegeneratorfunc)
- [Options](#options)
  - [WithStorage](#withstorage)
  - [WithFileValidatorFunc](#withfilevalidatorfunc)
  - [WithFileValidatorFunc](#withfilevalidatorfunc)
  - [WithIgnoreNonExistentKey](#withignorenonexistentkey)
- [Contributing](/CONTRIBUTING.md)
- [License](#license)

## Features 
✅ **Efficient File Parsing** – Handles multipart/form-data seamlessly.  
📂 **Flexible Storage** – Supports disk and in-memory storage.  
🔍 **File Filtering** – Restrict uploads by type, size, and other conditions.  
🏷 **Custom Naming** – Define unique filename strategies.  
⚡ **Concurrent Processing** – Optimized for high-speed uploads.  
🛠 **Middleware Support** – Easily extend functionality.  



## Installation
```sh
go get github.com/ghulamazad/GFileMux
```

## Quick Start
Here is a quick example to get you started with GFileMux:
```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/ghulamazad/GFileMux"
	"github.com/ghulamazad/GFileMux/storage"
	"github.com/google/uuid"
)

func main() {
	// Initialize disk storage
	disk, err := storage.NewDiskStorage("./uploads")
	if err != nil {
		log.Fatalf("Error initializing disk storage: %v", err)
	}

	// Create a file handler with desired configurations
	handler, err := GFileMux.New(
		GFileMux.WithMaxFileSize(10<<20), // Limit file size to 10MB
		GFileMux.WithFileValidatorFunc(
			GFileMux.ChainValidators(GFileMux.ValidateMimeType("image/jpeg", "image/png")),
		),
		GFileMux.WithFileNameGeneratorFunc(func(originalFileName string) string {
			// Generate a new unique file name using UUID and original file extension
			parts := strings.Split(originalFileName, ".")
			ext := parts[len(parts)-1]
			return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
		}),
		GFileMux.WithStorage(disk), // Use disk storage
	)
	if err != nil {
		log.Fatalf("Error initializing file handler: %v", err)
	}

	// Create a new HTTP ServeMux
	mux := http.NewServeMux()

	// Handle file uploads on the root route
	mux.Handle("/", handler.Upload("bucket_name", "files")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the uploaded files from the request context
		files, err := GFileMux.GetUploadedFilesFromContext(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get uploaded files: %v", err), http.StatusInternalServerError)
			return
		}

		// Retrieve files by the field name "files"
		fileField, err := GFileMux.GetFilesByFieldFromContext(r, "files")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get files by field 'files': %v", err), http.StatusInternalServerError)
			return
		}

		// Log the details of files in the "files" field
		fmt.Printf("Files in 'files' field: %+v\n", fileField)

		// Process each uploaded file and print details
		for _, file := range files {
			// Log the details of each uploaded file
			fmt.Printf("Uploaded file details: %+v\n", file)

			// Print the file path in disk storage
			filePath, err := disk.Path(context.Background(), GFileMux.PathOptions{
				Key:    file[0].StorageKey,
				Bucket: file[0].FolderDestination,
			})
			if err != nil {
				log.Printf("Error retrieving file path for %s: %v", file[0].StorageKey, err)
				continue // Skip to the next file if there's an error
			}
			// Print the file path if no error
			fmt.Println("File path:", filePath)
		}
	})))

	// Start the HTTP server on port 3300
	log.Println("Starting server on :3300")
	if err := http.ListenAndServe(":3300", mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
```

## Configuration
You can configure GFileMux with various options. Here is an example configuration:
```go
config := GFileMux.New(
    GFileMux.WithMaxFileSize(10<<20), // Limit file size to 10MB
    GFileMux.WithFileValidatorFunc(
        GFileMux.ChainValidators(GFileMux.ValidateMimeType("image/jpeg", "image/png"),
            func(file GFileMux.File) error {
                // Add custom validation logic here if necessary
                // Alternatively, you can remove the ChainValidators and use just the MimeTypeValidator
                // or implement only your custom validation function if preferred
                return nil
            })),
    GFileMux.WithFileNameGeneratorFunc(func(originalFileName string) string {
        // Generate a new unique file name using UUID and original file extension
        parts := strings.Split(originalFileName, ".")
		ext := parts[len(parts)-1]
		return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
    }),
    GFileMux.WithStorage(storage.NewMemoryStorage()), // Use memory storage
)
```

## Storage Backends 
### Disk Storage
Disk storage saves uploaded files to a specified directory on the local filesystem.
```go
disk, err := storage.NewDiskStorage("./uploads")
if err != nil {
    log.Fatalf("Error initializing disk storage: %v", err)
}
```

### Memory Storage 
Memory storage keeps uploaded files in memory.
```go
memory := storage.NewMemoryStorage()
```

### S3 Storage
S3 storage uploads files to an Amazon S3 bucket.
```go
cfg, err := config.LoadDefaultConfig(context.TODO())
if err != nil {
    log.Fatalf("Error loading AWS config: %v", err)
}

s3Options := storage.S3Options{
    DebugMode:    true,
    UsePathStyle: true,
    ACL:          types.ObjectCannedACLPublicRead,
}

s3Store, err := storage.NewS3FromConfig(cfg, s3Options)
if err != nil {
    log.Fatalf("Error initializing S3 storage: %v", err)
}
```

## API Reference 
### GFileMux
The main struct for configuring and handling file uploads.

### File
Represents an uploaded file with relevant metadata.

### UploadFileOptions 
Holds the configuration for uploading a file.

### UploadedFileMetadata 
Contains metadata about a file after it has been uploaded.

### PathOptions
Holds options for generating the file's path.

### Storage Interface 
Defines the interface for interacting with file storage systems.

### FileValidatorFunc 
A function type used to validate a file during upload.

### UploadErrorHandlerFunc 
A custom function type used to handle errors when an upload fails.

### FileNameGeneratorFunc 
A function type that allows you to alter the name of the file before it is uploaded and stored.

## Options 
### WithStorage 
Sets the storage backend for the GFileMux instance.
```go
GFileMux.WithStorage(storage.NewDiskStorage("./uploads"))
```

### WithMaxFileSize
Sets the maximum file size for uploads.
```go
GFileMux.WithMaxFileSize(10<<20) // 10MB
```

### WithFileValidatorFunc 
Sets the file validation function.
```go
GFileMux.WithFileValidatorFunc(GFileMux.ValidateMimeType("image/jpeg", "image/png"))
```

### WithFileNameGeneratorFunc 
Sets the function to generate file names.
```go
GFileMux.WithFileNameGeneratorFunc(func(originalFileName string) string {
    parts := strings.Split(originalFileName, ".")
    ext := parts[len(parts)-1]
    return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
})
```

### WithIgnoreNonExistentKey 
Sets whether to ignore non-existent keys during file retrieval.
```go
GFileMux.WithIgnoreNonExistentKey(true)
```

### WithUploadErrorHandlerFunc 
A custom function type used to handle errors when an upload fails.
```go
GFileMux.WithUploadErrorHandlerFunc(func(err error) http.HandlerFunc {
    return func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprintf(w, `{"status": "error", "message": "GFileMux: File upload failed", "error": "%s"}`, err.Error())
    }
})
```

### FileNameGeneratorFunc 
A function type that allows you to alter the name of the file before it is uploaded and stored.
### Example
```go
GFileMux.WithFileNameGeneratorFunc(func(originalFileName string) string {
    parts := strings.Split(originalFileName, ".")
    ext := parts[len(parts)-1]
    return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
})
```

### FileValidatorFunc 
A function type used to validate a file during upload.
### Example
```go
GFileMux.ValidateMimeType("image/jpeg", "image/png")
```

## License
This project is licensed under the MIT License. 
