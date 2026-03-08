package llm

import (
	"testing"
)

// TestAICLIClient_Generate_Success はai CLIの成功時の動作を検証。
func TestAICLIClient_Generate_Success(t *testing.T) {
	t.Parallel()

	// Given: プロンプトを echo で返すモック ai コマンド（テスト用）
	// 実装では ~/.bin/ai を呼び出すため、テスト時は別のパスを使用するか、
	// mock を設定します。ここは実装の仕様に応じて調整します。

	// ~理想的なテスト：AICLIClient に endpoint/path を設定可能にしておく~
	// When: Generate を呼ぶ
	// Then: stdout の内容がトリムされて返る

	// 注: 実装時に AICLIClient が ~/.bin/ai を直接呼び出す場合、
	// テスト環境でこのパスが存在しないため、テストはスキップまたは
	// システムのモックが必要になります。以下は仕様に合わせた
	// インターフェース定義です。
	t.Skip("AICLIClient implementation pending")
}

// TestAICLIClient_Generate_Timeout はタイムアウトの検出を検証。
func TestAICLIClient_Generate_Timeout(t *testing.T) {
	t.Parallel()

	// Given: 長時間実行されるコマンド、短いタイムアウト
	// When: Generate を呼ぶ
	// Then: タイムアウトエラーが返される

	t.Skip("AICLIClient implementation pending")
}

// TestAICLIClient_Generate_EmptyOutput は空出力の検出を検証。
func TestAICLIClient_Generate_EmptyOutput(t *testing.T) {
	t.Parallel()

	// Given: 空文字を出力するコマンド
	// When: Generate を呼ぶ
	// Then: エラーが返される

	t.Skip("AICLIClient implementation pending")
}

// TestAICLIClient_Generate_CommandNotFound はコマンド実行エラーを検証。
func TestAICLIClient_Generate_CommandNotFound(t *testing.T) {
	t.Parallel()

	// Given: 存在しないコマンドパス
	// When: Generate を呼ぶ
	// Then: エラーが返される

	t.Skip("AICLIClient implementation pending")
}

// TestAICLIClient_NewAICLIClient_DefaultTimeout はデフォルトタイムアウトを検証。
func TestAICLIClient_NewAICLIClient_DefaultTimeout(t *testing.T) {
	t.Parallel()

	// Note: aicli.go 実装後、以下のテストを記述します。
	// client := NewAICLIClient()
	// if client.timeout != 6*time.Second {
	//	t.Errorf("want timeout 6s, got %v", client.timeout)
	// }

	t.Skip("AICLIClient implementation pending")
}
