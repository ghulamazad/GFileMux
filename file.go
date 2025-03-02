package GFileMux

// File represents an uploaded file with relevant metadata.
type File struct {
	// FieldName indicates the name of the form field used for file upload in the multipart form.
	FieldName string `json:"field_name,omitempty"`

	// OriginalName is the name of the file as provided by the client.
	OriginalName string `json:"original_name,omitempty"`

	// UploadedFileName is the name of the file after it has been processed and stored.
	// This may differ from the original file name due to potential renaming during storage.
	UploadedFileName string `json:"uploaded_file_name,omitempty"`

	// FolderDestination is the path or directory where the file is stored within the storage system.
	FolderDestination string `json:"folder_destination,omitempty"`

	// StorageKey is the unique identifier used to retrieve the file from the storage backend.
	StorageKey string `json:"storage_key,omitempty"`

	// MimeType specifies the MIME type of the uploaded file (e.g., "image/jpeg", "application/pdf").
	MimeType string `json:"mime_type,omitempty"`

	// Size is the size of the uploaded file in bytes.
	Size int64 `json:"size,omitempty"`
}
