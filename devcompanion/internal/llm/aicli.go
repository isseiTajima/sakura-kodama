package llm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AICLIClient は ~/.bin/ai CLI を呼び出してLLM応答を生成する。
type AICLIClient struct {
	timeout time.Duration
}

// NewAICLIClient は AICLIClient を作成する。
func NewAICLIClient() *AICLIClient {
	return &AICLIClient{
		timeout: aicliRouterTimeout, // router.go の定数を使用
	}
}

// Generate は ~/.bin/ai を実行してプロンプトを処理し、結果を返す。
// コマンド実行失敗やタイムアウト時はエラーを返す。
func (c *AICLIClient) Generate(ctx context.Context, in OllamaInput) (string, error) {
	// renderPrompt は ollama.go で定義されており、共通の promptTemplate を使用する
	prompt, err := renderPrompt(in)
	if err != nil {
		return "", fmt.Errorf("render prompt: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	aiPath := filepath.Join(homeDir, ".bin", "ai")

	// タイムアウト付きコンテキストを作成
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// コマンド実行
	cmd := exec.CommandContext(timeoutCtx, aiPath, prompt)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		// コンテキストキャンセル時はタイムアウトエラーとして扱う
		if timeoutCtx.Err() != nil {
			return "", fmt.Errorf("ai cli timeout: %w", timeoutCtx.Err())
		}
		// その他のエラーはコマンド実行エラーとして扱う
		return "", fmt.Errorf("ai cli error: %w", err)
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", fmt.Errorf("ai cli returned empty output")
	}

	return result, nil
}

func (c *AICLIClient) IsAvailable() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	aiPath := filepath.Join(homeDir, ".bin", "ai")
	_, err = os.Stat(aiPath)
	return err == nil
}
