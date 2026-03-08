package types

import "time"

// --- Signal Layer ---

type SignalSource string

const (
	SourceProcess SignalSource = "process"
	SourceFS      SignalSource = "filesystem"
	SourceGit     SignalSource = "git"
	SourceAgent   SignalSource = "agent"
	SourceSystem  SignalSource = "system"
)

type SignalType string

const (
	SigProcessStarted   SignalType = "process_started"
	SigProcessStopped   SignalType = "process_stopped"
	SigFileModified     SignalType = "file_modified"
	SigManyFilesChanged SignalType = "many_files_changed"
	SigGitCommit        SignalType = "git_commit"
	SigLogHint          SignalType = "log_hint"
	SigIdleStart        SignalType = "idle_start"
	SigSystemWake       SignalType = "system_wake"
	SigSystemSleep      SignalType = "system_sleep"
)

type Signal struct {
	Type      SignalType   `json:"type"`
	Source    SignalSource `json:"source"`
	Value     string       `json:"value"`
	Message   string       `json:"message"`
	Timestamp time.Time    `json:"timestamp"`
}

// --- Behavior Layer ---

type BehaviorType string

const (
	BehaviorCoding          BehaviorType = "coding"
	BehaviorDebugging       BehaviorType = "debugging"
	BehaviorResearching     BehaviorType = "researching"
	BehaviorAIPairing       BehaviorType = "ai_pair_programming"
	BehaviorRefactoring     BehaviorType = "refactoring"
	BehaviorBreak           BehaviorType = "break"
	BehaviorProcrastinating BehaviorType = "procrastinating"
	BehaviorUnknown         BehaviorType = "unknown"
)

type Behavior struct {
	Type      BehaviorType `json:"type"`
	StartTime time.Time    `json:"start_time"`
	EndTime   time.Time    `json:"end_time"`
	Score     float64      `json:"score"`
}

// --- Session Layer ---

type Mode string

const (
	ModeDeepFocus     Mode = "deep_focus"
	ModeProductiveFlow Mode = "productive_flow"
	ModeStruggling     Mode = "struggling"
	ModeCasualWork     Mode = "casual_work"
	ModeOnBreak        Mode = "on_break"
	ModeIdle           Mode = "idle"
)

type SessionState struct {
	Mode           Mode      `json:"mode"`
	StartTime      time.Time `json:"start_time"`
	LastActivity   time.Time `json:"last_activity"`
	FocusLevel     float64   `json:"focus_level"`
	ProgressScore  int       `json:"progress_score"`
}

// --- Context Layer ---

type ContextState string

const (
	StateIdle             ContextState = "IDLE"
	StateCoding           ContextState = "CODING"
	StateAIPairing        ContextState = "AI_PAIR_PROGRAMMING"
	StateDeepWork         ContextState = "DEEP_WORK"
	StateStuck            ContextState = "STUCK"
	StateProcrastinating  ContextState = "PROCRASTINATING"
	StateSessionEnding    ContextState = "SESSION_ENDING"
	StateSuccess          ContextState = "SUCCESS"
	StateFail             ContextState = "FAIL"
)

type ContextInfo struct {
	State      ContextState `json:"state"`
	Confidence float64      `json:"confidence"`
	StartTime  time.Time    `json:"start_time"`
	LastSignal time.Time    `json:"last_signal"`
}

type ContextDecision struct {
	State      ContextState `json:"state"`
	Confidence float64      `json:"confidence"`
	Signals    []SignalType `json:"signals"`
	Reasons    []string     `json:"reasons"`
}

// --- Event Layer ---

type HighLevelEvent string

const (
	EventAISessionStarted       HighLevelEvent = "AI_SESSION_STARTED"
	EventAISessionActive        HighLevelEvent = "AI_SESSION_ACTIVE"
	EventDevSessionStarted      HighLevelEvent = "DEV_SESSION_STARTED"
	EventDevEditing             HighLevelEvent = "DEV_EDITING"
	EventGitActivity            HighLevelEvent = "GIT_ACTIVITY"
	EventProductiveToolActivity HighLevelEvent = "PRODUCTIVE_TOOL_ACTIVITY"
	EventDocWriting             HighLevelEvent = "DOC_WRITING"
	EventLongInactivity         HighLevelEvent = "LONG_INACTIVITY"
)

// Event はシステム全体で流通する汎用的なイベント構造体。
type Event struct {
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// --- Persona Layer ---

type PersonaStyle string

const (
	StyleSoft      PersonaStyle = "soft"
	StyleEnergetic PersonaStyle = "energetic"
	StyleStrict    PersonaStyle = "strict"
)
