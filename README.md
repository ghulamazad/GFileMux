# GFileMux

**GFileMux** is a fast, lightweight Go package for handling multipart file uploads. Inspired by Multer, it offers flexible storage options, middleware-style handling, and seamless processing with minimal overhead. Compatible with any Go HTTP framework, GFileMux simplifies file uploads for your web apps.

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
    "fmt"
    "log"
    "net/http"
    "github.com/ghulamazad/GFileMux"
    "github.com/ghulamazad/GFileMux/storage"
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
		GFileMux.WithValidationFunc(
			GFileMux.ChainValidators(GFileMux.ValidateMimeType("image/jpeg", "image/png")),
		),
		GFileMux.WithNameFuncGenerator(func(originalFileName string) string {
			// Generate a new unique file name using UUID and original file extension
            parts := strings.Split(fileName, ".")
			ext := parts[len(parts)-1]
			return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
		}),
		GFileMux.WithStorage(storage.NewMemoryStorage()), // Use memory storage
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
config := handler, err := GFileMux.New(
		GFileMux.WithMaxFileSize(10<<20), // Limit file size to 10MB
		GFileMux.WithValidationFunc(
			GFileMux.ChainValidators(GFileMux.ValidateMimeType("image/jpeg", "image/png"),
				func(file GFileMux.File) error {
					// Add custom validation logic here if necessary
                    // Alternatively, you can remove the ChainValidators and use just the MimeTypeValidator
                    // or implement only your custom validation function if preferred
					return nil
				})),
		GFileMux.WithNameFuncGenerator(func(originalFileName string) string {
			// Generate a new unique file name using UUID and original file extension
			ext := getFileExtension(originalFileName)
			return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
		}),
		GFileMux.WithStorage(disk), // Use disk storage
	)
```

## Contributing 🤝
Contributions are welcome! Feel free to open issues or submit pull requests to improve GFileMux.

## License
This project is licensed under the MIT License. 


