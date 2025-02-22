package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	uploadDir     = "./uploads"
	maxUploadSize = 100 * 1024 * 1024 // 100MB
)

type UploadResult struct {
	FileName string `json:"fileName"`
	URL      string `json:"url"`
	Error    string `json:"error,omitempty"`
}

func main() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/upload", singleUploadHandler)
	http.HandleFunc("/uploads", multipleUploadHandler)
	http.HandleFunc("/files/", downloadHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Enable CORS and handle OPTIONS requests
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// Single file upload handler
func singleUploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result := processFileUpload(header, r.Host)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Multiple file upload handler
func multipleUploadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "Files too large or invalid form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	results := make([]UploadResult, 0, len(files))

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			results = append(results, UploadResult{
				FileName: header.Filename,
				Error:    "Failed to open file",
			})
			continue
		}

		result := processFileUpload(header, r.Host)
		results = append(results, result)
		file.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// Common file processing logic
func processFileUpload(header *multipart.FileHeader, host string) UploadResult {
	result := UploadResult{FileName: header.Filename}

	file, err := header.Open()
	if err != nil {
		result.Error = "Failed to open file"
		return result
	}
	defer file.Close()

	fileExt := filepath.Ext(header.Filename)
	newFileName := uuid.New().String() + fileExt
	filePath := filepath.Join(uploadDir, newFileName)

	dst, err := os.Create(filePath)
	if err != nil {
		result.Error = "Failed to create file"
		return result
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		result.Error = "Failed to save file"
		return result
	}

	result.URL = fmt.Sprintf("http://%s/files/%s", host, newFileName)
	return result
}

// File download handler
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/files/")
	if fileID == "" {
		http.Error(w, "Invalid file ID", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, fileID)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set content type based on file extension
	ext := filepath.Ext(filePath)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, r, filePath)
}

