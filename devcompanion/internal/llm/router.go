package llm

import (
	"context"
	"log"
	"time"
)

const (
	ollamaRouterTimeout = 15 * time.Second
	claudeRouterTimeout = 10 * time.Second
	aicliRouterTimeout  = 10 * time.Second
)

// LLMRouter は複数のLLMバックエンドを優先度順にルーティングする。
type LLMRouter struct {
	ollama interface {
		Generate(ctx context.Context, in OllamaInput) (string, error)
	}
	claude interface {
		Generate(ctx context.Context, in OllamaInput) (string, error)
	}
	aiCLI interface {
		Generate(ctx context.Context, in OllamaInput) (string, error)
	}
}

// Route はプロンプトをLLMバックエンドにルーティングし、(応答テキスト, 使用したレイヤー名, エラー) を返す。
func (r *LLMRouter) Route(ctx context.Context, input OllamaInput) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}

	// Layer 1: Ollama
	result, ok := tryOllama(ctx, r.ollama, input)
	if ok {
		return result, "Ollama", nil
	}

	// Layer 2: Claude
	result, ok = tryClaude(ctx, r.claude, input)
	if ok {
		return result, "Claude", nil
	}

	// Layer 3: ai CLI (Gemini)
	result, ok = tryAICLI(ctx, r.aiCLI, input)
	if ok {
		return result, "Gemini", nil
	}

	// Layer 4: Fallback
	return FallbackSpeech(Reason(input.Reason)), "Fallback", nil
}

func tryOllama(ctx context.Context, client interface {
	Generate(ctx context.Context, in OllamaInput) (string, error)
}, input OllamaInput) (string, bool) {
	if client == nil {
		return "", false
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, ollamaRouterTimeout)
	defer cancel()
	result, err := client.Generate(timeoutCtx, input)
	if err != nil {
		log.Printf("[DEBUG] Ollama error: %v", err)
	}
	return result, err == nil && result != ""
}

func tryClaude(ctx context.Context, client interface {
	Generate(ctx context.Context, in OllamaInput) (string, error)
}, input OllamaInput) (string, bool) {
	if client == nil {
		return "", false
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, claudeRouterTimeout)
	defer cancel()
	result, err := client.Generate(timeoutCtx, input)
	if err != nil {
		log.Printf("[DEBUG] Claude error: %v", err)
	}
	return result, err == nil && result != ""
}

func tryAICLI(ctx context.Context, client interface {
	Generate(ctx context.Context, in OllamaInput) (string, error)
}, input OllamaInput) (string, bool) {
	if client == nil {
		return "", false
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, aicliRouterTimeout)
	defer cancel()
	result, err := client.Generate(timeoutCtx, input)
	if err != nil {
		log.Printf("[DEBUG] Gemini-CLI error: %v", err)
	}
	return result, err == nil && result != ""
}
