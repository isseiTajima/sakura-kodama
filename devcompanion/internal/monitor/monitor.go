package monitor

import (
	"context"
	"log"
	"time"

	"devcompanion/internal/agent"
	"devcompanion/internal/behavior"
	"devcompanion/internal/config"
	"devcompanion/internal/context"
	"devcompanion/internal/debug/recorder"
	"devcompanion/internal/metrics"
	"devcompanion/internal/pipeline"
	"devcompanion/internal/plugin"
	"devcompanion/internal/sensor"
	"devcompanion/internal/session"
	"devcompanion/internal/types"
)

// MonitorEvent はパイプラインの最終出力を表す。
type MonitorEvent struct {
	State    types.ContextState   `json:"state"`
	Task     TaskType             `json:"task"`
	Mood     MoodType             `json:"mood"`
	Event    types.HighLevelEvent `json:"event"`
	Behavior types.Behavior       `json:"behavior"`
	Session  types.SessionState   `json:"session"`
	Context  types.ContextInfo    `json:"context"`
	Decision types.ContextDecision `json:"decision"`
	Details  string               `json:"details"`
}

// Monitor はパイプライン（Sensors -> Signals -> Context）を管理する。
type Monitor struct {
	cfg           *config.AppConfig
	agentRegistry *agent.Registry
	contextEngine *contextengine.Estimator
	sensors       []sensor.Sensor
	recorder      *recorder.Recorder
	pluginRegistry *plugin.Registry
	
	behaviorInferrer *behavior.Inferrer
	sessionTracker   *session.Tracker

	aiSessionActive  bool
	devSessionActive bool

	signals chan types.Signal
	events  chan MonitorEvent
}

func New(cfg *config.AppConfig, watchDir string) (*Monitor, error) {
	rec, err := recorder.New(true)
	if err != nil {
		rec, _ = recorder.New(false)
	}

	m := &Monitor{
		cfg:              cfg,
		agentRegistry:    agent.NewRegistry(),
		contextEngine:    contextengine.NewEstimator(),
		behaviorInferrer: behavior.NewInferrer(5 * time.Minute),
		sessionTracker:   session.NewTracker(),
		recorder:         rec,
		pluginRegistry:   plugin.NewRegistry(),
		signals:          make(chan types.Signal, 64),
		events:           make(chan MonitorEvent, 16),
	}

	// センサーの登録
	m.sensors = append(m.sensors, sensor.NewFSSensor(watchDir))
	m.sensors = append(m.sensors, sensor.NewProcessSensor([]string{"claude", "cursor", "vscode", "code", "iterm"}, 2*time.Second))

	if cfg != nil {
		m.contextEngine.SetWeights(cfg.SignalWeights)
	}

	return m, nil
}

func (m *Monitor) Run(ctx context.Context) {
	defer m.recorder.Close()

	for _, s := range m.sensors {
		go func(sn sensor.Sensor) {
			_ = sn.Run(ctx, m.signals)
		}(s)
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 定期的な状態更新（ひとり言用）
			pipeline.SafeExecute("HeartbeatLoop", func() {
				now := time.Now()
				// シグナルがない状態での現在の推定結果を取得
				decision := m.contextEngine.LastDecision
				if decision.State == "" {
					decision.State = types.StateIdle
				}
				
				currentBehavior := m.behaviorInferrer.Infer()
				currentSession := m.sessionTracker.Update(currentBehavior, now)

				m.events <- MonitorEvent{
					State:    decision.State,
					Behavior: currentBehavior,
					Session:  currentSession,
					Context: types.ContextInfo{
						State:      decision.State,
						Confidence: decision.Confidence,
						LastSignal: now,
					},
					Decision: decision,
				}
			})

		case sig := <-m.signals:
			log.Printf("[DEBUG] Monitor received signal: %+v", sig)
			pipeline.SafeExecute("MonitorLoop", func() {
				now := sig.Timestamp
				if now.IsZero() {
					now = time.Now()
				}

				metrics.IncrementSignalsReceived()
				m.recorder.Record(sig)
				m.pluginRegistry.NotifySignal(sig)
				
				highLevelEvent := m.classifySignal(sig)

				prevState := m.contextEngine.LastDecision.State
				
				// Context Engine で状態推定
				decision := m.contextEngine.ProcessSignal(sig)
				
				if decision.State != prevState {
					metrics.IncrementContextSwitch()
				}

				m.behaviorInferrer.AddSignal(sig)
				currentBehavior := m.behaviorInferrer.Infer()
				currentSession := m.sessionTracker.Update(currentBehavior, now)

				m.events <- MonitorEvent{
					State:    decision.State,
					Event:    highLevelEvent,
					Behavior: currentBehavior,
					Session:  currentSession,
					Context: types.ContextInfo{
						State:      decision.State,
						Confidence: decision.Confidence,
						LastSignal: now,
					},
					Decision: decision,
					Details:  sig.Value,
				}
			})
		}
	}
}

func (m *Monitor) Events() <-chan MonitorEvent {
	return m.events
}

func (m *Monitor) InjectSignal(sig types.Signal) {
	select {
	case m.signals <- sig:
	default:
	}
}

func (m *Monitor) classifySignal(sig types.Signal) types.HighLevelEvent {
	switch sig.Type {
	case types.SigProcessStarted:
		if sig.Source == types.SourceAgent {
			if !m.aiSessionActive {
				m.aiSessionActive = true
				return types.EventAISessionStarted
			}
			return types.EventAISessionActive
		}
	case types.SigFileModified, types.SigGitCommit:
		if !m.devSessionActive {
			m.devSessionActive = true
			return types.EventDevSessionStarted
		}
		if sig.Source == types.SourceGit {
			return types.EventGitActivity
		}
		return types.EventDevEditing
	}
	return ""
}
