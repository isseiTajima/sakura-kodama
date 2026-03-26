package sensor

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"sakura-kodama/internal/types"
)

// aiAgentKeywords はプロセス名またはコマンドラインにマッチするキーワード一覧。
// 新しいAIコーディングツールが出たらここに追加するだけでよい。
var aiAgentKeywords = []string{
	// ターミナル型エージェント
	"claude", "aider", "codex", "devin", "cline",
	// エディタ内蔵型
	"cursor", "windsurf", "copilot", "continue", "cody",
	// その他
	"gpt-engineer", "open-interpreter", "smol-developer",
	"bolt", "lovable", "replit-agent",
}

// AIAgentSensor はプロセス名・コマンドラインをスキャンして
// AI コーディングエージェントの起動・終了を動的に検知する。
type AIAgentSensor struct {
	interval time.Duration
}

func NewAIAgentSensor(interval time.Duration) *AIAgentSensor {
	return &AIAgentSensor{interval: interval}
}

func (s *AIAgentSensor) Name() string {
	return "AIAgentSensor"
}

func (s *AIAgentSensor) Run(ctx context.Context, signals chan<- types.Signal) error {
	log.Printf("[SENSOR] Starting AIAgentSensor with interval %v", s.interval)
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// 前回検知したエージェント名のセット（重複発火防止）
	running := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			current := detectRunningAgents()

			for name := range current {
				if !running[name] {
					signals <- types.Signal{
						Type:      types.SigProcessStarted,
						Source:    types.SourceAgent,
						Value:     name,
						Timestamp: types.TimeToStr(time.Now()),
					}
				}
			}
			for name := range running {
				if !current[name] {
					signals <- types.Signal{
						Type:      types.SigProcessStopped,
						Source:    types.SourceAgent,
						Value:     name,
						Timestamp: types.TimeToStr(time.Now()),
					}
				}
			}
			running = current
		}
	}
}

// detectRunningAgents は現在起動中のAIエージェントを検出してマップで返す。
func detectRunningAgents() map[string]bool {
	found := make(map[string]bool)

	procs, err := process.Processes()
	if err != nil {
		return found
	}

	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		nameLower := strings.ToLower(name)

		// プロセス名でマッチ
		if kw := matchKeyword(nameLower); kw != "" {
			found[kw] = true
			continue
		}

		// コマンドライン全体でマッチ（python3 -m aider 等の場合）
		cmd, err := p.Cmdline()
		if err != nil {
			continue
		}
		cmdLower := strings.ToLower(cmd)
		if kw := matchKeyword(cmdLower); kw != "" {
			found[kw] = true
		}
	}

	return found
}

// matchKeyword はテキストにキーワードが含まれていればそのキーワードを返す。
func matchKeyword(text string) string {
	for _, kw := range aiAgentKeywords {
		if strings.Contains(text, kw) {
			return kw
		}
	}
	return ""
}
