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
			ext := getFileExtension(originalFileName)
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

// Helper function to extract the file extension from a file name
func getFileExtension(fileName string) string {
	parts := strings.Split(fileName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}
