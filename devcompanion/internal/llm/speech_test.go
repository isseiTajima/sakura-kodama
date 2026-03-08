package llm

import (
	"testing"
	"time"

	"devcompanion/internal/config"
	"devcompanion/internal/monitor"
	"devcompanion/internal/profile"
	"devcompanion/internal/types"
)

func testConfig() *config.Config {
	return config.DefaultConfig()
}

func TestFrequencyController_ThinkingTick_SpeaksAfterMinInterval(t *testing.T) {
	fc := NewFrequencyController()
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := testConfig()

	fc.RecordSpeak(ReasonThinkingTick, types.StateDeepWork, cfg, baseTime)

	// ThinkingTick interval depends on SpeechFrequency
	now := baseTime.Add(16 * time.Minute) 
	can := fc.ShouldSpeak(ReasonThinkingTick, types.StateDeepWork, cfg, now)

	if !can {
		t.Error("want true after long interval, got false")
	}
}

func TestFrequencyController_UserClick_AlwaysSpeaks(t *testing.T) {
	fc := NewFrequencyController()
	cfg := testConfig()

	can := fc.ShouldSpeak(ReasonUserClick, types.StateIdle, cfg, time.Now())

	if !can {
		t.Error("want true for user click, got false")
	}
}

func TestSpeechGenerator_FallbackBackend_ReturnsTemplateText(t *testing.T) {
	cfg := testConfig()
	sg := NewSpeechGenerator(cfg)
	prof := profile.DevProfile{}

	event := monitor.MonitorEvent{
		State: types.StateIdle,
	}

	speech := sg.Generate(event, cfg, ReasonUserClick, prof)

	if speech == "" {
		t.Error("want non-empty speech")
	}
}
