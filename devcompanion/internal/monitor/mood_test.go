package monitor

import "testing"

func TestInferMood_SuccessReturnsHappy(t *testing.T) {
	mood := InferMood(StateSuccess, TaskRunTests)
	if mood != MoodHappy {
		t.Errorf("want %s for success state, got %s", MoodHappy, mood)
	}
}

func TestInferMood_FailReturnsNervous(t *testing.T) {
	mood := InferMood(StateFail, TaskDebug)
	if mood != MoodNervous {
		t.Errorf("want %s for fail state, got %s", MoodNervous, mood)
	}
}

func TestInferMood_RunningGenerateCodeIsFocus(t *testing.T) {
	mood := InferMood(StateRunning, TaskGenerateCode)
	if mood != MoodFocus {
		t.Errorf("want %s when running & generating code, got %s", MoodFocus, mood)
	}
}

func TestInferMood_RunningFixFailingTestsIsFocus(t *testing.T) {
	mood := InferMood(StateRunning, TaskFixFailingTests)
	if mood != MoodFocus {
		t.Errorf("want %s for fix-failing-tests focus, got %s", MoodFocus, mood)
	}
}

func TestInferMood_ThinkingPlanDefaultsCalm(t *testing.T) {
	mood := InferMood(StateThinking, TaskPlan)
	if mood != MoodCalm {
		t.Errorf("want %s for thinking plan, got %s", MoodCalm, mood)
	}
}
