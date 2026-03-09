package llm

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

const (
	defaultOllamaEndpoint = "http://localhost:11434/api/generate"
	ollamaTimeout         = 30 * time.Second
	retryAttempts         = 3
)

//go:embed prompts/*.tmpl
var promptFS embed.FS

var promptTemplates = make(map[string]*template.Template)

func init() {
	langs := []string{"ja", "en"}
	for _, lang := range langs {
		data, err := promptFS.ReadFile(fmt.Sprintf("prompts/%s.tmpl", lang))
		if err != nil {
			log.Printf("[WARN] Failed to load prompt template for %s: %v", lang, err)
			continue
		}
		tmpl, err := template.New(lang).Parse(string(data))
		if err != nil {
			log.Printf("[WARN] Failed to parse prompt template for %s: %v", lang, err)
			continue
		}
		promptTemplates[lang] = tmpl
	}
}

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
	Language        string // ja, en
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
	log.Printf("[DEBUG] Ollama requesting model: '%s' at %s", c.model, c.endpoint)
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

func (c *OllamaClient) IsAvailable() bool {
	return c.endpoint != ""
}

func renderPrompt(in OllamaInput) (string, error) {
	lang := in.Language
	if lang == "" {
		lang = "ja"
	}
	tmpl, ok := promptTemplates[lang]
	if !ok {
		// Fallback to ja
		tmpl = promptTemplates["ja"]
	}
	if tmpl == nil {
		return "", fmt.Errorf("no prompt template available for language: %s", lang)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}
