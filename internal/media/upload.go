package media

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"cloud.google.com/go/storage"
)

// Uploader handles gallery fallback uploads to Cloud Storage.
type Uploader struct {
	bucket string
	client *storage.Client
}

// NewUploader creates a new uploader with a Cloud Storage client.
// If client is nil, uploads will return an error.
func NewUploader(bucket string, client *storage.Client) *Uploader {
	return &Uploader{bucket: bucket, client: client}
}

// Upload stores a video/image in Cloud Storage and returns the gs:// URI for Gemini analysis.
func (u *Uploader) Upload(ctx context.Context, filename string, reader io.Reader, contentType string) (string, error) {
	if u.client == nil {
		return "", fmt.Errorf("storage client not initialized")
	}
	if u.bucket == "" {
		return "", fmt.Errorf("storage bucket not configured")
	}

	objectPath := "uploads/" + filename
	slog.Info("upload_start", "filename", filename, "bucket", u.bucket, "object", objectPath)

	obj := u.client.Bucket(u.bucket).Object(objectPath)
	w := obj.NewWriter(ctx)
	w.ContentType = contentType

	if _, err := io.Copy(w, reader); err != nil {
		w.Close()
		return "", fmt.Errorf("copy to storage: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("close storage writer: %w", err)
	}

	gsURI := fmt.Sprintf("gs://%s/%s", u.bucket, objectPath)
	slog.Info("upload_complete", "uri", gsURI, "content_type", contentType)
	return gsURI, nil
}
