package llm

import (
	"sync"
	"time"
)

// SpeechState は発言状態と履歴を管理する。
type SpeechState struct {
	mu          sync.RWMutex
	recentLines []string        // 最大10件の発言履歴
	recentEvents []SpeechEvent  // 最大10件のイベント履歴
}

type SpeechEvent struct {
	Type string
	Time time.Time
}

// NewSpeechState は SpeechState を初期化する。
func NewSpeechState() *SpeechState {
	return &SpeechState{
		recentLines: make([]string, 0, 10),
		recentEvents: make([]SpeechEvent, 0, 10),
	}
}

// AddLine は発言を履歴に追加する。
// 最大10件を保持し、超過時は最古を削除する。
func (ss *SpeechState) AddLine(text string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.recentLines = append(ss.recentLines, text)
	if len(ss.recentLines) > 10 {
		ss.recentLines = ss.recentLines[1:]
	}
}

// AddEvent はイベントを履歴に追加する。
// 最大10件を保持し、超過時は最古を削除する。
func (ss *SpeechState) AddEvent(eventType string, timestamp time.Time) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	ss.recentEvents = append(ss.recentEvents, SpeechEvent{
		Type: eventType,
		Time: timestamp,
	})
	if len(ss.recentEvents) > 10 {
		ss.recentEvents = ss.recentEvents[1:]
	}
}

// IsDuplicate は直近の重複判定を行う。
func (ss *SpeechState) IsDuplicate(text string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// 直近2件のみ確認（応答性を優先）
	start := len(ss.recentLines) - 2
	if start < 0 {
		start = 0
	}

	for i := start; i < len(ss.recentLines); i++ {
		if ss.recentLines[i] == text {
			return true
		}
	}
	return false
}

// GetRecentLines は直近の発言履歴を返す。
func (ss *SpeechState) GetRecentLines(limit int) []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if limit > len(ss.recentLines) {
		limit = len(ss.recentLines)
	}
	return append([]string{}, ss.recentLines[len(ss.recentLines)-limit:]...)
}

// GetLastEventTime は最後のイベント時刻を返す。
// イベント履歴がない場合は zero time を返す。
func (ss *SpeechState) GetLastEventTime() time.Time {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if len(ss.recentEvents) == 0 {
		return time.Time{}
	}
	return ss.recentEvents[len(ss.recentEvents)-1].Time
}

// HasDeepFocus は3分以上イベントがないかを判定する。
func (ss *SpeechState) HasDeepFocus(now time.Time) bool {
	lastEventTime := ss.GetLastEventTime()
	if lastEventTime.IsZero() {
		return false // イベント履歴がない
	}
	return now.Sub(lastEventTime) >= 3*time.Minute
}

// GetRecentEvents は直近のイベント履歴を返す。
func (ss *SpeechState) GetRecentEvents(limit int) []SpeechEvent {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if limit > len(ss.recentEvents) {
		limit = len(ss.recentEvents)
	}
	events := make([]SpeechEvent, limit)
	copy(events, ss.recentEvents[len(ss.recentEvents)-limit:])
	return events
}

// EventCount はイベント履歴の件数を返す。
func (ss *SpeechState) EventCount() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.recentEvents)
}
