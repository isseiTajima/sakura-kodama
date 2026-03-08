package llm

import "time"

// MoodState は感情状態を表す。
type MoodState string

const (
	MoodStateHappy   MoodState = "happy"
	MoodStateCalm    MoodState = "calm"
	MoodStateSleepy  MoodState = "sleepy"
	MoodStateExcited MoodState = "excited"
	MoodStateFail    MoodState = "fail" // 追加: ビルド失敗などのネガティブな状態
)

// InferMoodState はイベント時刻と連続成功カウントから感情状態を推論する。
func InferMoodState(lastEventTime time.Time, successStreak int, lastReason Reason) MoodState {
	// 1. ビルド失敗時はネガティブな状態を優先
	if lastReason == ReasonFail {
		return MoodStateFail
	}

	// 2. 成功ストリーク優先判定
	if successStreak >= 3 {
		return MoodStateExcited
	}

	// 3. Sleepy 判定（LOCAL TIME で 01:00-05:00）
	loc, err := time.LoadLocation("Local")
	if err == nil {
		localTime := lastEventTime.In(loc)
		hour := localTime.Hour()
		if hour >= 1 && hour < 5 {
			return MoodStateSleepy
		}
	}

	// デフォルト
	return MoodStateHappy
}
