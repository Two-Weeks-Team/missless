package handler

import (
	"log/slog"
	"net/http"

	"github.com/Two-Weeks-Team/missless/internal/config"
)

// RegisterUpload registers the media upload endpoint (gallery fallback).
func RegisterUpload(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("POST /api/upload", func(w http.ResponseWriter, r *http.Request) {
		handleUpload(w, r, cfg)
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// TODO: T08 - Implement gallery fallback upload
	// For unlisted/private YouTube videos, users can upload directly
	// 1. Accept multipart file upload
	// 2. Upload to Cloud Storage (uploads/ prefix)
	// 3. Return upload URL for Gemini analysis
	slog.Info("upload_attempt", "content_length", r.ContentLength)
	http.Error(w, "Upload not yet implemented", http.StatusNotImplemented)
}
