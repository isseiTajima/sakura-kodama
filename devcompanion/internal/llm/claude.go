package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	anthropicEndpoint   = "https://api.anthropic.com/v1/messages"
	anthropicModel      = "claude-haiku-4-5-20251001"
	anthropicAPIVersion = "2023-06-01"
	anthropicTimeout    = 5 * time.Second
)

// AnthropicClient は Anthropic Messages API のクライアント。
type AnthropicClient struct {
	apiKey     string
	endpoint   string
	timeout    time.Duration
	httpClient *http.Client
}

// NewAnthropicClient は AnthropicClient を作成する。
// apiKey が空の場合は ANTHROPIC_API_KEY 環境変数を使用する。
func NewAnthropicClient(apiKey string) *AnthropicClient {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	// DisableKeepAlives: コンテキストキャンセル時に TCP 接続を即座に閉じる
	return &AnthropicClient{
		apiKey:   apiKey,
		endpoint: anthropicEndpoint,
		timeout:  anthropicTimeout,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
	}
}

// Generate は Anthropic Messages API へリクエストし、生成されたテキストを返す。
func (c *AnthropicClient) Generate(ctx context.Context, in OllamaInput) (string, error) {
	prompt, err := renderPrompt(in)
	if err != nil {
		return "", fmt.Errorf("prompt render: %w", err)
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":      anthropicModel,
		"max_tokens": 100,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", anthropicAPIVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic returned status %d", resp.StatusCode)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty content in response")
	}
	return strings.TrimSpace(result.Content[0].Text), nil
}
