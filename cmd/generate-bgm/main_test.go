package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// 1. Preset validation
// ---------------------------------------------------------------------------

func TestPresets_Valid(t *testing.T) {
	for i, p := range presets {
		t.Run(p.Mood, func(t *testing.T) {
			if p.Mood == "" {
				t.Errorf("preset %d has empty mood", i)
			}
			if len(p.Prompts) == 0 {
				t.Errorf("preset %d has no prompts", i)
			}
			if p.BPM <= 0 {
				t.Errorf("preset %d has invalid BPM: %d", i, p.BPM)
			}
			if p.Brightness < 0 || p.Brightness > 1 {
				t.Errorf("preset %d brightness out of range: %f", i, p.Brightness)
			}
			if p.Density < 0 || p.Density > 1 {
				t.Errorf("preset %d density out of range: %f", i, p.Density)
			}
			for j, prompt := range p.Prompts {
				if prompt.Text == "" {
					t.Errorf("preset %d prompt %d has empty text", i, j)
				}
				if prompt.Weight <= 0 {
					t.Errorf("preset %d prompt %d has non-positive weight: %f", i, j, prompt.Weight)
				}
			}
		})
	}
}

func TestPresets_AllMoods(t *testing.T) {
	expectedMoods := []string{"warm", "romantic", "nostalgic", "playful", "emotional", "farewell"}
	if len(presets) != len(expectedMoods) {
		t.Fatalf("expected %d presets, got %d", len(expectedMoods), len(presets))
	}
	for i, mood := range expectedMoods {
		if presets[i].Mood != mood {
			t.Errorf("preset %d: expected mood %q, got %q", i, mood, presets[i].Mood)
		}
	}
}

func TestPresets_UniqueMoods(t *testing.T) {
	seen := make(map[string]bool)
	for _, p := range presets {
		if seen[p.Mood] {
			t.Errorf("duplicate mood: %q", p.Mood)
		}
		seen[p.Mood] = true
	}
}

func TestPresets_BPMRanges(t *testing.T) {
	for _, p := range presets {
		t.Run(p.Mood, func(t *testing.T) {
			// BPM should be in a musically reasonable range (40-200).
			if p.BPM < 40 || p.BPM > 200 {
				t.Errorf("BPM %d outside reasonable range [40,200]", p.BPM)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. Protocol message JSON serialization (client -> server)
// ---------------------------------------------------------------------------

func TestSetupMsg_JSON(t *testing.T) {
	msg := setupMsg{}
	msg.Setup.Model = lyriaModel

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify the JSON contains expected fields.
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal round-trip failed: %v", err)
	}

	setup, ok := parsed["setup"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'setup' key in JSON, got: %s", data)
	}
	model, ok := setup["model"].(string)
	if !ok || model != lyriaModel {
		t.Errorf("expected model %q, got %q (JSON: %s)", lyriaModel, model, data)
	}
}

func TestClientContentMsg_JSON(t *testing.T) {
	msg := clientContentMsg{}
	msg.ClientContent.WeightedPrompts = []weightedPrompt{
		{Text: "gentle piano", Weight: 0.8},
		{Text: "ambient pads", Weight: 0.5},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Verify JSON structure using the correct key "client_content".
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal round-trip failed: %v", err)
	}

	cc, ok := parsed["client_content"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'client_content' key in JSON, got: %s", data)
	}

	prompts, ok := cc["weightedPrompts"].([]interface{})
	if !ok || len(prompts) != 2 {
		t.Errorf("expected 2 weightedPrompts, got: %s", data)
	}
}

func TestMusicGenConfigMsg_JSON(t *testing.T) {
	msg := musicGenConfigMsg{}
	msg.MusicGenerationConfig.BPM = 80
	msg.MusicGenerationConfig.Brightness = 0.5
	msg.MusicGenerationConfig.Density = 0.4
	msg.MusicGenerationConfig.Temperature = 1.1
	msg.MusicGenerationConfig.Guidance = 4.0

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal round-trip failed: %v", err)
	}

	cfg, ok := parsed["music_generation_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'music_generation_config' key in JSON, got: %s", data)
	}

	if bpm, ok := cfg["bpm"].(float64); !ok || int(bpm) != 80 {
		t.Errorf("expected bpm=80, got %v", cfg["bpm"])
	}
	if brightness, ok := cfg["brightness"].(float64); !ok || brightness != 0.5 {
		t.Errorf("expected brightness=0.5, got %v", cfg["brightness"])
	}
	if density, ok := cfg["density"].(float64); !ok || density != 0.4 {
		t.Errorf("expected density=0.4, got %v", cfg["density"])
	}
	if temp, ok := cfg["temperature"].(float64); !ok || temp != 1.1 {
		t.Errorf("expected temperature=1.1, got %v", cfg["temperature"])
	}
	if guidance, ok := cfg["guidance"].(float64); !ok || guidance != 4.0 {
		t.Errorf("expected guidance=4.0, got %v", cfg["guidance"])
	}
}

func TestMusicGenConfigMsg_OmitsZeroValues(t *testing.T) {
	msg := musicGenConfigMsg{}
	// All zero values -- omitempty should produce an empty config object.
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	cfg, ok := parsed["music_generation_config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'music_generation_config' key, got: %s", data)
	}

	// With omitempty, zero-valued fields should be absent.
	for _, key := range []string{"bpm", "brightness", "density", "temperature", "guidance"} {
		if _, exists := cfg[key]; exists {
			t.Errorf("expected %q to be omitted for zero value, got: %s", key, data)
		}
	}
}

func TestPlaybackControlMsg_JSON(t *testing.T) {
	tests := []struct {
		name   string
		action string
	}{
		{"play", "PLAY"},
		{"stop", "STOP"},
		{"pause", "PAUSE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := playbackControlMsg{PlaybackControl: tt.action}

			data, err := json.Marshal(msg)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatalf("unmarshal round-trip failed: %v", err)
			}

			action, ok := parsed["playback_control"].(string)
			if !ok || action != tt.action {
				t.Errorf("expected playback_control=%q, got %v (JSON: %s)", tt.action, parsed["playback_control"], data)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Server response deserialization
// ---------------------------------------------------------------------------

func TestServerMsg_SetupComplete(t *testing.T) {
	raw := `{"setupComplete": {}}`
	var msg serverMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if msg.SetupComplete == nil {
		t.Error("expected setupComplete to be non-nil")
	}
	if msg.ServerContent != nil {
		t.Error("expected serverContent to be nil")
	}
}

func TestServerMsg_AudioChunks(t *testing.T) {
	raw := `{
		"serverContent": {
			"audioChunks": [
				{"data": "AQID", "mimeType": "audio/pcm;rate=48000"},
				{"data": "BAUG", "mimeType": "audio/pcm;rate=48000"}
			]
		}
	}`
	var msg serverMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if msg.ServerContent == nil {
		t.Fatal("expected serverContent to be non-nil")
	}
	if len(msg.ServerContent.AudioChunks) != 2 {
		t.Fatalf("expected 2 audio chunks, got %d", len(msg.ServerContent.AudioChunks))
	}
	if msg.ServerContent.AudioChunks[0].MimeType != "audio/pcm;rate=48000" {
		t.Errorf("unexpected mimeType: %s", msg.ServerContent.AudioChunks[0].MimeType)
	}
}

func TestServerMsg_EmptyAudioChunks(t *testing.T) {
	raw := `{"serverContent": {"audioChunks": []}}`
	var msg serverMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if msg.ServerContent == nil {
		t.Fatal("expected serverContent to be non-nil")
	}
	if len(msg.ServerContent.AudioChunks) != 0 {
		t.Errorf("expected 0 chunks, got %d", len(msg.ServerContent.AudioChunks))
	}
}

func TestServerMsg_Warning(t *testing.T) {
	raw := `{"warning": {"message": "rate limited"}}`
	var msg serverMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if msg.Warning == nil {
		t.Error("expected warning to be non-nil")
	}
	if msg.SetupComplete != nil {
		t.Error("expected setupComplete to be nil")
	}
	if msg.ServerContent != nil {
		t.Error("expected serverContent to be nil")
	}
}

func TestServerMsg_UnknownFields(t *testing.T) {
	// Server may send fields we don't know about; unmarshal should not fail.
	raw := `{"unknownField": "something", "serverContent": {"audioChunks": [{"data": "AA=="}]}}`
	var msg serverMsg
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal should succeed with unknown fields: %v", err)
	}
	if msg.ServerContent == nil {
		t.Error("expected serverContent to be parsed despite unknown field")
	}
}

// ---------------------------------------------------------------------------
// 4. Weighted prompt roundtrip
// ---------------------------------------------------------------------------

func TestWeightedPrompt_Roundtrip(t *testing.T) {
	original := weightedPrompt{Text: "test prompt", Weight: 0.75}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded weightedPrompt
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Text != original.Text {
		t.Errorf("text mismatch: got %q, want %q", decoded.Text, original.Text)
	}
	if decoded.Weight != original.Weight {
		t.Errorf("weight mismatch: got %f, want %f", decoded.Weight, original.Weight)
	}
}

func TestWeightedPrompt_JSONTags(t *testing.T) {
	p := weightedPrompt{Text: "hello", Weight: 1.0}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify JSON keys use the tags ("text", "weight"), not Go field names.
	if _, ok := parsed["text"]; !ok {
		t.Errorf("expected 'text' key in JSON, got keys: %v", parsed)
	}
	if _, ok := parsed["weight"]; !ok {
		t.Errorf("expected 'weight' key in JSON, got keys: %v", parsed)
	}
}

// ---------------------------------------------------------------------------
// 5. Constants validation
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	if lyriaModel != "models/lyria-realtime-exp" {
		t.Errorf("unexpected lyriaModel: %s", lyriaModel)
	}
	if durationSec != 30 {
		t.Errorf("unexpected durationSec: %d", durationSec)
	}
	if sampleRate != 48000 {
		t.Errorf("unexpected sampleRate: %d", sampleRate)
	}
	if channels != 2 {
		t.Errorf("unexpected channels: %d", channels)
	}

	// Verify the target byte calculation matches expectation:
	// 30s * 48000Hz * 2ch * 2bytes = 5,760,000 bytes.
	expected := 5_760_000
	actual := durationSec * sampleRate * channels * 2
	if actual != expected {
		t.Errorf("target bytes: got %d, want %d", actual, expected)
	}
}

func TestWSEndpoint(t *testing.T) {
	if wsEndpoint == "" {
		t.Fatal("wsEndpoint should not be empty")
	}
	// Must use secure WebSocket.
	if !strings.HasPrefix(wsEndpoint, "wss:") {
		t.Errorf("wsEndpoint should use wss:// scheme, got: %s", wsEndpoint)
	}
}

// ---------------------------------------------------------------------------
// 6. Full protocol message sequence (simulates what generateWithLyria sends)
// ---------------------------------------------------------------------------

func TestProtocolSequence_Marshals(t *testing.T) {
	// Reproduce the exact sequence of messages the client sends, verifying
	// each one can be marshalled to valid JSON without error.

	preset := presets[0] // warm

	// Step 1: setup
	setup := setupMsg{}
	setup.Setup.Model = lyriaModel
	if _, err := json.Marshal(setup); err != nil {
		t.Fatalf("setup marshal: %v", err)
	}

	// Step 2: client content with prompts
	prompt := clientContentMsg{}
	prompt.ClientContent.WeightedPrompts = preset.Prompts
	if _, err := json.Marshal(prompt); err != nil {
		t.Fatalf("prompt marshal: %v", err)
	}

	// Step 3: music generation config
	cfg := musicGenConfigMsg{}
	cfg.MusicGenerationConfig.BPM = preset.BPM
	cfg.MusicGenerationConfig.Brightness = preset.Brightness
	cfg.MusicGenerationConfig.Density = preset.Density
	cfg.MusicGenerationConfig.Temperature = 1.1
	cfg.MusicGenerationConfig.Guidance = 4.0
	if _, err := json.Marshal(cfg); err != nil {
		t.Fatalf("config marshal: %v", err)
	}

	// Step 4: PLAY
	play := playbackControlMsg{PlaybackControl: "PLAY"}
	if _, err := json.Marshal(play); err != nil {
		t.Fatalf("play marshal: %v", err)
	}

	// Step 5: STOP (after receiving audio)
	stop := playbackControlMsg{PlaybackControl: "STOP"}
	if _, err := json.Marshal(stop); err != nil {
		t.Fatalf("stop marshal: %v", err)
	}
}

func TestProtocolSequence_AllPresets(t *testing.T) {
	// Every preset must produce valid JSON for the full protocol sequence.
	for _, preset := range presets {
		t.Run(preset.Mood, func(t *testing.T) {
			prompt := clientContentMsg{}
			prompt.ClientContent.WeightedPrompts = preset.Prompts
			if _, err := json.Marshal(prompt); err != nil {
				t.Errorf("prompt marshal failed for %s: %v", preset.Mood, err)
			}

			cfg := musicGenConfigMsg{}
			cfg.MusicGenerationConfig.BPM = preset.BPM
			cfg.MusicGenerationConfig.Brightness = preset.Brightness
			cfg.MusicGenerationConfig.Density = preset.Density
			cfg.MusicGenerationConfig.Temperature = 1.1
			cfg.MusicGenerationConfig.Guidance = 4.0
			if _, err := json.Marshal(cfg); err != nil {
				t.Errorf("config marshal failed for %s: %v", preset.Mood, err)
			}
		})
	}
}
