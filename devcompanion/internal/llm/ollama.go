package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"
)

const (
	defaultOllamaEndpoint = "http://localhost:11434/api/generate"
	ollamaTimeout         = 15 * time.Second
	retryAttempts         = 3
)

var promptTemplate = template.Must(template.New("prompt").Parse(
	`あなたは{{.Name}}という名前のデスクトップキャラクター、話しかけている相手は「{{.UserName}}」です。
あなたの役割は、{{.UserName}}の作業を隣で見守り、応援することです。

[現在の{{.UserName}}の状況]
時間: {{.TimeOfDay}} (深夜)
作業モード: {{.SessionMode}} (集中度: {{.FocusLevel}})
最近の具体的な行動: {{.Behavior}}
進捗（信頼）: {{.Trust}}/100
あなたとの親密度: {{.RelationshipLvl}}/100

[指示]
- 40文字以内で、今の{{.UserName}}の心に寄り添う一言を。
- あなた自身の感想ではなく、あくまで「{{.UserName}}の状態」を観察して声をかけてください。
- {{if eq .SessionMode "deep_focus"}}{{.UserName}}は今、ゾーンに入っています。声をかけるのは控えめに、邪魔にならない短い言葉で。{{end}}
- {{if eq .SessionMode "struggling"}}{{.UserName}}は苦戦しているようです。そっと寄り添う言葉を。{{end}}
- 定型的な挨拶（「こんばんは」等）は禁止です。
- 親密度が高い場合は、幼馴染のような少し崩した口調で。
- 特殊記号やタグは出力せず、セリフのみ。`,
))


// OllamaInput はLLMへの入力パラメータ。
type OllamaInput struct {
	State           string
	Task            string
	Behavior        string // 行動 (coding, debugging, etc)
	SessionMode     string // モード (deep_focus, struggling, etc)
	FocusLevel      float64
	Mood            string
	Name            string
	UserName        string
	Tone            string
	Reason          string
	Event           string
	Details         string
	RelationshipLvl int
	Trust           int
	NightCoder      bool
	CommitFrequency string
	BuildFailRate   string
	TimeOfDay       string
}

// OllamaClient はOllama APIのクライアント。
type OllamaClient struct {
	endpoint string
	model    string
	timeout  time.Duration
}

// NewOllamaClient は OllamaClient を作成する。
func NewOllamaClient(endpoint, model string) *OllamaClient {
	if endpoint == "" {
		endpoint = defaultOllamaEndpoint
	}
	return &OllamaClient{
		endpoint: endpoint,
		model:    model,
		timeout:  ollamaTimeout,
	}
}

// Generate はOllama APIへリクエストし、生成されたテキストを返す。
func (c *OllamaClient) Generate(ctx context.Context, in OllamaInput) (string, error) {
	prompt, err := renderPrompt(in)
	if err != nil {
		return "", fmt.Errorf("prompt render: %w", err)
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"stream": false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < retryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(1 * time.Second)
		}
		timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
		req, reqErr := http.NewRequestWithContext(timeoutCtx, http.MethodPost, c.endpoint, bytes.NewReader(reqBody))
		if reqErr != nil {
			cancel()
			lastErr = fmt.Errorf("create request: %w", reqErr)
			break
		}
		req.Header.Set("Content-Type", "application/json")

		resp, httpErr := http.DefaultClient.Do(req)
		if httpErr != nil {
			cancel()
			lastErr = fmt.Errorf("Ollama connection failed (Is Ollama running?): %w", httpErr)
			continue
		}

		var result struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}
		decodeErr := json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		cancel()

		if decodeErr != nil {
			lastErr = fmt.Errorf("decode response: %w", decodeErr)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("ollama returned status %d", resp.StatusCode)
			continue
		}
		return strings.TrimSpace(result.Response), nil
	}

	return "", lastErr
}

func renderPrompt(in OllamaInput) (string, error) {
	var buf bytes.Buffer
	if err := promptTemplate.Execute(&buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
