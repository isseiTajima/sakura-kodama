package monitor

import (
	"regexp"
	"strings"
	"time"
)

// TaskType はClaudeが現在行っている作業の種類を表す。
type TaskType string

const (
	TaskPlan            TaskType = "Plan"
	TaskGenerateCode    TaskType = "GenerateCode"
	TaskRunTests        TaskType = "RunTests"
	TaskLintFormat      TaskType = "LintFormat"
	TaskDebug           TaskType = "Debug"
	TaskFixFailingTests TaskType = "FixFailingTests"
)

const (
	ringBufferCap    = 20
	silenceBonus     = 2
	recencyDecay     = 0.3
	historyThreshold = 10
	historyBoost     = 0.5
	stickyWeight     = 1.5
)

var exitCodePattern = regexp.MustCompile(`(?i)exit(?:ed)?[^0-9-]*code[^0-9-]*(-?[0-9]+)`)

// TaskInferrer はシグナル行バッファを保持し、最もスコアの高いTaskを推論する。
type TaskInferrer struct {
	buffer []string
}

// NewTaskInferrer は TaskInferrer を初期化する。
func NewTaskInferrer() *TaskInferrer {
	return &TaskInferrer{
		buffer: make([]string, 0, ringBufferCap),
	}
}

// AddLine はシグナル行をリングバッファに追加する。
// バッファが満杯の場合は最古の行を削除する。
func (ti *TaskInferrer) AddLine(line string) {
	if len(ti.buffer) >= ringBufferCap {
		ti.buffer = ti.buffer[1:]
	}
	ti.buffer = append(ti.buffer, line)
}

// RecentContext は最新のログ数行を結合して返す。
func (ti *TaskInferrer) RecentContext() string {
	if len(ti.buffer) == 0 {
		return ""
	}
	start := len(ti.buffer) - 3
	if start < 0 {
		start = 0
	}
	return strings.Join(ti.buffer[start:], "\n")
}

// Infer は現在のバッファと無音時間からTaskを推論する。
// 全スコアが0の場合は TaskPlan をデフォルトとして返す。
func (ti *TaskInferrer) Infer(silenceDuration time.Duration) TaskType {
	scores := map[TaskType]float64{}
	occurrences := map[TaskType]int{}
	weight := 1.0

	var latestTask TaskType
	var latestTaskOK bool
	if len(ti.buffer) > 0 {
		if task, ok := taskFromLine(ti.buffer[len(ti.buffer)-1]); ok {
			latestTask = task
			latestTaskOK = true
		}
	}

	for i := len(ti.buffer) - 1; i >= 0; i-- {
		line := ti.buffer[i]
		for task, base := range ti.scoreLine(line) {
			scores[task] += float64(base) * weight
			occurrences[task]++
		}
		weight *= recencyDecay
	}

	for task, count := range occurrences {
		if count > historyThreshold {
			extra := float64(count-historyThreshold) * historyBoost
			scores[task] += extra
		}
		scores[task] += float64(count) * stickyWeight
	}

	if silenceDuration >= 5*time.Second {
		scores[TaskPlan] += silenceBonus
	}

	best := argmax(scores)

	if latestTaskOK {
		switch latestTask {
		case TaskFixFailingTests:
			return latestTask
		case TaskRunTests, TaskGenerateCode:
			return latestTask
		}
	}

	return best
}

func (ti *TaskInferrer) scoreLine(line string) map[TaskType]int {
	results := map[TaskType]int{}
	if line == "" {
		return results
	}

	if strings.Contains(line, "go test") {
		results[TaskRunTests] += 3
	}
	if strings.Contains(line, "FAIL") {
		results[TaskRunTests] += 2
	}
	if strings.Contains(line, "panic") {
		results[TaskDebug] += 4
	}
	if strings.Contains(line, "lint") || strings.Contains(line, "fmt") {
		results[TaskLintFormat] += 2
	}
	if strings.Contains(line, "generate") || strings.Contains(line, "写") || strings.Contains(line, "実装") || strings.Contains(line, "Gemini") {
		results[TaskGenerateCode] += 2
	}
	if matches := exitCodePattern.FindStringSubmatch(line); len(matches) == 2 {
		if matches[1] != "0" {
			results[TaskDebug] += 3
		}
	}

	return results
}

// argmax はスコアマップの最大値を持つTaskを返す。
// 全スコアが0またはマップが空の場合は TaskPlan を返す。
func argmax(scores map[TaskType]float64) TaskType {
	best := TaskPlan
	bestScore := 0.0
	for task, score := range scores {
		if score > bestScore {
			bestScore = score
			best = task
		}
	}
	return best
}

func taskFromLine(line string) (TaskType, bool) {
	switch {
	case line == "":
		return TaskPlan, false
	case strings.Contains(line, "panic"):
		return TaskDebug, true
	case strings.Contains(line, "FAIL"):
		return TaskFixFailingTests, true
	case exitCodePattern.MatchString(line):
		return TaskDebug, true
	case strings.Contains(line, "go test"):
		return TaskRunTests, true
	case strings.Contains(line, "lint") || strings.Contains(line, "fmt"):
		return TaskLintFormat, true
	case strings.Contains(line, "generate") || strings.Contains(line, "写") || strings.Contains(line, "実装") || strings.Contains(line, "Gemini"):
		return TaskGenerateCode, true
	default:
		return TaskPlan, false
	}
}
