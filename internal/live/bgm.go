package live

// BGMPreset defines a mood-to-BGM URL mapping.
type BGMPreset struct {
	Mood string `json:"mood"`
	URL  string `json:"url"`
}

// presetBGMs maps mood keywords to BGM URLs.
// Local fallback paths serve from web/public/bgm/.
var presetBGMs = map[string]BGMPreset{
	"warm":      {Mood: "warm", URL: "/bgm/warm.mp3"},
	"romantic":  {Mood: "romantic", URL: "/bgm/romantic.mp3"},
	"nostalgic": {Mood: "nostalgic", URL: "/bgm/nostalgic.mp3"},
	"playful":   {Mood: "playful", URL: "/bgm/playful.mp3"},
	"emotional": {Mood: "emotional", URL: "/bgm/emotional.mp3"},
	"farewell":  {Mood: "farewell", URL: "/bgm/farewell.mp3"},
}

// defaultMood is used when the requested mood doesn't match any preset.
const defaultMood = "nostalgic"

// GetPresetBGMURL returns the BGM URL for a mood.
// Falls back to "nostalgic" for unknown moods.
func GetPresetBGMURL(mood string) BGMPreset {
	if p, ok := presetBGMs[mood]; ok {
		return p
	}
	return presetBGMs[defaultMood]
}

// AllPresetMoods returns all available BGM mood names.
func AllPresetMoods() []string {
	moods := make([]string, 0, len(presetBGMs))
	for k := range presetBGMs {
		moods = append(moods, k)
	}
	return moods
}
