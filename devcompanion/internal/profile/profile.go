package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Relationship はユーザーとサクラの親密度モデル。
type Relationship struct {
	Level                  int     `json:"relationship_level"`        // 0-100
	Trust                  int     `json:"trust"`                     // 0-100
	EncouragementPreference string  `json:"encouragement_preference"` // "gentle"|"strict"
}

// DevProfile はセリフ生成に渡す開発者プロファイル。
type DevProfile struct {
	NightCoder      bool         `json:"night_coder"`
	CommitFrequency string       `json:"commit_frequency"` // "low"|"medium"|"high"
	BuildFailRate   string       `json:"build_fail_rate"`  // "low"|"medium"|"high"
	LastActive      string       `json:"last_active"`      // ISO 8601
	Relationship    Relationship `json:"relationship"`
}

// fileData はファイルに永続化する統計データと DevProfile を合わせた構造体。
type fileData struct {
	DevProfile
	CommitCount   int `json:"commit_count"`
	BuildSuccess  int `json:"build_success"`
	BuildFail     int `json:"build_fail"`
	NightActivity int `json:"night_activity"` // 深夜帯のイベント数
	TotalActivity int `json:"total_activity"` // 全イベント数
}

// ProfileStore は開発者プロファイルを管理する。
type ProfileStore struct {
	mu   sync.Mutex
	path string
	data fileData
}

// NewProfileStore は ProfileStore を初期化する。
// path のファイルが存在する場合は既存データをロードし、累積する。
func NewProfileStore(path string) (*ProfileStore, error) {
	ps := &ProfileStore{path: path}

	raw, err := os.ReadFile(path)
	if err == nil {
		if parseErr := json.Unmarshal(raw, &ps.data); parseErr != nil {
			return nil, fmt.Errorf("parse profile: %w", parseErr)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read profile: %w", err)
	}

	// 統計から DevProfile を再計算（ファイルの DevProfile フィールドは再計算で上書き）
	ps.data.DevProfile = computeProfile(ps.data)
	return ps, nil
}

// RecordCommit はコミットを記録してファイルに書き込む。
func (ps *ProfileStore) RecordCommit() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.data.CommitCount++
	ps.data.DevProfile = computeProfile(ps.data)
	_ = ps.save()
}

// RecordBuildSuccess はビルド成功を記録してファイルに書き込む。
func (ps *ProfileStore) RecordBuildSuccess() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.data.BuildSuccess++
	ps.data.DevProfile = computeProfile(ps.data)
	_ = ps.save()
}

// RecordBuildFail はビルド失敗を記録してファイルに書き込む。
func (ps *ProfileStore) RecordBuildFail() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.data.BuildFail++
	ps.data.DevProfile = computeProfile(ps.data)
	_ = ps.save()
}

// RecordActivity はアクティビティを記録する。高頻度で呼ばれるためファイル I/O は行わない。
func (ps *ProfileStore) RecordActivity(now time.Time) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.data.TotalActivity++
	if isNightHour(now.Hour()) {
		ps.data.NightActivity++
	}
	ps.data.LastActive = now.UTC().Format(time.RFC3339)
	ps.data.DevProfile = computeProfile(ps.data)
}

// Get は現在の DevProfile を返す。
func (ps *ProfileStore) Get() DevProfile {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.data.DevProfile
}

// Stop は LastActive を含む最終データをファイルに書き込む。
func (ps *ProfileStore) Stop() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.save()
}

// save はデータをファイルに書き込む。呼び出し元で mu を保持していること。
func (ps *ProfileStore) save() error {
	b, err := json.MarshalIndent(ps.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ps.path, b, 0644)
}

// computeProfile は統計データから DevProfile を計算する。
func computeProfile(d fileData) DevProfile {
	return DevProfile{
		CommitFrequency: commitFrequency(d.CommitCount),
		BuildFailRate:   buildFailRate(d.BuildSuccess, d.BuildFail),
		NightCoder:      nightCoder(d.NightActivity, d.TotalActivity),
		LastActive:      d.LastActive,
		Relationship:    computeRelationship(d),
	}
}

func computeRelationship(d fileData) Relationship {
	// 基礎レベル: アクティビティ100回ごとに1レベルアップ (上限100)
	level := d.TotalActivity / 100
	if level > 100 {
		level = 100
	}

	// 信頼度: 成功体験の共有 (成功5回ごとに1)
	trust := d.BuildSuccess / 5
	if trust > 100 {
		trust = 100
	}

	// 励まし方針の初期値
	pref := "gentle"
	if d.BuildFail > d.BuildSuccess*2 {
		pref = "strict" // 失敗が多すぎる場合は少し厳しく
	}

	return Relationship{
		Level:                  level,
		Trust:                  trust,
		EncouragementPreference: pref,
	}
}

func commitFrequency(count int) string {
	switch {
	case count >= 5:
		return "high"
	case count >= 2:
		return "medium"
	default:
		return "low"
	}
}

func buildFailRate(success, fail int) string {
	total := success + fail
	if total == 0 {
		return "low"
	}
	rate := float64(fail) / float64(total) * 100
	switch {
	case rate > 60:
		return "high"
	case rate > 30:
		return "medium"
	default:
		return "low"
	}
}

func nightCoder(nightActivity, totalActivity int) bool {
	if totalActivity == 0 {
		return false
	}
	return float64(nightActivity)/float64(totalActivity) >= 0.3
}

// isNightHour は時刻が深夜帯（23:00〜04:59）かを返す。
func isNightHour(hour int) bool {
	return hour >= 23 || hour < 5
}
