package llm

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	"devcompanion/internal/config"
	"devcompanion/internal/monitor"
	"devcompanion/internal/profile"
	"devcompanion/internal/types"
)

// SpeechGenerator はLLMを使用してセリフを生成する。
type SpeechGenerator struct {
	mu             sync.RWMutex
	router         *LLMRouter
	cache          *SpeechCache
	state          *SpeechState
	freq           *FrequencyController
	currentMood    MoodState
	successStreak  int
	usingFallback  bool
	lastSpeech     string
}

// NewSpeechGenerator は SpeechGenerator を作成する。
func NewSpeechGenerator(cfg *config.Config) *SpeechGenerator {
	return &SpeechGenerator{
		router: &LLMRouter{
			ollama: NewOllamaClient(cfg.OllamaEndpoint, cfg.Model),
			claude: NewAnthropicClient(cfg.AnthropicAPIKey),
			aiCLI:  NewAICLIClient(),
		},
		cache:       NewSpeechCache(),
		state:       NewSpeechState(),
		freq:        NewFrequencyController(),
		currentMood: MoodStateHappy,
	}
}

func (sg *SpeechGenerator) Generate(e monitor.MonitorEvent, cfg *config.Config, reason Reason, prof profile.DevProfile) string {
	if cfg.Mute {
		return ""
	}

	now := time.Now()
	if !sg.freq.ShouldSpeak(reason, e.State, cfg, now) {
		return ""
	}

	speech := sg.generateText(e, cfg, reason, prof)
	sg.freq.RecordSpeak(reason, e.State, cfg, now)

	// Update successStreak
	sg.mu.Lock()
	if reason == ReasonSuccess {
		sg.successStreak++
	} else if reason == ReasonFail {
		sg.successStreak = 0
	}
	sg.mu.Unlock()

	return postProcess(speech)
}

func (sg *SpeechGenerator) IsUsingFallback() bool {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.usingFallback
}

func (sg *SpeechGenerator) setUsingFallback(v bool) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.usingFallback = v
}

// UpdateLLMConfig は設定変更時にLLMクライアントを更新する。
func (sg *SpeechGenerator) UpdateLLMConfig(cfg *config.Config) {
	sg.mu.Lock()
	defer sg.mu.Unlock()
	sg.router.ollama = NewOllamaClient(cfg.OllamaEndpoint, cfg.Model)
	sg.router.claude = NewAnthropicClient(cfg.AnthropicAPIKey)
}

// OnUserClick はユーザークリック時のセリフを生成する。
func (sg *SpeechGenerator) OnUserClick(e monitor.MonitorEvent, cfg *config.Config, prof profile.DevProfile) string {
	return sg.Generate(e, cfg, ReasonUserClick, prof)
}

func (sg *SpeechGenerator) generateText(e monitor.MonitorEvent, cfg *config.Config, reason Reason, prof profile.DevProfile) string {
	sg.mu.RLock()
	router := sg.router
	cache := sg.cache
	state := sg.state
	sg.mu.RUnlock()

	// Update current mood
	now := time.Now().UTC()
	newMood := InferMoodState(now, sg.successStreak, reason)

	sg.mu.Lock()
	sg.currentMood = newMood
	sg.mu.Unlock()

	moodStr := string(newMood)
	eventStr := reasonToEventContext(reason)
	cacheKey := eventStr + ":" + moodStr

	// 1. キャッシュのチェック
	if eventStr != "" && cache != nil {
		if cached, hit := cache.Get(cacheKey); hit {
			log.Printf("[DEBUG] Cache hit: key=%s, text='%s'", cacheKey, cached)
			return cached
		}
	}

	input := OllamaInput{
		State:           string(e.State),
		Task:            string(e.Task),
		Behavior:        string(e.Behavior.Type),
		SessionMode:     string(e.Session.Mode),
		FocusLevel:      e.Session.FocusLevel,
		Mood:            moodStr,
		Name:            cfg.Name,
		UserName:        cfg.UserName,
		Tone:            cfg.Tone,
		Reason:          string(reason),
		Event:           eventStr,
		Details:         e.Details,
		RelationshipLvl: prof.Relationship.Level,
		Trust:           prof.Relationship.Trust,
		NightCoder:      prof.NightCoder,
		CommitFrequency: prof.CommitFrequency,
		BuildFailRate:   prof.BuildFailRate,
		TimeOfDay:       getTimeOfDay(time.Now().Hour()),
	}

	var text string
	var backend string
	maxRetries := 2

	// 2. LLM 呼び出しと重複チェック
	for retry := 0; retry < maxRetries; retry++ {
		var err error
		text, backend, err = router.Route(context.Background(), input)
		if err != nil {
			log.Printf("[DEBUG] Router error: %v", err)
			return FallbackSpeech(reason)
		}

		if state != nil && state.IsDuplicate(text) {
			log.Printf("[INFO] Speech rejected due to duplicate: line='%s' (retry %d/%d)", text, retry+1, maxRetries)
			if retry < maxRetries-1 {
				continue 
			}
		}
		break
	}

	// キャッシュに保存
	if eventStr != "" && cache != nil && text != "" {
		cache.Put(cacheKey, text)
	}
	if state != nil && text != "" {
		state.AddLine(text)
	}

	if text == "" {
		log.Printf("[DEBUG] LLM returned empty/invalid text, using fallback")
		return FallbackSpeech(reason)
	}

	log.Printf("[DEBUG] Final generated speech [%s]: '%s'", backend, text)
	return text
}

func reasonToEventContext(reason Reason) string {
	switch reason {
	case ReasonSuccess:
		return "build_success"
	case ReasonFail:
		return "build_failed"
	case ReasonGreeting:
		return "greeting"
	default:
		return ""
	}
}

func postProcess(s string) string {
	s = strings.TrimSpace(s)
	// 無効なブラケットのみの回答を排除
	if s == "[]" || s == "{}" || s == "「」" || s == "()" || s == "" {
		return ""
	}

	// セリフ全体を囲む引用符やカッコを削除
	// 左右がペアになっている場合のみ削除することで、文中のカッコを誤削除しないようにする
	prefixes := []string{"「", "『", "(", "（", "\"", "'"}
	suffixes := []string{"」", "』", ")", "）", "\"", "'"}

	for i := range prefixes {
		if strings.HasPrefix(s, prefixes[i]) && strings.HasSuffix(s, suffixes[i]) {
			s = strings.TrimPrefix(s, prefixes[i])
			s = strings.TrimSuffix(s, suffixes[i])
		}
	}

	s = strings.TrimSpace(s)
	
	runes := []rune(s)
	if len(runes) > 80 {
		runes = runes[:80]
	}
	return string(runes)
}

// FrequencyController は発話頻度を制御する。
type FrequencyController struct {
	mu             sync.RWMutex
	lastState      types.ContextState
	lastSpeakTime  time.Time
	cooldownUntil  time.Time
	consecutive    int
}

// NewFrequencyController は FrequencyController を作成する。
func NewFrequencyController() *FrequencyController {
	return &FrequencyController{}
}

// ShouldSpeak は今喋るべきかを判定する。
func (fc *FrequencyController) ShouldSpeak(reason Reason, state types.ContextState, cfg *config.Config, now time.Time) bool {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// ユーザーがクリックした場合は必ず喋る
	if reason == ReasonUserClick {
		return true
	}

	// 前回の発話から短時間（例: 30秒）は、どんな理由でも沈黙を守る（連続発話防止）
	if !fc.lastSpeakTime.IsZero() && now.Sub(fc.lastSpeakTime) < 30*time.Second {
		return false
	}

	// 理由ごとの詳細制御
	switch reason {
	case ReasonSuccess, ReasonFail:
		// 状態が変化した瞬間だけ喋る
		if state != fc.lastState {
			return true
		}
		return false

	case ReasonThinkingTick:
		if !cfg.Monologue || now.Before(fc.cooldownUntil) {
			return false
		}
		
		// 頻度設定に応じた間隔（ThinkingTick用）
		interval := 10 * time.Minute
		switch cfg.SpeechFrequency {
		case 1: // 控えめ
			interval = 20 * time.Minute
		case 3: // お喋り
			interval = 1 * time.Minute 
		}

		if now.Sub(fc.lastSpeakTime) < interval {
			return false
		}
		return true
	}

	// その他の理由（Gitコミット、セッション開始など）は、
	// 最初の30秒間隔チェックをパスしていれば許可
	return true
}

// RecordSpeak は発話したことを記録する。
func (fc *FrequencyController) RecordSpeak(reason Reason, state types.ContextState, cfg *config.Config, now time.Time) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.lastSpeakTime = now
	fc.lastState = state
	
	if reason == ReasonThinkingTick {
		fc.consecutive++
		if fc.consecutive >= 3 {
			fc.cooldownUntil = now.Add(15 * time.Minute)
			fc.consecutive = 0
		}
	} else {
		fc.consecutive = 0
	}
}

func getTimeOfDay(h int) string {
	switch {
	case h >= 5 && h < 10:
		return "朝"
	case h >= 10 && h < 17:
		return "昼"
	case h >= 17 && h < 20:
		return "夕方"
	case h >= 20 && h < 23:
		return "夜"
	default:
		return "深夜"
	}
}

func (sg *SpeechGenerator) fallbackSpeech(reason Reason) string {
	return FallbackSpeech(reason)
}
