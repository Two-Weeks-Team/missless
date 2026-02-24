package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"
)

// Uploader handles gallery fallback uploads to Cloud Storage.
type Uploader struct {
	bucket string
	// TODO: T08 - Add storage.Client field
}

// NewUploader creates a new uploader.
func NewUploader(bucket string) *Uploader {
	return &Uploader{bucket: bucket}
}

// Upload stores a video/image in Cloud Storage and returns the URL.
func (u *Uploader) Upload(ctx context.Context, filename string, reader io.Reader, contentType string) (string, error) {
	slog.Info("upload_start", "filename", filename, "bucket", u.bucket)

	// TODO: T08 - Upload to Cloud Storage (uploads/ prefix)
	// 1. Create object writer
	// 2. Copy reader → writer
	// 3. Return gs:// URL for Gemini analysis

	return "", fmt.Errorf("not yet implemented")
}
