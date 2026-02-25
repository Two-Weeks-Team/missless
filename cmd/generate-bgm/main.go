// Command generate-bgm generates BGM preset audio files using Google's Lyria RealTime
// streaming API via direct WebSocket connection to the BidiGenerateMusic endpoint.
//
// Usage:
//
//	GEMINI_API_KEY=your_key go run cmd/generate-bgm/main.go [--output web/public/bgm] [--mood warm]
//
// By default, generates all 6 mood presets. Use --mood to generate a specific one.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

const (
	lyriaModel  = "models/lyria-realtime-exp"
	durationSec = 30
	sampleRate  = 48000 // Lyria outputs 48kHz
	channels    = 2     // Lyria outputs stereo
	wsEndpoint  = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1alpha.GenerativeService.BidiGenerateMusic"
)

// bgmPreset defines a mood preset for BGM generation.
type bgmPreset struct {
	Mood       string
	Prompts    []weightedPrompt
	BPM        int
	Brightness float64
	Density    float64
}

type weightedPrompt struct {
	Text   string  `json:"text"`
	Weight float64 `json:"weight"`
}

var presets = []bgmPreset{
	{
		Mood: "warm",
		Prompts: []weightedPrompt{
			{Text: "warm gentle ambient piano with soft pads", Weight: 1.0},
			{Text: "comforting peaceful acoustic", Weight: 0.7},
		},
		BPM: 70, Brightness: 0.4, Density: 0.3,
	},
	{
		Mood: "romantic",
		Prompts: []weightedPrompt{
			{Text: "romantic soft strings with gentle harp", Weight: 1.0},
			{Text: "dreamy love ballad instrumental", Weight: 0.8},
		},
		BPM: 65, Brightness: 0.3, Density: 0.3,
	},
	{
		Mood: "nostalgic",
		Prompts: []weightedPrompt{
			{Text: "nostalgic acoustic guitar with warm pads", Weight: 1.0},
			{Text: "bittersweet melody evoking old memories", Weight: 0.8},
		},
		BPM: 75, Brightness: 0.4, Density: 0.4,
	},
	{
		Mood: "playful",
		Prompts: []weightedPrompt{
			{Text: "playful light marimba with ukulele", Weight: 1.0},
			{Text: "cheerful bright happy instrumental", Weight: 0.7},
		},
		BPM: 100, Brightness: 0.7, Density: 0.6,
	},
	{
		Mood: "emotional",
		Prompts: []weightedPrompt{
			{Text: "emotional piano with deep cello", Weight: 1.0},
			{Text: "melancholic touching orchestral", Weight: 0.8},
		},
		BPM: 60, Brightness: 0.3, Density: 0.3,
	},
	{
		Mood: "farewell",
		Prompts: []weightedPrompt{
			{Text: "farewell gentle music box with soft ambient pads", Weight: 1.0},
			{Text: "peaceful goodbye ethereal ambience", Weight: 0.7},
		},
		BPM: 55, Brightness: 0.2, Density: 0.2,
	},
}

// Lyria WebSocket protocol message types (client → server).
type setupMsg struct {
	Setup struct {
		Model string `json:"model"`
	} `json:"setup"`
}

type clientContentMsg struct {
	ClientContent struct {
		WeightedPrompts []weightedPrompt `json:"weightedPrompts"`
	} `json:"client_content"`
}

type musicGenConfigMsg struct {
	MusicGenerationConfig struct {
		BPM         int     `json:"bpm,omitempty"`
		Brightness  float64 `json:"brightness,omitempty"`
		Density     float64 `json:"density,omitempty"`
		Temperature float64 `json:"temperature,omitempty"`
		Guidance    float64 `json:"guidance,omitempty"`
	} `json:"music_generation_config"`
}

type playbackControlMsg struct {
	PlaybackControl string `json:"playback_control"`
}

// Server response types.
type serverMsg struct {
	SetupComplete json.RawMessage `json:"setupComplete,omitempty"`
	ServerContent *serverContent  `json:"serverContent,omitempty"`
	Warning       json.RawMessage `json:"warning,omitempty"`
}

type serverContent struct {
	AudioChunks []audioChunk `json:"audioChunks,omitempty"`
}

type audioChunk struct {
	Data     string `json:"data"` // base64-encoded PCM
	MimeType string `json:"mimeType,omitempty"`
}

func main() {
	outputDir := flag.String("output", "web/public/bgm", "Output directory for BGM files")
	targetMood := flag.String("mood", "", "Generate specific mood only (empty = all)")
	flag.Parse()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: GEMINI_API_KEY environment variable required")
		os.Exit(1)
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: create output dir: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Filter presets if specific mood requested.
	targets := presets
	if *targetMood != "" {
		targets = nil
		for _, p := range presets {
			if p.Mood == *targetMood {
				targets = append(targets, p)
			}
		}
		if len(targets) == 0 {
			fmt.Fprintf(os.Stderr, "ERROR: unknown mood %q\n", *targetMood)
			os.Exit(1)
		}
	}

	fmt.Printf("Generating %d BGM presets to %s\n", len(targets), *outputDir)
	fmt.Printf("Using Lyria RealTime API (direct WebSocket)\n\n")

	for i, preset := range targets {
		fmt.Printf("[%d/%d] Generating: %s\n", i+1, len(targets), preset.Mood)

		pcmData, err := generateWithLyria(ctx, apiKey, preset)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: Lyria generation failed: %v\n", err)
			continue
		}

		mp3Path := filepath.Join(*outputDir, preset.Mood+".mp3")
		if err := pcmToMP3(pcmData, mp3Path); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR: convert to MP3: %v\n", err)
			continue
		}

		info, _ := os.Stat(mp3Path)
		fmt.Printf("  Saved: %s (%d KB)\n\n", mp3Path, info.Size()/1024)
	}

	fmt.Println("Done!")
}

// generateWithLyria connects directly to the Lyria RealTime WebSocket endpoint
// and generates music for the given preset.
func generateWithLyria(ctx context.Context, apiKey string, preset bgmPreset) ([]byte, error) {
	genCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// Build WebSocket URL with API key.
	wsURL, _ := url.Parse(wsEndpoint)
	q := wsURL.Query()
	q.Set("key", apiKey)
	wsURL.RawQuery = q.Encode()

	fmt.Printf("  Connecting to Lyria RealTime...\n")

	conn, _, err := websocket.DefaultDialer.DialContext(genCtx, wsURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial: %w", err)
	}
	defer conn.Close()

	// 1. Send setup message.
	setup := setupMsg{}
	setup.Setup.Model = lyriaModel
	if err := conn.WriteJSON(setup); err != nil {
		return nil, fmt.Errorf("send setup: %w", err)
	}

	// 2. Wait for setupComplete.
	if err := waitForSetupComplete(conn, genCtx); err != nil {
		return nil, fmt.Errorf("setup: %w", err)
	}
	fmt.Printf("  Session established.\n")

	// 3. Send weighted prompts.
	promptMsg := clientContentMsg{}
	promptMsg.ClientContent.WeightedPrompts = preset.Prompts
	if err := conn.WriteJSON(promptMsg); err != nil {
		return nil, fmt.Errorf("send prompts: %w", err)
	}

	// 4. Send music generation config.
	cfgMsg := musicGenConfigMsg{}
	cfgMsg.MusicGenerationConfig.BPM = preset.BPM
	cfgMsg.MusicGenerationConfig.Brightness = preset.Brightness
	cfgMsg.MusicGenerationConfig.Density = preset.Density
	cfgMsg.MusicGenerationConfig.Temperature = 1.1
	cfgMsg.MusicGenerationConfig.Guidance = 4.0
	if err := conn.WriteJSON(cfgMsg); err != nil {
		return nil, fmt.Errorf("send config: %w", err)
	}

	// 5. Send PLAY command.
	if err := conn.WriteJSON(playbackControlMsg{PlaybackControl: "PLAY"}); err != nil {
		return nil, fmt.Errorf("send play: %w", err)
	}

	// 6. Receive audio chunks until we have enough data.
	// 48kHz * 2ch * 2bytes * 30sec = 5,760,000 bytes
	targetBytes := durationSec * sampleRate * channels * 2
	var audioData []byte

	fmt.Printf("  Receiving audio...")
	for len(audioData) < targetBytes {
		select {
		case <-genCtx.Done():
			if len(audioData) > 0 {
				fmt.Println()
				fmt.Printf("  Timeout reached, using %d bytes collected\n", len(audioData))
				goto done
			}
			return nil, genCtx.Err()
		default:
		}

		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			if len(audioData) > 0 {
				break // Use whatever we collected.
			}
			return nil, fmt.Errorf("read message: %w", err)
		}

		var msg serverMsg
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			continue // Skip unparseable messages.
		}

		if msg.ServerContent != nil {
			for _, chunk := range msg.ServerContent.AudioChunks {
				if chunk.Data == "" {
					continue
				}
				decoded, err := base64.StdEncoding.DecodeString(chunk.Data)
				if err != nil {
					continue
				}
				audioData = append(audioData, decoded...)
				pct := min(100, len(audioData)*100/targetBytes)
				fmt.Printf("\r  Receiving audio... %d%%", pct)
			}
		}
	}

done:
	fmt.Println()

	// 7. Send STOP to end generation.
	_ = conn.WriteJSON(playbackControlMsg{PlaybackControl: "STOP"})

	if len(audioData) == 0 {
		return nil, fmt.Errorf("no audio data received")
	}

	// Trim to exact duration.
	if len(audioData) > targetBytes {
		audioData = audioData[:targetBytes]
	}

	fmt.Printf("  Generated %d bytes of audio (%.1fs at %dHz stereo)\n",
		len(audioData), float64(len(audioData))/float64(sampleRate*channels*2), sampleRate)

	return audioData, nil
}

// waitForSetupComplete reads messages until a setupComplete response is received.
func waitForSetupComplete(conn *websocket.Conn, ctx context.Context) error {
	deadline := time.Now().Add(15 * time.Second)
	conn.SetReadDeadline(deadline)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, rawMsg, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read setupComplete: %w", err)
		}

		var msg serverMsg
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			continue
		}

		if msg.SetupComplete != nil {
			return nil
		}
	}
}

// pcmToMP3 converts raw 48kHz stereo 16-bit PCM audio to MP3 using ffmpeg.
func pcmToMP3(pcmData []byte, outputPath string) error {
	tmpFile, err := os.CreateTemp("", "bgm-*.pcm")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(pcmData); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	cmd := exec.Command("ffmpeg", "-y",
		"-f", "s16le",
		"-ar", fmt.Sprintf("%d", sampleRate),
		"-ac", fmt.Sprintf("%d", channels),
		"-i", tmpFile.Name(),
		"-af", "afade=t=in:st=0:d=2,afade=t=out:st=28:d=2",
		"-codec:a", "libmp3lame",
		"-b:a", "192k",
		outputPath,
	)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
