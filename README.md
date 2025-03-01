# GFileMux

**GFileMux** is a lightweight, high-performance Golang library for handling multipart file uploads, inspired by Multer. It provides flexible storage options, middleware-style file handling, and efficient file processing.

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
```go
package main

import (
    "fmt"
    "net/http"
    "github.com/ghulamazad/GFileMux"
)

func main() {
    mux := http.NewServeMux()

    uploader := GFileMux.New(GFileMux.Config{
        Storage: GFileMux.NewDiskStorage("uploads/"),
        MaxSize: 10 * 1024 * 1024, // 10MB limit
    })

    mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
        file, err := uploader.Upload(r, "file")
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        fmt.Fprintf(w, "File uploaded: %s", file.Filename)
    })

    http.ListenAndServe(":8080", mux)
}
```

## Configuration
```go
config := GFileMux.Config{
    Storage: GFileMux.NewDiskStorage("uploads/"),
    MaxSize: 5 * 1024 * 1024, // 5MB file size limit
    FileFilter: func(file GFileMux.FileHeader) error {
        if file.ContentType != "image/png" && file.ContentType != "image/jpeg" {
            return errors.New("Only PNG and JPEG are allowed")
        }
        return nil
    },
}
```

## Contributing 🤝
Contributions are welcome! Feel free to open issues or submit pull requests to improve GFileMux.

## License
This project is licensed under the MIT License. 


