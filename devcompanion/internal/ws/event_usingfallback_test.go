package ws

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// TestEvent_IncludesUsingFallbackField は Event 構造体に UsingFallback フィールドがあるか確認。
func TestEvent_IncludesUsingFallbackField(t *testing.T) {
	t.Parallel()

	// Given: Event 構造体の型情報
	typ := reflect.TypeOf(Event{})

	// When: UsingFallback フィールドを取得
	usingFallbackField, ok := typ.FieldByName("UsingFallback")

	// Then: フィールドが存在し、JSON タグが正しい
	if !ok {
		t.Fatal("Event struct must include UsingFallback field")
	}
	if tag := usingFallbackField.Tag.Get("json"); tag != "using_fallback" {
		t.Fatalf("UsingFallback field must have json tag \"using_fallback\", got %q", tag)
	}
	if usingFallbackField.Type.Kind() != reflect.Bool {
		t.Fatalf("UsingFallback field must be bool, got %v", usingFallbackField.Type.Kind())
	}
}

// TestEvent_UsingFallbackMarshal は Event の JSON マーシャリング時に using_fallback フィールドが含まれるか確認。
func TestEvent_UsingFallbackMarshal(t *testing.T) {
	t.Parallel()

	// Given: UsingFallback=true の Event
	event := Event{
		State:        "thinking",
		Task:         "plan",
		Mood:         "happy",
		Speech:       "がんばろう",
		Timestamp:    time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		UsingFallback: true,
		Profile: EventProfile{
			Name: "テスト",
			Tone: "genki",
		},
	}

	// When: JSON にマーシャル
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Then: JSON に using_fallback フィールドが含まれている
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	usingFallback, ok := result["using_fallback"]
	if !ok {
		t.Error("JSON output must include 'using_fallback' field")
	}
	if usingFallback != true {
		t.Errorf("want using_fallback=true, got %v", usingFallback)
	}
}

// TestEvent_UsingFallbackUnmarshal は JSON から Event へのアンマーシャリング時にフィールドが正しく読み込まれるか確認。
func TestEvent_UsingFallbackUnmarshal(t *testing.T) {
	t.Parallel()

	// Given: JSON string with using_fallback=true
	jsonStr := `{
		"state":"thinking",
		"task":"plan",
		"mood":"happy",
		"speech":"テスト",
		"timestamp":"2025-01-01T12:00:00Z",
		"using_fallback":true,
		"profile":{"name":"テスト","tone":"calm"}
	}`

	// When: JSON からアンマーシャル
	var event Event
	if err := json.Unmarshal([]byte(jsonStr), &event); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	// Then: UsingFallback フィールドが正しく読み込まれている
	if !event.UsingFallback {
		t.Errorf("want UsingFallback=true, got %v", event.UsingFallback)
	}
}

// TestEvent_UsingFallbackFalse は Event の UsingFallback=false のテスト。
func TestEvent_UsingFallbackFalse(t *testing.T) {
	t.Parallel()

	// Given: UsingFallback=false の Event
	event := Event{
		State:         "idle",
		Task:          "plan",
		Mood:          "calm",
		Speech:        "待ってる",
		Timestamp:     time.Now(),
		UsingFallback: false,
		Profile: EventProfile{
			Name: "テスト",
			Tone: "calm",
		},
	}

	// When: JSON にマーシャル
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}

	// Then: JSON の using_fallback は false
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	usingFallback, ok := result["using_fallback"]
	if !ok {
		t.Error("JSON output must include 'using_fallback' field")
	}
	if usingFallback != false {
		t.Errorf("want using_fallback=false, got %v", usingFallback)
	}
}

// TestEvent_AllFieldsPresent は Event 構造体のすべての必須フィールドが存在するか確認。
func TestEvent_AllFieldsPresent(t *testing.T) {
	t.Parallel()

	// Given: Event 構造体
	typ := reflect.TypeOf(Event{})

	// When: 必須フィールドのリストをチェック
	requiredFields := []string{"State", "Task", "Mood", "Speech", "Timestamp", "Profile", "UsingFallback"}

	// Then: すべてのフィールドが存在
	for _, fieldName := range requiredFields {
		if _, ok := typ.FieldByName(fieldName); !ok {
			t.Errorf("Event struct must include %s field", fieldName)
		}
	}
}

// TestEvent_DefaultUsingFallbackValue は Event の UsingFallback フィールドのデフォルト値テスト。
func TestEvent_DefaultUsingFallbackValue(t *testing.T) {
	t.Parallel()

	// Given: 初期化されていない Event（デフォルト値）
	event := Event{}

	// When: UsingFallback フィールドの値を確認
	// Then: bool のデフォルト値は false
	if event.UsingFallback {
		t.Error("want UsingFallback=false by default, got true")
	}
}
