package ws

import "time"

// Event はUIへブロードキャストするイベント型。
type Event struct {
	State         string       `json:"state"`
	Task          string       `json:"task"`
	Mood          string       `json:"mood"`
	Speech        string       `json:"speech"`
	Timestamp     time.Time    `json:"timestamp"`
	UsingFallback bool         `json:"using_fallback"`
	Profile       EventProfile `json:"profile"`
}

// EventProfile はキャラクターのプロフィール情報をまとめる。
type EventProfile struct {
	Name string `json:"name"`
	Tone string `json:"tone"`
}
