package media

import (
	"context"
	"testing"

	"golang.org/x/oauth2"
)

type mockYouTubeAPI struct {
	uploadsPlaylistID string
	videoIDs          []string
	videos            []YouTubeVideo
	err               error
}

func (m *mockYouTubeAPI) GetUploadsPlaylistID(ctx context.Context) (string, error) {
	return m.uploadsPlaylistID, m.err
}

func (m *mockYouTubeAPI) ListPlaylistVideos(ctx context.Context, playlistID string) ([]string, error) {
	return m.videoIDs, m.err
}

func (m *mockYouTubeAPI) GetVideoDetails(ctx context.Context, videoIDs []string) ([]YouTubeVideo, error) {
	return m.videos, m.err
}

func TestYouTube_ListUserVideos_Empty(t *testing.T) {
	mock := &mockYouTubeAPI{
		uploadsPlaylistID: "", // Empty channel — no uploads playlist.
	}

	client := &YouTubeClient{
		newService: func(ctx context.Context, token *oauth2.Token) (youtubeAPI, error) {
			return mock, nil
		},
	}

	token := &oauth2.Token{AccessToken: "test-token"}
	videos, err := client.ListUserVideos(context.Background(), token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(videos) != 0 {
		t.Fatalf("expected 0 videos for empty channel, got %d", len(videos))
	}
}

func TestYouTube_PrivacyClassification(t *testing.T) {
	videos := []YouTubeVideo{
		{ID: "v1", Title: "Public Video", Privacy: "public"},
		{ID: "v2", Title: "Unlisted Video", Privacy: "unlisted"},
		{ID: "v3", Title: "Private Video", Privacy: "private"},
		{ID: "v4", Title: "Another Public", Privacy: "public"},
	}

	analyzable, needsUpload := ClassifyVideos(videos)

	if len(analyzable) != 2 {
		t.Fatalf("expected 2 analyzable videos, got %d", len(analyzable))
	}
	if len(needsUpload) != 2 {
		t.Fatalf("expected 2 needsUpload videos, got %d", len(needsUpload))
	}

	for _, v := range analyzable {
		if v.Privacy != "public" {
			t.Fatalf("expected public video in analyzable, got privacy=%s", v.Privacy)
		}
		if !PrivacyStatus(v.Privacy).CanAnalyzeViaURL() {
			t.Fatalf("expected CanAnalyzeViaURL() = true for public video")
		}
	}

	for _, v := range needsUpload {
		if v.Privacy == "public" {
			t.Fatalf("unexpected public video in needsUpload")
		}
		if PrivacyStatus(v.Privacy).CanAnalyzeViaURL() {
			t.Fatalf("expected CanAnalyzeViaURL() = false for %s video", v.Privacy)
		}
	}
}

func TestGetVideoURL(t *testing.T) {
	url := GetVideoURL("dQw4w9WgXcQ")
	expected := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	if url != expected {
		t.Fatalf("expected %q, got %q", expected, url)
	}
}
