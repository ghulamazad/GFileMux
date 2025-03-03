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
	// Initialize the memory storage for files
	memory := storage.NewMemoryStorage()

	// Set up the file handler with desired configurations
	handler, err := GFileMux.New(
		GFileMux.WithMaxFileSize(10<<20), // Limit file size to 10MB
		GFileMux.WithFileValidatorFunc(
			GFileMux.ChainValidators(GFileMux.ValidateMimeType("image/jpeg", "image/png")), // Validate file types
		),
		GFileMux.WithFileNameGeneratorFunc(func(originalFileName string) string {
			// Generate a new unique file name based on the UUID
			ext := getFileExtension(originalFileName)
			return fmt.Sprintf("%s.%s", uuid.NewString(), ext)
		}),
		GFileMux.WithStorage(memory), // Use in-memory storage
	)
	if err != nil {
		log.Fatalf("Error initializing file handler: %v", err)
	}

	// Create a new HTTP ServeMux
	mux := http.NewServeMux()

	// Handle file uploads on the root route
	mux.Handle("/", handler.Upload("bucket_name", "file1", "file2")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the uploaded files from the request context
		files, err := GFileMux.GetUploadedFilesFromContext(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get uploaded files: %v", err), http.StatusInternalServerError)
			return
		}

		// Retrieve the files by the field name "file1"
		file1, err := GFileMux.GetFilesByFieldFromContext(r, "file1")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get files by field 'file1': %v", err), http.StatusInternalServerError)
			return
		}

		// Log the details of the file uploaded in field "file1"
		fmt.Printf("Files in 'file1': %+v\n", file1)

		// Loop through all uploaded files and print their paths in memory
		for _, v := range files {
			fmt.Printf("Uploaded file: %+v\n", v)
			fmt.Println()

			// Print the path of the uploaded file in memory storage
			filePath, err := memory.Path(context.Background(), GFileMux.PathOptions{
				Key:    v[0].StorageKey,
				Bucket: v[0].FolderDestination,
			})

			if err != nil {
				// Handle the error properly and print it
				fmt.Printf("Error retrieving file path for storage key %s: %v\n", v[0].StorageKey, err)
			}

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
