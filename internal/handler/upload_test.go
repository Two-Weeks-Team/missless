package handler

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Two-Weeks-Team/missless/internal/media"
)

// newMultipartRequest creates a multipart/form-data POST request with a file
// field containing the given body bytes and Content-Type header on the part.
func newMultipartRequest(t *testing.T, fieldName, filename, contentType string, body []byte) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	partHeader := make(map[string][]string)
	partHeader["Content-Disposition"] = []string{
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename),
	}
	if contentType != "" {
		partHeader["Content-Type"] = []string{contentType}
	}

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed to create multipart part: %v", err)
	}
	if _, err := part.Write(body); err != nil {
		t.Fatalf("failed to write part body: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

// setupUploadMux creates a ServeMux with the upload handler registered using
// a nil-client uploader. The nil client causes Upload to return an error,
// which is useful for verifying that validation passes before the upload call.
func setupUploadMux() (*http.ServeMux, *media.Uploader) {
	mux := http.NewServeMux()
	uploader := media.NewUploader("test-bucket", nil)
	RegisterUpload(mux, uploader)
	return mux, uploader
}

func TestUpload_NoFileProvided(t *testing.T) {
	mux, _ := setupUploadMux()

	// Send a POST with an empty multipart form (no "file" field).
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing file, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "No file provided" {
		t.Fatalf("expected 'No file provided' message, got %q", body)
	}
}

func TestUpload_UnsupportedMIMEType(t *testing.T) {
	mux, _ := setupUploadMux()

	unsupportedTypes := []struct {
		mime     string
		filename string
	}{
		{"application/pdf", "doc.pdf"},
		{"text/plain", "readme.txt"},
		{"audio/mpeg", "song.mp3"},
		{"application/zip", "archive.zip"},
		{"video/x-msvideo", "clip.avi"},
	}

	for _, tc := range unsupportedTypes {
		t.Run(tc.mime, func(t *testing.T) {
			req := newMultipartRequest(t, "file", tc.filename, tc.mime, []byte("fake content"))
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			if w.Code != http.StatusUnsupportedMediaType {
				t.Fatalf("expected status 415 for MIME %s, got %d", tc.mime, w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != "Unsupported file type" {
				t.Fatalf("expected 'Unsupported file type' message, got %q", body)
			}
		})
	}
}

func TestUpload_AllowedMIMETypes_ReachUploader(t *testing.T) {
	// Verifies that each allowed MIME type passes validation and reaches the
	// uploader. Since the uploader has a nil storage client, it returns 500.
	// A 500 here proves the MIME check succeeded (not 415).
	mux, _ := setupUploadMux()

	allowedTypes := []struct {
		mime     string
		filename string
	}{
		{"video/mp4", "clip.mp4"},
		{"video/webm", "clip.webm"},
		{"video/quicktime", "clip.mov"},
		{"image/jpeg", "photo.jpg"},
		{"image/png", "photo.png"},
		{"image/webp", "photo.webp"},
	}

	for _, tc := range allowedTypes {
		t.Run(tc.mime, func(t *testing.T) {
			req := newMultipartRequest(t, "file", tc.filename, tc.mime, []byte("fake file data"))
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// The nil-client uploader returns an error, so the handler responds 500.
			// This proves the MIME type passed validation (otherwise we'd get 415).
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected status 500 (upload fail with nil client) for MIME %s, got %d", tc.mime, w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != "Upload failed" {
				t.Fatalf("expected 'Upload failed' message, got %q", body)
			}
		})
	}
}

func TestUpload_UploadFailure_NilClient(t *testing.T) {
	mux, _ := setupUploadMux()

	req := newMultipartRequest(t, "file", "video.mp4", "video/mp4", []byte("video bytes"))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 for upload failure, got %d", w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "Upload failed" {
		t.Fatalf("expected 'Upload failed' message, got %q", body)
	}
}

func TestUpload_FileTooLarge(t *testing.T) {
	mux, _ := setupUploadMux()

	// Create a payload that exceeds 100MB. We don't need to allocate 100MB+
	// of real data; we set Content-Length to trick MaxBytesReader into
	// rejecting the request during ParseMultipartForm.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	partHeader := make(map[string][]string)
	partHeader["Content-Disposition"] = []string{
		`form-data; name="file"; filename="huge.mp4"`,
	}
	partHeader["Content-Type"] = []string{"video/mp4"}

	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("failed to create multipart part: %v", err)
	}

	// Write slightly more than 100MB of zeros to exceed the limit.
	// To keep the test fast, we write in 1MB chunks.
	chunk := make([]byte, 1<<20) // 1MB
	for i := 0; i < 101; i++ {
		if _, err := part.Write(chunk); err != nil {
			t.Fatalf("failed to write chunk %d: %v", i, err)
		}
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413 for oversized file, got %d", w.Code)
	}
}

func TestUpload_MIMETypeWithParams(t *testing.T) {
	// Verify that MIME types with parameters (e.g., charset) are normalized
	// and still pass validation.
	mux, _ := setupUploadMux()

	req := newMultipartRequest(t, "file", "photo.jpg", "image/jpeg; charset=utf-8", []byte("jpeg data"))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should pass MIME validation (normalized to "image/jpeg") and reach the
	// uploader, which fails with nil client -> 500.
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 (MIME params normalized, upload fails), got %d", w.Code)
	}
}

func TestUpload_EmptyContentType_Rejected(t *testing.T) {
	// When the part has no Content-Type header, the handler defaults to
	// "application/octet-stream" which is not in the allowed list.
	mux, _ := setupUploadMux()

	req := newMultipartRequest(t, "file", "data.bin", "", []byte("binary data"))
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status 415 for empty content-type (defaults to application/octet-stream), got %d", w.Code)
	}
}

func TestUpload_MethodNotAllowed(t *testing.T) {
	mux, _ := setupUploadMux()

	req := httptest.NewRequest(http.MethodGet, "/api/upload", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Go 1.22+ ServeMux returns 405 for wrong method when pattern has method prefix.
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405 for GET request, got %d", w.Code)
	}
}
