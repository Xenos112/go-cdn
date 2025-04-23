package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Xenos112/go-cdn/constants"
	"github.com/Xenos112/go-cdn/middleware"
	"github.com/Xenos112/go-cdn/processros"
)

func SingleUploadHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	result := processros.ProcessFileUpload(header, r.Host)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
