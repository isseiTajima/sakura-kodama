package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"devcompanion/internal/config"
	"devcompanion/internal/monitor"
	"devcompanion/internal/profile"
	"devcompanion/internal/types"
)

func TestSpeechGenerator_WithRouter_OllamaFail_ClaudeSuccess(t *testing.T) {
	t.Parallel()

	claudeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"content": []map[string]string{{"text": "claude says hello"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer claudeServer.Close()

	cfg := &config.Config{
		Name:           "TestBot",
		Tone:           "calm",
		EncourageFreq:  "mid",
		Monologue:      true,
		OllamaEndpoint: "http://localhost:9999",
		AnthropicAPIKey: "test-key",
	}

	sg := NewSpeechGenerator(cfg)
	sg.router.claude.(*AnthropicClient).endpoint = claudeServer.URL
	sg.router.claude.(*AnthropicClient).timeout = 100 * time.Millisecond

	event := monitor.MonitorEvent{
		State: types.StateDeepWork,
	}

	speech := sg.Generate(event, cfg, ReasonThinkingTick, profile.DevProfile{})
	if speech != "claude says hello" {
		t.Fatalf("want 'claude says hello' from Router fallback, got %q", speech)
	}
}

func TestSpeechGenerator_WithRouter_AllLayersFail_ReturnsFallback(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Name:            "TestBot",
		Tone:            "calm",
		EncourageFreq:   "mid",
		Monologue:       true,
		OllamaEndpoint:  "http://localhost:9999",
		AnthropicAPIKey: "",
	}

	sg := NewSpeechGenerator(cfg)
	sg.router.aiCLI = nil

	event := monitor.MonitorEvent{
		State: types.StateDeepWork,
	}

	speech := sg.Generate(event, cfg, ReasonThinkingTick, profile.DevProfile{})
	expectedFallback := FallbackSpeech(ReasonThinkingTick)
	if speech != expectedFallback {
		t.Fatalf("want fallback %q, got %q", expectedFallback, speech)
	}
}

func TestSpeechGenerator_NoNilPanicWithoutRouter(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Name:           "TestBot",
		Tone:           "calm",
		EncourageFreq:  "mid",
		Monologue:      true,
		OllamaEndpoint: "http://localhost:11434/api/generate",
	}

	sg := NewSpeechGenerator(cfg)

	event := monitor.MonitorEvent{
		State: types.StateIdle,
	}

	speech := sg.Generate(event, cfg, ReasonUserClick, profile.DevProfile{})
	if speech == "" {
		t.Fatal("want non-empty speech")
	}
}
