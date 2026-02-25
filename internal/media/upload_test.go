package media

import (
	"context"
	"strings"
	"testing"
)

func TestNewUploader(t *testing.T) {
	u := NewUploader("test-bucket", nil)
	if u == nil {
		t.Fatal("expected non-nil Uploader")
	}
	if u.bucket != "test-bucket" {
		t.Fatalf("expected bucket %q, got %q", "test-bucket", u.bucket)
	}
	if u.client != nil {
		t.Fatal("expected nil client")
	}
}

func TestUpload_NilClient(t *testing.T) {
	u := NewUploader("test-bucket", nil)

	uri, err := u.Upload(context.Background(), "photo.jpg", nil, "image/jpeg")
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
	if !strings.Contains(err.Error(), "storage client not initialized") {
		t.Fatalf("expected 'storage client not initialized' error, got: %v", err)
	}
	if uri != "" {
		t.Fatalf("expected empty URI on error, got %q", uri)
	}
}

func TestUpload_ContextCancellation(t *testing.T) {
	u := NewUploader("test-bucket", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	uri, err := u.Upload(ctx, "photo.jpg", nil, "image/jpeg")
	if err == nil {
		t.Fatal("expected error for nil client even with cancelled context, got nil")
	}
	// nil client check comes before any context usage, so the error should still
	// be about the uninitialized client.
	if !strings.Contains(err.Error(), "storage client not initialized") {
		t.Fatalf("expected 'storage client not initialized' error, got: %v", err)
	}
	if uri != "" {
		t.Fatalf("expected empty URI on error, got %q", uri)
	}
}
