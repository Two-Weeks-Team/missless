package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Two-Weeks-Team/missless/internal/media"
)

// maxUploadSize is 100MB — covers most short video clips.
const maxUploadSize = 100 << 20

// allowedMIMETypes are the MIME types accepted for upload.
var allowedMIMETypes = map[string]bool{
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
}

// RegisterUpload registers the media upload endpoint (gallery fallback).
func RegisterUpload(mux *http.ServeMux, uploader *media.Uploader) {
	mux.HandleFunc("POST /api/upload", func(w http.ResponseWriter, r *http.Request) {
		handleUpload(w, r, uploader)
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request, uploader *media.Uploader) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		slog.Warn("upload_parse_failed", "error", err)
		http.Error(w, "File too large (max 100MB)", http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Warn("upload_no_file", "error", err)
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	// Normalize MIME type (strip params like charset)
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	if !allowedMIMETypes[contentType] {
		http.Error(w, "Unsupported file type", http.StatusUnsupportedMediaType)
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".bin"
	}
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	slog.Info("upload_attempt", "filename", header.Filename, "size", header.Size, "type", contentType)

	gsURI, err := uploader.Upload(r.Context(), uniqueName, file, contentType)
	if err != nil {
		slog.Error("upload_failed", "error", err)
		http.Error(w, "Upload failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uri":          gsURI,
		"content_type": contentType,
		"filename":     header.Filename,
	})
}
