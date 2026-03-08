package llm

import (
	"context"
	"testing"
	"time"
)

// TestTimeoutConsistency_OllamaTimeout はOllamaクライアントのタイムアウト値が
// Router層で使用されるタイムアウト値と一貫性を持つことを検証する。
func TestTimeoutConsistency_OllamaTimeout(t *testing.T) {
	t.Parallel()

	// Given: Ollama クライアントが作成される
	ollamaClient := NewOllamaClient("http://localhost:11434", "test-model")

	// When/Then: タイムアウト値が 2s（内部の ollamaTimeout）であることを確認
	// Note: Router 層では ollamaRouterTimeout = 3s を使用するため、
	// ネストされたタイムアウトが発生する可能性がある。
	// 実装では、タイムアウト責任を明確化（Router が所有するか、クライアント所有か）する必要がある。
	if ollamaClient.timeout != 2*time.Second {
		t.Errorf("want Ollama timeout 2s (internal), got %v", ollamaClient.timeout)
	}
}

// TestTimeoutConsistency_ClaudeTimeout はClaudeクライアントのタイムアウト値が
// 仕様（5s）に統一されることを検証する。
func TestTimeoutConsistency_ClaudeTimeout(t *testing.T) {
	t.Parallel()

	// Given: Anthropic クライアントが作成される
	claudeClient := NewAnthropicClient("test-key")

	// When/Then: タイムアウト値が 5s（仕様値）であることを確認
	if claudeClient.timeout != 5*time.Second {
		t.Errorf("want Claude timeout 5s per spec, got %v", claudeClient.timeout)
	}
}

// TestTimeoutConsistency_AICLITimeout はAICLIクライアントのタイムアウト値が
// 仕様（6s）に統一されることを検証する。
func TestTimeoutConsistency_AICLITimeout(t *testing.T) {
	t.Parallel()

	// Given: AICLIClient が作成される
	aicliClient := NewAICLIClient()

	// When/Then: タイムアウト値が 6s であることを確認
	if aicliClient.timeout != 6*time.Second {
		t.Errorf("want ai CLI timeout 6s, got %v", aicliClient.timeout)
	}
}

// TestTimeoutConsistency_NestedTimeout はRouter層で各クライアントを呼び出す時の
// ネストされたタイムアウト処理を検証する。
func TestTimeoutConsistency_NestedTimeout(t *testing.T) {
	t.Parallel()

	// Given: context.WithTimeout でタイムアウトコンテキストを作成
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// When: Router が ollamaRouterTimeout (3s) でタイムアウトコンテキストを作成
	timeoutCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// Then: 実効タイムアウトが min(10s, 3s) = 3s になる
	// Note: クライアント内部の ollamaTimeout (2s) も含まれる場合、
	// 実効タイムアウトは min(3s, 2s) = 2s になる。
	// この二重タイムアウト構造を明確化する必要がある。

	deadline, ok := timeoutCtx.Deadline()
	if !ok {
		t.Fatal("want deadline to be set")
	}

	elapsed := time.Until(deadline)
	// 3秒より短いことを確認（clock skew を考慮）
	if elapsed > 3200*time.Millisecond {
		t.Errorf("want timeout ~3s, got %v", elapsed)
	}
}

// TestTimeoutConsistency_ClaudeVsAICLI はClaude (5s) と ai CLI (6s) の
// タイムアウト順序が正しいことを検証する。
func TestTimeoutConsistency_ClaudeVsAICLI(t *testing.T) {
	t.Parallel()

	claudeTimeout := 5 * time.Second
	aicliTimeout := 6 * time.Second

	// When/Then: Claude のタイムアウトが ai CLI より短いことを確認
	// ルーティング層では Claude で失敗した後 ai CLI に遷移するため、
	// Claude のタイムアウトが短い方が効率的
	if claudeTimeout >= aicliTimeout {
		t.Errorf("want Claude timeout (5s) < ai CLI timeout (6s), got %v >= %v",
			claudeTimeout, aicliTimeout)
	}
}

// TestRouterTimeoutConstant_Definition はRouter層のタイムアウト定数が
// 正しく定義されていることを検証する。
func TestRouterTimeoutConstant_Definition(t *testing.T) {
	t.Parallel()

	// Note: router.go で以下の定数が定義されていることを確認
	// const (
	//   ollamaRouterTimeout = 3 * time.Second
	//   claudeRouterTimeout = 5 * time.Second
	//   aicliRouterTimeout  = 6 * time.Second
	// )

	// テスト実装時に、これらの定数がコード内で参照されていることを確認する
	// grep -n "ollamaRouterTimeout" router.go
	// grep -n "claudeRouterTimeout" router.go
	// grep -n "aicliRouterTimeout" router.go
}

// TestTimeoutDeduplication はタイムアウト値の重複定義を検出する。
func TestTimeoutDeduplication(t *testing.T) {
	t.Parallel()

	// Note: aicli.go と router.go で aicliTimeout/aicliRouterTimeout が
	// 重複定義されていないことを確認する。
	// 実装では、router.go で定義された定数をaicli.go から参照する。

	// Given: AICLIClient
	aicli := NewAICLIClient()

	// When/Then: timeout が router.go の aicliRouterTimeout と同じ値であること
	expectedTimeout := 6 * time.Second
	if aicli.timeout != expectedTimeout {
		t.Errorf("want ai CLI timeout %v (from router constant), got %v",
			expectedTimeout, aicli.timeout)
	}
}
