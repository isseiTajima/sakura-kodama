package profile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// newTestStore は一時ファイルを使って ProfileStore を作成するヘルパー。
func newTestStore(t *testing.T) *ProfileStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dev_profile.json")
	store, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	return store
}

// --- 初期化 ---

func TestNewProfileStore_NewFile_DefaultValues(t *testing.T) {
	// Given: 存在しないファイルパス
	path := filepath.Join(t.TempDir(), "dev_profile.json")

	// When: 新規作成
	store, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}

	// Then: デフォルト値で初期化される
	prof := store.Get()
	if prof.CommitFrequency != "low" {
		t.Errorf("want commit_frequency=%q initially, got %q", "low", prof.CommitFrequency)
	}
	if prof.BuildFailRate != "low" {
		t.Errorf("want build_fail_rate=%q initially, got %q", "low", prof.BuildFailRate)
	}
	if prof.NightCoder {
		t.Error("want night_coder=false initially")
	}
}

// --- commit_frequency の閾値テスト ---

func TestRecordCommit_0Commits_LowFrequency(t *testing.T) {
	// Given: コミット未記録
	store := newTestStore(t)

	// Then: low（0〜1 = low）
	if got := store.Get().CommitFrequency; got != "low" {
		t.Errorf("want low for 0 commits, got %q", got)
	}
}

func TestRecordCommit_1Commit_LowFrequency(t *testing.T) {
	// Given: コミット 1 回
	store := newTestStore(t)
	store.RecordCommit()

	// Then: low（0〜1 = low）
	if got := store.Get().CommitFrequency; got != "low" {
		t.Errorf("want low for 1 commit, got %q", got)
	}
}

func TestRecordCommit_2Commits_MediumFrequency(t *testing.T) {
	// Given: コミット 2 回（medium の下限境界）
	store := newTestStore(t)
	store.RecordCommit()
	store.RecordCommit()

	// Then: medium（2〜4 = medium）
	if got := store.Get().CommitFrequency; got != "medium" {
		t.Errorf("want medium for 2 commits, got %q", got)
	}
}

func TestRecordCommit_4Commits_MediumFrequency(t *testing.T) {
	// Given: コミット 4 回（medium の上限境界）
	store := newTestStore(t)
	for i := 0; i < 4; i++ {
		store.RecordCommit()
	}

	// Then: medium
	if got := store.Get().CommitFrequency; got != "medium" {
		t.Errorf("want medium for 4 commits, got %q", got)
	}
}

func TestRecordCommit_5Commits_HighFrequency(t *testing.T) {
	// Given: コミット 5 回（high の下限境界）
	store := newTestStore(t)
	for i := 0; i < 5; i++ {
		store.RecordCommit()
	}

	// Then: high（5+ = high）
	if got := store.Get().CommitFrequency; got != "high" {
		t.Errorf("want high for 5 commits, got %q", got)
	}
}

func TestRecordCommit_10Commits_HighFrequency(t *testing.T) {
	// Given: コミット 10 回
	store := newTestStore(t)
	for i := 0; i < 10; i++ {
		store.RecordCommit()
	}

	// Then: high
	if got := store.Get().CommitFrequency; got != "high" {
		t.Errorf("want high for 10 commits, got %q", got)
	}
}

// --- build_fail_rate の閾値テスト ---

func TestBuildFailRate_NoBuilds_Low(t *testing.T) {
	// Given: ビルド記録なし
	store := newTestStore(t)

	// Then: low（0% = low）
	if got := store.Get().BuildFailRate; got != "low" {
		t.Errorf("want low for 0 builds, got %q", got)
	}
}

func TestBuildFailRate_AllSuccess_Low(t *testing.T) {
	// Given: 成功のみ（失敗率 0%）
	store := newTestStore(t)
	store.RecordBuildSuccess()
	store.RecordBuildSuccess()
	store.RecordBuildSuccess()

	// Then: low
	if got := store.Get().BuildFailRate; got != "low" {
		t.Errorf("want low for 0%% fail rate, got %q", got)
	}
}

func TestBuildFailRate_30Percent_Low(t *testing.T) {
	// Given: 3 失敗 / 10 合計 = 30%（low の上限境界）
	store := newTestStore(t)
	for i := 0; i < 3; i++ {
		store.RecordBuildFail()
	}
	for i := 0; i < 7; i++ {
		store.RecordBuildSuccess()
	}

	// Then: low（0〜30% = low）
	if got := store.Get().BuildFailRate; got != "low" {
		t.Errorf("want low for 30%% fail rate, got %q", got)
	}
}

func TestBuildFailRate_31Percent_Medium(t *testing.T) {
	// Given: 31 失敗 / 100 合計 = 31%（medium の下限境界）
	store := newTestStore(t)
	for i := 0; i < 31; i++ {
		store.RecordBuildFail()
	}
	for i := 0; i < 69; i++ {
		store.RecordBuildSuccess()
	}

	// Then: medium（31〜60% = medium）
	if got := store.Get().BuildFailRate; got != "medium" {
		t.Errorf("want medium for 31%% fail rate, got %q", got)
	}
}

func TestBuildFailRate_60Percent_Medium(t *testing.T) {
	// Given: 3 失敗 / 5 合計 = 60%（medium の上限境界）
	store := newTestStore(t)
	for i := 0; i < 3; i++ {
		store.RecordBuildFail()
	}
	for i := 0; i < 2; i++ {
		store.RecordBuildSuccess()
	}

	// Then: medium
	if got := store.Get().BuildFailRate; got != "medium" {
		t.Errorf("want medium for 60%% fail rate, got %q", got)
	}
}

func TestBuildFailRate_61Percent_High(t *testing.T) {
	// Given: 61 失敗 / 100 合計 = 61%（high の下限境界）
	store := newTestStore(t)
	for i := 0; i < 61; i++ {
		store.RecordBuildFail()
	}
	for i := 0; i < 39; i++ {
		store.RecordBuildSuccess()
	}

	// Then: high（61%+ = high）
	if got := store.Get().BuildFailRate; got != "high" {
		t.Errorf("want high for 61%% fail rate, got %q", got)
	}
}

func TestBuildFailRate_AllFail_High(t *testing.T) {
	// Given: 失敗のみ（失敗率 100%）
	store := newTestStore(t)
	store.RecordBuildFail()
	store.RecordBuildFail()
	store.RecordBuildFail()

	// Then: high
	if got := store.Get().BuildFailRate; got != "high" {
		t.Errorf("want high for 100%% fail rate, got %q", got)
	}
}

// --- night_coder の閾値テスト ---

func TestNightCoder_NoActivity_False(t *testing.T) {
	// Given: アクティビティ記録なし
	store := newTestStore(t)

	// Then: false
	if store.Get().NightCoder {
		t.Error("want night_coder=false with no activity")
	}
}

func TestNightCoder_29Percent_False(t *testing.T) {
	// Given: 深夜 29 / 合計 100 = 29%（30% 未満）
	store := newTestStore(t)
	dayTime := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
	nightTime := time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC)

	for i := 0; i < 71; i++ {
		store.RecordActivity(dayTime)
	}
	for i := 0; i < 29; i++ {
		store.RecordActivity(nightTime)
	}

	// Then: false（29% < 30%）
	if store.Get().NightCoder {
		t.Error("want night_coder=false for 29% night activity")
	}
}

func TestNightCoder_Exactly30Percent_True(t *testing.T) {
	// Given: 深夜 30 / 合計 100 = 30%（境界値、true に変わる）
	store := newTestStore(t)
	dayTime := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
	nightTime := time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC)

	for i := 0; i < 70; i++ {
		store.RecordActivity(dayTime)
	}
	for i := 0; i < 30; i++ {
		store.RecordActivity(nightTime)
	}

	// Then: true（>= 30%）
	if !store.Get().NightCoder {
		t.Error("want night_coder=true for exactly 30% night activity (boundary)")
	}
}

func TestNightCoder_50Percent_True(t *testing.T) {
	// Given: 深夜 50 / 合計 100 = 50%
	store := newTestStore(t)
	nightTime := time.Date(2024, 1, 1, 23, 30, 0, 0, time.UTC)
	dayTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	for i := 0; i < 50; i++ {
		store.RecordActivity(nightTime)
	}
	for i := 0; i < 50; i++ {
		store.RecordActivity(dayTime)
	}

	// Then: true
	if !store.Get().NightCoder {
		t.Error("want night_coder=true for 50% night activity")
	}
}

func TestNightCoder_AllNight_True(t *testing.T) {
	// Given: 全アクティビティが深夜帯
	store := newTestStore(t)
	nightTime := time.Date(2024, 1, 1, 2, 0, 0, 0, time.UTC) // 02:00

	for i := 0; i < 100; i++ {
		store.RecordActivity(nightTime)
	}

	// Then: true
	if !store.Get().NightCoder {
		t.Error("want night_coder=true for 100% night activity")
	}
}

func TestNightCoder_NightBoundaryHour4_CountsAsNight(t *testing.T) {
	// Given: 04:59（深夜帯内）のアクティビティのみ
	store := newTestStore(t)
	lateNight := time.Date(2024, 1, 1, 4, 59, 0, 0, time.UTC)

	for i := 0; i < 100; i++ {
		store.RecordActivity(lateNight)
	}

	// Then: true（04:59 は深夜帯内）
	if !store.Get().NightCoder {
		t.Error("want night_coder=true for 04:59 activity (within night range)")
	}
}

func TestNightCoder_NightBoundaryHour5_CountsAsDay(t *testing.T) {
	// Given: 05:00（深夜帯外）のアクティビティのみ
	store := newTestStore(t)
	morning := time.Date(2024, 1, 1, 5, 0, 0, 0, time.UTC)

	for i := 0; i < 100; i++ {
		store.RecordActivity(morning)
	}

	// Then: false（05:00 は深夜帯外）
	if store.Get().NightCoder {
		t.Error("want night_coder=false for 05:00 activity (outside night range)")
	}
}

// --- LastActive ---

func TestRecordActivity_UpdatesLastActive(t *testing.T) {
	// Given: 初期状態
	store := newTestStore(t)
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// When: アクティビティ記録
	store.RecordActivity(now)

	// Then: LastActive が空でない
	prof := store.Get()
	if prof.LastActive == "" {
		t.Error("want non-empty LastActive after RecordActivity")
	}
}

// --- JSON ラウンドトリップ ---

func TestProfileStore_JSONRoundtrip_PreservesStats(t *testing.T) {
	// Given: 複数の記録を持つ store
	dir := t.TempDir()
	path := filepath.Join(dir, "dev_profile.json")

	store1, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}

	// コミット 5 回 → high
	for i := 0; i < 5; i++ {
		store1.RecordCommit()
	}
	// ビルド失敗 2、成功 1 → 66% → high
	store1.RecordBuildFail()
	store1.RecordBuildFail()
	store1.RecordBuildSuccess()
	// 深夜 50 / 昼間 50 → 50% → true
	nightTime := time.Date(2024, 1, 1, 23, 0, 0, 0, time.UTC)
	dayTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 50; i++ {
		store1.RecordActivity(nightTime)
	}
	for i := 0; i < 50; i++ {
		store1.RecordActivity(dayTime)
	}

	if err := store1.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// When: ファイルから再ロード
	store2, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore (reload): %v", err)
	}

	// Then: 値が引き継がれている
	prof := store2.Get()
	if prof.CommitFrequency != "high" {
		t.Errorf("want high commit_frequency after reload, got %q", prof.CommitFrequency)
	}
	if prof.BuildFailRate != "high" {
		t.Errorf("want high build_fail_rate after reload, got %q", prof.BuildFailRate)
	}
	if !prof.NightCoder {
		t.Error("want night_coder=true after reload")
	}
}

func TestProfileStore_JSONFileHasRequiredFields(t *testing.T) {
	// Given: 記録済み store
	dir := t.TempDir()
	path := filepath.Join(dir, "dev_profile.json")

	store, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore: %v", err)
	}
	store.RecordCommit()
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	store.RecordActivity(now)
	if err := store.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	// When: ファイルを読み込んで JSON パース
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON in dev_profile.json: %v", err)
	}

	// Then: DevProfile フィールドが全て含まれている
	requiredFields := []string{"night_coder", "commit_frequency", "build_fail_rate", "last_active"}
	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("want field %q in JSON output, got keys: %v", field, decoded)
		}
	}
}

func TestProfileStore_ExistingFile_IncrementalUpdate(t *testing.T) {
	// Given: 既存ファイル（コミット 4 回記録済み）
	dir := t.TempDir()
	path := filepath.Join(dir, "dev_profile.json")

	store1, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore first: %v", err)
	}
	for i := 0; i < 4; i++ {
		store1.RecordCommit()
	}
	if err := store1.Stop(); err != nil {
		t.Fatalf("Stop first: %v", err)
	}

	// When: 2 回目のセッションでさらに 1 回コミット（計 5 回）
	store2, err := NewProfileStore(path)
	if err != nil {
		t.Fatalf("NewProfileStore second: %v", err)
	}
	store2.RecordCommit()

	// Then: 累積カウントが引き継がれ high になる（4+1=5）
	if got := store2.Get().CommitFrequency; got != "high" {
		t.Errorf("want high after incremental 5th commit across sessions, got %q", got)
	}
}
