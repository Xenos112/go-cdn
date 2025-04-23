package main

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/Xenos112/go-cdn/constants"
	"github.com/Xenos112/go-cdn/handlers"
)

func main() {
	if err := mime.AddExtensionType(".webp", "image/webp"); err != nil {
		log.Fatal("Failed to register WebP MIME type:", err)
	}

	if err := os.MkdirAll(constants.UploadDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", handlers.SingleUploadHandler)
	http.HandleFunc("/uploads", handlers.MultipleUploadHandler)
	http.HandleFunc("/files/", handlers.DownloadHandler)

	fmt.Print("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
