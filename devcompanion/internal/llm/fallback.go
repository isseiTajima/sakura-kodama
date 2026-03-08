package llm

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	rnd   = rand.New(rand.NewSource(time.Now().UnixNano()))
	rndMu sync.Mutex
)

// SetSeed は乱数シードを固定する（テスト用）。
func SetSeed(seed int64) {
	rndMu.Lock()
	defer rndMu.Unlock()
	rnd = rand.New(rand.NewSource(seed))
}

// Reason はセリフ生成の理由を表す。
type Reason string

const (
	ReasonSuccess      Reason = "success"
	ReasonFail         Reason = "fail"
	ReasonThinkingTick Reason = "thinking_tick"
	ReasonUserClick    Reason = "user_click"

	// 開発観察コメントの Reason
	ReasonGitCommit  Reason = "git_commit"
	ReasonGitPush    Reason = "git_push"
	ReasonGitAdd     Reason = "git_add"
	ReasonIdle       Reason = "idle"
	ReasonNightWork  Reason = "night_work"
	ReasonActiveEdit Reason = "active_edit"
	ReasonInitSetup  Reason = "init_setup"
	ReasonGreeting   Reason = "greeting"

	// 新しい高レベルイベントの Reason
	ReasonAISessionStarted      Reason = "ai_session_started"
	ReasonAISessionActive       Reason = "ai_session_active"
	ReasonDevSessionStarted     Reason = "dev_session_started"
	ReasonProductiveToolActivity Reason = "productive_tool_activity"
	ReasonDocWriting            Reason = "doc_writing"
	ReasonLongInactivity        Reason = "long_inactivity"
	)

	// 固定セリフにバリエーションを持たせる
	var fallbackTexts = map[Reason][]string{
	ReasonInitSetup:    {"サクラだよ！もっとお喋りできるように、「考える力」を授けてね！[INSTALL_OLLAMA]"},
	ReasonGreeting:     {}, // speech.go 内で時間帯別に生成するため空にする
	ReasonThinkingTick: {"考え中…", "何しよっかな", "ふふっ"},
	ReasonSuccess:      {"よし、できた！", "いい感じですね！", "お疲れ様です！"},
	ReasonFail:         {"うっ…でも大丈夫", "次、頑張りましょう", "ドンマイです！"},
	ReasonUserClick:    {"サクラだよ！よろしくね！", "お疲れ様です！", "なにかお手伝いしましょうか？"},
	ReasonGitCommit:    {"お、コミットしましたね", "記録、大事ですね", "着々と進んでますね"},
	ReasonGitPush:      {"プッシュしましたね", "同期完了！", "お疲れ様です！"},
	ReasonGitAdd:       {"ステージング完了！", "コミットの準備、お疲れ様です", "いい感じですね"},
	ReasonIdle:         {"ちょっと休憩ですか？", "お茶でも飲みますか？", "まったりタイム"},
	ReasonNightWork:    {"もうこんな時間ですよ…", "夜更かしは体に毒ですよ", "頑張りすぎないでくださいね"},
	ReasonActiveEdit:   {"いっぱい書いてますね", "ノリノリですね！", "集中してますね"},
	ReasonAISessionStarted: {"AIエージェント、起動！", "相棒もやる気満々ですね", "一緒に頑張りましょう"},
	ReasonAISessionActive:  {"AIがバリバリ動いてますね", "頼もしいですね", "いい連携です"},
	ReasonDevSessionStarted: {"開発セッション、スタート！", "今日もバリバリ書きましょう", "気合入ってますね"},
	ReasonProductiveToolActivity: {"ツールを使いこなしてますね", "効率的でいいですね", "エンジニアって感じです！"},
	ReasonDocWriting:       {"ドキュメント作成中ですか？", "整理するのも大事ですよね", "執筆お疲れ様です"},
	ReasonLongInactivity:   {"おーい、生きてますかー？", "そろそろ休憩おしまい？", "リフレッシュできました？"},
	}


// FallbackSpeech はLLM呼び出し失敗時のテンプレートセリフを返す。
func FallbackSpeech(r Reason) string {
	if r == ReasonGreeting {
		h := time.Now().Hour()
		switch {
		case h >= 5 && h < 10:
			return "おはよう！今日も一日頑張ろうね！"
		case h >= 10 && h < 17:
			return "こんにちは！調子はどう？"
		case h >= 17 && h < 20:
			return "お疲れ様！そろそろ夕方だね"
		case h >= 20 && h < 23:
			return "こんばんは！夜の開発も捗ってる？"
		default:
			return "こんばんは！夜更かししすぎてない？"
		}
	}

	if texts, ok := fallbackTexts[r]; ok && len(texts) > 0 {
		rndMu.Lock()
		idx := rnd.Intn(len(texts))
		rndMu.Unlock()
		return texts[idx]
	}
	return "…"
}

// isTooManyRequestsError は error が HTTP 429 (Too Many Requests) を示すかチェックする。
func isTooManyRequestsError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "status 429")
}

// callGeminiCLI は ~/.bin/ai コマンドを呼び出して Gemini からセリフを生成する。
func callGeminiCLI(in OllamaInput) (string, error) {
	prompt, err := renderPrompt(in)
	if err != nil {
		return "", fmt.Errorf("render prompt: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}

	aiPath := filepath.Join(homeDir, ".bin", "ai")
	cmd := exec.Command(aiPath, "-p", prompt)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gemini cli: %w", err)
	}

	return strings.TrimSpace(out.String()), nil
}
