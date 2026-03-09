package llm

import (
	"context"
	"fmt"
	"log"
	"time"
)

const (
	ollamaRouterTimeout = 15 * time.Second
	claudeRouterTimeout = 10 * time.Second
	geminiRouterTimeout = 20 * time.Second
	aicliRouterTimeout  = 6 * time.Second
)

type LLMClient interface {
	Generate(ctx context.Context, in OllamaInput) (string, error)
	IsAvailable() bool
}

// LLMRouter は複数のLLMバックエンドを優先度順にルーティングする。
type LLMRouter struct {
	ollama LLMClient
	claude LLMClient
	gemini LLMClient
	aiCLI  LLMClient
}

// Route はプロンプトをLLMバックエンドにルーティングし、(応答テキスト, 使用したレイヤー名, エラー) を返す。
func (r *LLMRouter) Route(ctx context.Context, input OllamaInput) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}

	// Layer 1: Ollama
	if result, ok := r.try(ctx, r.ollama, ollamaRouterTimeout, input, "Ollama"); ok {
		return result, "Ollama", nil
	}

	// Layer 2: Claude
	if result, ok := r.try(ctx, r.claude, claudeRouterTimeout, input, "Claude"); ok {
		return result, "Claude", nil
	}

	// Layer 3: Gemini (API)
	if result, ok := r.try(ctx, r.gemini, geminiRouterTimeout, input, "Gemini"); ok {
		return result, "Gemini", nil
	}

	// Layer 4: ai CLI (Legacy)
	if result, ok := r.try(ctx, r.aiCLI, aicliRouterTimeout, input, "Gemini-CLI"); ok {
		return result, "Gemini-CLI", nil
	}

	return "", "", fmt.Errorf("all LLM backends failed")
}

func (r *LLMRouter) try(ctx context.Context, client LLMClient, timeout time.Duration, input OllamaInput, name string) (string, bool) {
	if client == nil || !client.IsAvailable() {
		return "", false
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	result, err := client.Generate(timeoutCtx, input)
	if err != nil {
		log.Printf("[DEBUG] %s error: %v", name, err)
		return "", false
	}
	return result, result != ""
}
