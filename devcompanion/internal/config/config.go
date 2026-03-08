package config

import (
	"devcompanion/internal/types"
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	Name                     string   `json:"name" yaml:"name"`
	UserName                 string   `json:"user_name" yaml:"user_name"`
	Tone                     string   `json:"tone" yaml:"tone"`
	EncourageFreq            string   `json:"encourage_freq" yaml:"encourage_freq"`
	Monologue                bool     `json:"monologue" yaml:"monologue"`
	AlwaysOnTop              bool     `json:"always_on_top" yaml:"always_on_top"`
	Mute                     bool     `json:"mute" yaml:"mute"`
	Model                    string   `json:"model" yaml:"model"`
	OllamaEndpoint           string   `json:"ollama_endpoint" yaml:"ollama_endpoint"`
	AnthropicAPIKey          string   `json:"anthropic_api_key" yaml:"anthropic_api_key"`
	LLMBackend               string   `json:"llm_backend" yaml:"llm_backend"`
	LogPaths                 []string `json:"log_paths" yaml:"log_paths"`
	AutoStart                bool     `json:"auto_start" yaml:"auto_start"`
	Scale                    float64  `json:"scale" yaml:"scale"`
	IndependentWindowOpacity float64  `json:"independent_window_opacity" yaml:"independent_window_opacity"`
	ClickThrough             bool     `json:"click_through" yaml:"click_through"`
	SetupCompleted           bool     `json:"setup_completed" yaml:"setup_completed"`
	SpeechFrequency          int      `json:"speech_frequency" yaml:"speech_frequency"`
}

type AppConfig struct {
	Config        `yaml:",inline"`
	IdleTimeout   int                        `yaml:"idle_timeout"`
	FocusWindow   int                        `yaml:"focus_window"`
	SignalWeights map[types.SignalType]float64 `yaml:"signal_weights"`
	PersonaStyle  types.PersonaStyle         `yaml:"persona_style"`
}

func DefaultConfig() *Config {
	return &Config{
		Name:                     "サクラ",
		UserName:                 "開発者",
		Tone:                     "フレンドリーな後輩",
		EncourageFreq:            "mid",
		Monologue:                true,
		AlwaysOnTop:              true,
		Mute:                     false,
		Model:                    "gemma3:4b",
		OllamaEndpoint:           "http://localhost:11434/api/generate",
		LogPaths:                 []string{""},
		AutoStart:                false,
		Scale:                    1.0,
		IndependentWindowOpacity: 1.0,
		ClickThrough:             true,
		SetupCompleted:           false,
		SpeechFrequency:          2,
	}
}

func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		Config:      *DefaultConfig(),
		IdleTimeout: 300,
		FocusWindow: 300,
		SignalWeights: map[types.SignalType]float64{
			types.SigProcessStarted: 0.5,
			types.SigFileModified:   0.1,
			types.SigGitCommit:      0.7,
			types.SigIdleStart:      0.8,
		},
		PersonaStyle: types.StyleSoft,
	}
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// Wails の設定やシグナルログと同じディレクトリに統一
	return filepath.Join(home, ".devcompanion", "config.yaml"), nil
}

func LoadConfig() (*AppConfig, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return DefaultAppConfig(), err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultAppConfig(), nil
	}

	cfg := DefaultAppConfig()
	// YAML優先
	if err := yaml.Unmarshal(data, cfg); err != nil {
		// 失敗した場合はJSONとして試行
		if err := json.Unmarshal(data, cfg); err != nil {
			return DefaultAppConfig(), err
		}
	}

	return cfg, nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
