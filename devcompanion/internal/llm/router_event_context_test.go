package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRouter_EventContext_BuildSuccessEmbedding(t *testing.T) {
	t.Parallel()
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"ok","done":true}`))
	}))
	defer server.Close()

	ollama := NewOllamaClient(server.URL, "test")
	ollama.timeout = 100 * time.Millisecond
	router := &LLMRouter{ollama: ollama}

	input := OllamaInput{
		Event: "build_success",
	}
	_, _, _ = router.Route(context.Background(), input)

	if !strings.Contains(receivedPrompt, "build_success") {
		t.Errorf("prompt missing event: %s", receivedPrompt)
	}
}

func TestRouter_EventContext_EmptyEvent(t *testing.T) {
	t.Parallel()
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"ok","done":true}`))
	}))
	defer server.Close()

	ollama := NewOllamaClient(server.URL, "test")
	router := &LLMRouter{ollama: ollama}

	input := OllamaInput{
		Event: "",
	}
	_, _, _ = router.Route(context.Background(), input)

	if strings.Contains(receivedPrompt, "直近のイベント:") {
		t.Errorf("prompt should not contain event label when empty: %s", receivedPrompt)
	}
}
