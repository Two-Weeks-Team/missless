package media

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const maxResults = 50

// YouTubeVideo represents a video from user's YouTube channel.
type YouTubeVideo struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl"`
	PublishedAt  string `json:"publishedAt"`
	Privacy      string `json:"privacy"` // "public" | "unlisted" | "private"
	Duration     string `json:"duration"`
}

// youtubeAPI abstracts YouTube Data API v3 for testability.
type youtubeAPI interface {
	GetUploadsPlaylistID(ctx context.Context) (string, error)
	ListPlaylistVideos(ctx context.Context, playlistID string) ([]string, error)
	GetVideoDetails(ctx context.Context, videoIDs []string) ([]YouTubeVideo, error)
}

// serviceFactory creates a youtubeAPI from an OAuth token.
type serviceFactory func(ctx context.Context, token *oauth2.Token) (youtubeAPI, error)

// YouTubeClient wraps YouTube Data API v3 operations.
type YouTubeClient struct {
	newService serviceFactory
}

// NewYouTubeClient creates a new YouTube client.
func NewYouTubeClient() *YouTubeClient {
	return &YouTubeClient{
		newService: func(ctx context.Context, token *oauth2.Token) (youtubeAPI, error) {
			svc, err := youtube.NewService(ctx, option.WithTokenSource(
				oauth2.StaticTokenSource(token),
			))
			if err != nil {
				return nil, fmt.Errorf("youtube service: %w", err)
			}
			return &realYouTubeService{svc: svc}, nil
		},
	}
}

// ListUserVideos fetches the user's YouTube videos using OAuth token.
// Only public videos can be analyzed via Gemini URL (V7 constraint).
func (yc *YouTubeClient) ListUserVideos(ctx context.Context, token *oauth2.Token) ([]YouTubeVideo, error) {
	slog.Info("youtube_list_start")

	svc, err := yc.newService(ctx, token)
	if err != nil {
		return nil, err
	}

	playlistID, err := svc.GetUploadsPlaylistID(ctx)
	if err != nil {
		return nil, fmt.Errorf("get uploads playlist: %w", err)
	}
	if playlistID == "" {
		return []YouTubeVideo{}, nil
	}

	videoIDs, err := svc.ListPlaylistVideos(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("list playlist videos: %w", err)
	}
	if len(videoIDs) == 0 {
		return []YouTubeVideo{}, nil
	}

	videos, err := svc.GetVideoDetails(ctx, videoIDs)
	if err != nil {
		return nil, fmt.Errorf("get video details: %w", err)
	}

	slog.Info("youtube_list_done", "count", len(videos))
	return videos, nil
}

// GetVideoURL returns the embeddable URL for Gemini analysis.
func GetVideoURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

// realYouTubeService wraps the actual YouTube Data API v3 client.
type realYouTubeService struct {
	svc *youtube.Service
}

func (r *realYouTubeService) GetUploadsPlaylistID(ctx context.Context) (string, error) {
	resp, err := r.svc.Channels.List([]string{"contentDetails"}).
		Mine(true).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("channels.list: %w", err)
	}
	if len(resp.Items) == 0 {
		return "", nil
	}
	return resp.Items[0].ContentDetails.RelatedPlaylists.Uploads, nil
}

func (r *realYouTubeService) ListPlaylistVideos(ctx context.Context, playlistID string) ([]string, error) {
	resp, err := r.svc.PlaylistItems.List([]string{"contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(maxResults).
		Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("playlistItems.list: %w", err)
	}

	var videoIDs []string
	for _, item := range resp.Items {
		videoIDs = append(videoIDs, item.ContentDetails.VideoId)
	}
	return videoIDs, nil
}

func (r *realYouTubeService) GetVideoDetails(ctx context.Context, videoIDs []string) ([]YouTubeVideo, error) {
	resp, err := r.svc.Videos.List([]string{"snippet", "contentDetails", "status"}).
		Id(videoIDs...).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("videos.list: %w", err)
	}

	var videos []YouTubeVideo
	for _, item := range resp.Items {
		thumb := ""
		if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
			thumb = item.Snippet.Thumbnails.Medium.Url
		}
		videos = append(videos, YouTubeVideo{
			ID:           item.Id,
			Title:        item.Snippet.Title,
			ThumbnailURL: thumb,
			PublishedAt:  item.Snippet.PublishedAt,
			Privacy:      item.Status.PrivacyStatus,
			Duration:     item.ContentDetails.Duration,
		})
	}
	return videos, nil
}
