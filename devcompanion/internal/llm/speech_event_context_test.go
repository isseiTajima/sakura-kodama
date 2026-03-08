package llm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devcompanion/internal/config"
	"devcompanion/internal/monitor"
	"devcompanion/internal/profile"
	"devcompanion/internal/types"
)

func TestSpeechGenerator_EventContext_BuildSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"test response","done":true}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Name:           "TestBot",
		Tone:           "calm",
		OllamaEndpoint: server.URL,
		Model:          "test-model",
	}

	sg := NewSpeechGenerator(cfg)

	event := monitor.MonitorEvent{
		State: types.StateCoding,
	}

	_ = sg.Generate(event, cfg, ReasonSuccess, profile.DevProfile{})
}

func TestSpeechGenerator_EventContext_BuildFailed(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"hang in there","done":true}`))
	}))
	defer server.Close()

	cfg := &config.Config{
		Name:           "TestBot",
		Tone:           "calm",
		OllamaEndpoint: server.URL,
		Model:          "test-model",
	}

	sg := NewSpeechGenerator(cfg)

	event := monitor.MonitorEvent{
		State: types.StateStuck,
	}

	_ = sg.Generate(event, cfg, ReasonFail, profile.DevProfile{})
}
