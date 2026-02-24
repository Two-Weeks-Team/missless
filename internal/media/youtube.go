package media

import (
	"context"
	"fmt"
	"log/slog"
)

// YouTubeVideo represents a video from user's YouTube channel.
type YouTubeVideo struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
	PublishedAt  string `json:"publishedAt"`
	Privacy      string `json:"privacy"` // "public" | "unlisted" | "private"
	Duration     string `json:"duration"`
}

// YouTubeClient wraps YouTube Data API v3 operations.
type YouTubeClient struct {
	// TODO: T07 - Add google.golang.org/api/youtube/v3 client
}

// ListUserVideos fetches the user's YouTube videos using OAuth token.
// Only public videos can be analyzed via Gemini URL (V7 constraint).
func (yc *YouTubeClient) ListUserVideos(ctx context.Context, accessToken string) ([]YouTubeVideo, error) {
	slog.Info("youtube_list_start")

	// TODO: T07 - Implement YouTube Data API v3 calls
	// 1. channels.list (mine=true) → get uploads playlist ID
	// 2. playlistItems.list → get video IDs
	// 3. videos.list → get details + privacy status
	// 4. Classify: public, unlisted, private

	return nil, fmt.Errorf("not yet implemented")
}

// GetVideoURL returns the embeddable URL for Gemini analysis.
func GetVideoURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}
