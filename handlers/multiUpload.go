package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Xenos112/go-cdn/constants"
	"github.com/Xenos112/go-cdn/middleware"
	"github.com/Xenos112/go-cdn/processros"
)

func MultipleUploadHandler(w http.ResponseWriter, r *http.Request) {
	middleware.EnableCORS(w, r)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, constants.MaxUploadSize)
	if err := r.ParseMultipartForm(constants.MaxUploadSize); err != nil {
		http.Error(w, "Files too large or invalid form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	results := make([]constants.UploadResult, 0, len(files))

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			results = append(results, constants.UploadResult{
				FileName: header.Filename,
				Error:    "Failed to open file",
			})
			continue
		}

		result := processros.ProcessFileUpload(header, r.Host)
		results = append(results, result)
		file.Close()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
