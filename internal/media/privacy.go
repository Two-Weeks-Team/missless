package media

// PrivacyStatus classifies video accessibility for Gemini analysis.
type PrivacyStatus string

const (
	PrivacyPublic   PrivacyStatus = "public"   // Can be analyzed via YouTube URL
	PrivacyUnlisted PrivacyStatus = "unlisted" // Cannot be analyzed (needs gallery upload)
	PrivacyPrivate  PrivacyStatus = "private"  // Cannot be analyzed (needs gallery upload)
)

// CanAnalyzeViaURL returns true if the video can be analyzed directly.
// V7: Only public videos supported by Gemini FileData.
func (p PrivacyStatus) CanAnalyzeViaURL() bool {
	return p == PrivacyPublic
}

// ClassifyVideos separates videos by analyzability.
func ClassifyVideos(videos []YouTubeVideo) (analyzable, needsUpload []YouTubeVideo) {
	for _, v := range videos {
		if PrivacyStatus(v.Privacy).CanAnalyzeViaURL() {
			analyzable = append(analyzable, v)
		} else {
			needsUpload = append(needsUpload, v)
		}
	}
	return
}
