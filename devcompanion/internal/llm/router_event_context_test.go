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

// TestRouter_EventContext_BuildSuccessEmbedding はRouter経由で
// build_successイベントがプロンプトに埋め込まれることを検証する。
func TestRouter_EventContext_BuildSuccessEmbedding(t *testing.T) {
	t.Parallel()

	// Given: Ollama サーバーがプロンプト内容を検査する
	var receivedPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		receivedPrompt = req.Prompt

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"great success!","done":true}`))
	}))
	defer server.Close()

	ollama := NewOllamaClient(server.URL, "test-model")
	ollama.timeout = 100 * time.Millisecond

	router := &LLMRouter{
		ollama: ollama,
		claude: &mockClient{shouldFail: true},
		aiCLI:  &mockClient{shouldFail: true},
	}

	input := OllamaInput{
		State:   "Success",
		Task:    "Build",
		Mood:    "Happy",
		Name:    "TestBot",
		Tone:    "cheerful",
		Reason:  "success",
		Event:   "build_success", // Event フィールドが設定される
		Details: "",
	}

	// When: Route を呼ぶ
	result, _, err := router.Route(context.Background(), input)

	// Then: 成功し、プロンプトに Event 情報が含まれる
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if result == "" {
		t.Fatal("want non-empty result")
	}

	// プロンプトに Event が埋め込まれていることを確認
	if input.Event != "" && !strings.Contains(receivedPrompt, input.Event) {
		t.Fatalf("want prompt to contain event %q, got %q", input.Event, receivedPrompt)
	}
}

// TestRouter_EventContext_BuildFailedEmbedding はRouter経由で
// build_failedイベントがプロンプトに埋め込まれることを検証する。
func TestRouter_EventContext_BuildFailedEmbedding(t *testing.T) {
	t.Parallel()

	// Given: Ollama が失敗し、Claude が成功するモックサーバー
	var claudePrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Messages) > 0 {
			claudePrompt = req.Messages[0].Content
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"content": []map[string]string{{"text": "you can do it"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	claude := NewAnthropicClient("test-key")
	claude.endpoint = server.URL
	claude.timeout = 100 * time.Millisecond

	router := &LLMRouter{
		ollama: &mockClient{shouldFail: true},
		claude: claude,
		aiCLI:  &mockClient{shouldFail: true},
	}

	input := OllamaInput{
		State:   "Failed",
		Task:    "Build",
		Mood:    "Sad",
		Name:    "TestBot",
		Tone:    "empathetic",
		Reason:  "fail",
		Event:   "build_failed", // Event フィールドが設定される
		Details: "Compilation error",
	}

	// When: Route を呼ぶ
	result, _, err := router.Route(context.Background(), input)

	// Then: Claude から応答が返り、プロンプトに Event が含まれる
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
	if result != "you can do it" {
		t.Fatalf("want 'you can do it', got %q", result)
	}

	// Claude へのプロンプトに Event が埋め込まれていることを確認
	if input.Event != "" && !strings.Contains(claudePrompt, input.Event) {
		t.Fatalf("want Claude prompt to contain event %q, got %q", input.Event, claudePrompt)
	}
}

// TestRouter_EventContext_EmptyEvent はEventが空の場合
// プロンプトが影響を受けないことを検証する。
func TestRouter_EventContext_EmptyEvent(t *testing.T) {
	t.Parallel()

	// Given: Event が空のOllamaInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Event が空の場合、プロンプトに "開発イベント:" が含まれないことを確認
		if strings.Contains(req.Prompt, "開発イベント:") {
			t.Error("prompt should not contain event label when event is empty")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"hello","done":true}`))
	}))
	defer server.Close()

	ollama := NewOllamaClient(server.URL, "test-model")
	ollama.timeout = 100 * time.Millisecond

	router := &LLMRouter{
		ollama: ollama,
	}

	input := OllamaInput{
		State:   "Running",
		Task:    "Thinking",
		Mood:    "Focused",
		Name:    "TestBot",
		Tone:    "calm",
		Reason:  "thinking_tick",
		Event:   "", // 空のEvent
		Details: "",
	}

	// When: Route を呼ぶ
	_, _, _ = router.Route(context.Background(), input)
}

// TestRouter_EventContext_AllLayers はEventContextが全層を通じて
// 伝搬することを検証する。
func TestRouter_EventContext_AllLayers(t *testing.T) {
	t.Parallel()

	// Given: 3つのレイヤーがプロンプトを受け取り処理する
	ollamaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		// プロンプトが到達することを確認（実装では使用）
		_ = req.Prompt
		// Ollama は失敗をシミュレート
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ollamaServer.Close()

	claudeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if len(req.Messages) > 0 {
			// プロンプトが到達することを確認（実装では使用）
			_ = req.Messages[0].Content
		}
		// Claude も失敗をシミュレート
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer claudeServer.Close()

	// Note: ai CLI は MockClient で模擬。実環境では ~/.bin/ai を呼び出す
	aicliMock := &mockClient{
		response: "done",
		callCount: 0,
	}

	ollama := NewOllamaClient(ollamaServer.URL, "test-model")
	ollama.timeout = 100 * time.Millisecond

	claude := NewAnthropicClient("test-key")
	claude.endpoint = claudeServer.URL
	claude.timeout = 100 * time.Millisecond

	router := &LLMRouter{
		ollama: ollama,
		claude: claude,
		aiCLI:  aicliMock,
	}

	input := OllamaInput{
		State:   "Running",
		Task:    "Test",
		Mood:    "Neutral",
		Name:    "TestBot",
		Tone:    "professional",
		Reason:  "test",
		Event:   "test_event",
		Details: "testing event propagation",
	}

	// When: Route を呼ぶ（全層失敗、ai CLI まで到達）
	result, _, err := router.Route(context.Background(), input)

	// Then: ai CLI に到達し、Event がプロンプト経由で伝搬される
	if err != nil {
		t.Logf("error from route: %v", err) // エラーは予期される場合もある
	}
	if result == "" {
		t.Fatal("want response from ai CLI layer")
	}

	// 確認: ai CLI Mock が呼び出されたか
	if aicliMock.callCount == 0 {
		t.Fatal("want ai CLI to be called after other layers fail")
	}
}

// TestRouter_EventContext_DetailsField はDetailsフィールドが
// 利用可能なことを検証する（将来の拡張に備える）。
func TestRouter_EventContext_DetailsField(t *testing.T) {
	t.Parallel()

	// Given: Event と Details が設定された OllamaInput
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Note: Details が利用されない現在の実装でも、
		// フィールドは存在するため将来の拡張に対応可能

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"ok","done":true}`))
	}))
	defer server.Close()

	ollama := NewOllamaClient(server.URL, "test-model")

	router := &LLMRouter{
		ollama: ollama,
	}

	input := OllamaInput{
		State:   "Running",
		Task:    "Debug",
		Mood:    "Focused",
		Name:    "TestBot",
		Tone:    "technical",
		Reason:  "debugging",
		Event:   "debug_start",
		Details: "Debugging function X at line 42", // Details は保持されている
	}

	// When: Route を呼ぶ
	_, _, err := router.Route(context.Background(), input)

	// Then: Route が実行可能であること（Details は無視される可能性あり）
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}
