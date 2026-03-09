package engine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"devcompanion/internal/config"
	contextengine "devcompanion/internal/context"
	"devcompanion/internal/llm"
	"devcompanion/internal/monitor"
	"devcompanion/internal/observer"
	"devcompanion/internal/persona"
	"devcompanion/internal/profile"
	"devcompanion/internal/types"
)

// Notifier はエンジンからのイベントを外部（Wails/WebSocket）に通知するインターフェース。
type Notifier interface {
	Notify(event types.Event)
}

// Engine はモニタリング、状況推定、人格生成、通知を統合するコアエンジン。
type Engine struct {
	monitor  *monitor.Monitor
	context  *contextengine.Estimator
	persona  *persona.PersonaEngine
	speech   *llm.SpeechGenerator
	profile  *profile.ProfileStore
	observer *observer.DevObserver
	notifier Notifier
	cfg      *config.Config

	lastEvent monitor.MonitorEvent
}

// New は新しい Engine を作成する。
func New(m *monitor.Monitor, ctxEng *contextengine.Estimator, p *persona.PersonaEngine, s *llm.SpeechGenerator, prof *profile.ProfileStore, obs *observer.DevObserver, c *config.Config, n Notifier) *Engine {
	return &Engine{
		monitor:  m,
		context:  ctxEng,
		persona:  p,
		speech:   s,
		profile:  prof,
		observer: obs,
		notifier: n,
		cfg:      c,
		lastEvent: monitor.MonitorEvent{
			State: types.StateIdle,
		},
	}
}

// Run はメインのパイプライン処理を開始する。
func (e *Engine) Run(ctx context.Context) {
	events := e.monitor.Events()
	var observations <-chan observer.DevObservation
	if e.observer != nil {
		observations = e.observer.Observations()
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("[ENGINE] Pipeline stopped")
			return

		case ev := <-events:
			log.Printf("[ENGINE] Received monitor event: state=%s, event=%s", ev.State, ev.Event)
			reason := e.reasonFromEvent(ev)
			
			e.profile.RecordActivity(time.Now())
			if ev.State == types.StateSuccess {
				e.profile.RecordBuildSuccess()
			} else if ev.State == types.StateFail {
				e.profile.RecordBuildFail()
			}
			
			if e.observer != nil {
				e.observer.OnMonitorEvent(ev, time.Now())
			}
			
			prof := e.profile.Get()
			text := e.speech.Generate(ev, e.cfg, reason, prof)
			e.lastEvent = ev
			e.AppendSpeechHistory(string(reason), text)
			
			e.notifier.Notify(types.Event{
				Type: "monitor_event",
				Payload: map[string]interface{}{
					"state":          string(ev.State),
					"task":           string(ev.Task),
					"mood":           string(ev.Mood),
					"speech":         text,
					"timestamp":      time.Now(),
					"using_fallback": e.speech.IsUsingFallback(),
					"profile":        map[string]string{"name": e.cfg.Name, "tone": e.cfg.Tone},
				},
			})

		case obs := <-observations:
			reason := e.reasonFromObservation(obs)
			if obs.Type == observer.ObsGitCommit {
				e.profile.RecordCommit()
			}
			
			prof := e.profile.Get()
			text := e.speech.Generate(e.lastEvent, e.cfg, reason, prof)
			e.AppendSpeechHistory(string(reason), text)
			
			e.notifier.Notify(types.Event{
				Type: "observation_event",
				Payload: map[string]interface{}{
					"state":          string(e.lastEvent.State),
					"task":           string(e.lastEvent.Task),
					"mood":           string(e.lastEvent.Mood),
					"speech":         text,
					"timestamp":      time.Now(),
					"using_fallback": e.speech.IsUsingFallback(),
					"profile":        map[string]string{"name": e.cfg.Name, "tone": e.cfg.Tone},
				},
			})
		}
	}
}

// StartupGreeting は起動時の挨拶を行う。
func (e *Engine) StartupGreeting(ctx context.Context) {
	log.Println("[ENGINE] StartupGreeting process started")
	time.Sleep(2 * time.Second)
	// Ollamaの準備を待つ（短縮）
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:11434/api/tags")
		if err == nil {
			resp.Body.Close()
			log.Println("[ENGINE] Ollama detected during startup")
			break
		}
		time.Sleep(1 * time.Second)
		if ctx.Err() != nil {
			return
		}
	}

	prof := e.profile.Get()
	reason := llm.ReasonInitSetup
	if e.cfg.SetupCompleted {
		reason = llm.ReasonGreeting
	}
	
	log.Printf("[ENGINE] Requesting greeting speech for reason: %s", reason)
	text := e.speech.Generate(e.lastEvent, e.cfg, reason, prof)
	log.Printf("[ENGINE] Greeting result: %s", text)
	
	e.AppendSpeechHistory(string(reason), text)
	
	e.notifier.Notify(types.Event{
		Type: "greeting_event",
		Payload: map[string]interface{}{
			"state":          string(e.lastEvent.State),
			"task":           string(e.lastEvent.Task),
			"mood":           string(e.lastEvent.Mood),
			"speech":         text,
			"timestamp":      time.Now(),
			"using_fallback": e.speech.IsUsingFallback(),
			"profile":        map[string]string{"name": e.cfg.Name, "tone": e.cfg.Tone},
		},
	})
}

// OnUserClick はユーザークリック時のセリフを生成する。
func (e *Engine) OnUserClick() {
	var prof profile.DevProfile
	if e.profile != nil {
		prof = e.profile.Get()
	}
	speech := e.speech.OnUserClick(e.lastEvent, e.cfg, prof)
	e.AppendSpeechHistory(string(llm.ReasonUserClick), speech)
	
	e.notifier.Notify(types.Event{
		Type: "click_event",
		Payload: map[string]interface{}{
			"state":          string(e.lastEvent.State),
			"task":           string(e.lastEvent.Task),
			"mood":           string(e.lastEvent.Mood),
			"speech":         speech,
			"timestamp":      time.Now(),
			"using_fallback": e.speech.IsUsingFallback(),
			"profile":        map[string]string{"name": e.cfg.Name, "tone": e.cfg.Tone},
		},
	})
}

// UpdateConfig はエンジンの設定を更新する。
func (e *Engine) UpdateConfig(cfg *config.Config) {
	e.cfg = cfg
	if e.speech != nil {
		e.speech.UpdateLLMConfig(cfg)
	}
}

func (e *Engine) reasonFromEvent(ev monitor.MonitorEvent) llm.Reason {
	switch ev.Event {
	case types.EventAISessionStarted:      return llm.ReasonAISessionStarted
	case types.EventAISessionActive:       return llm.ReasonAISessionActive
	case types.EventDevSessionStarted:     return llm.ReasonDevSessionStarted
	case types.EventDevEditing:            return llm.ReasonActiveEdit
	case types.EventGitActivity:           return llm.ReasonGitCommit
	case types.EventProductiveToolActivity: return llm.ReasonProductiveToolActivity
	case types.EventDocWriting:            return llm.ReasonDocWriting
	case types.EventLongInactivity:        return llm.ReasonLongInactivity
	}

	switch ev.State {
	case types.StateSuccess: return llm.ReasonSuccess
	case types.StateFail:    return llm.ReasonFail
	case types.StateCoding:  return llm.ReasonActiveEdit
	case types.StateIdle:    return llm.ReasonIdle
	}
	return llm.ReasonThinkingTick
}

func (e *Engine) reasonFromObservation(obs observer.DevObservation) llm.Reason {
	switch obs.Type {
	case observer.ObsGitCommit:    return llm.ReasonGitCommit
	case observer.ObsGitPush:      return llm.ReasonGitPush
	case observer.ObsGitAdd:       return llm.ReasonGitAdd
	case observer.ObsIdleStart:    return llm.ReasonIdle
	case observer.ObsNightWork:    return llm.ReasonNightWork
	case observer.ObsActiveEditing: return llm.ReasonActiveEdit
	}
	return llm.ReasonThinkingTick
}

func (e *Engine) AppendSpeechHistory(reason, text string) {
	if text == "" {
		return
	}
	cfgPath, err := config.DefaultConfigPath()
	if err != nil {
		return
	}
	dir := filepath.Dir(cfgPath)
	historyPath := filepath.Join(dir, "SPEECH_HISTORY.txt")

	f, err := os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	prefix := "[サクラ]"
	if reason != "" {
		prefix = fmt.Sprintf("[サクラ (%s)]", reason)
	}

	entry := fmt.Sprintf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), prefix, text)
	_, _ = f.WriteString(entry)
}
