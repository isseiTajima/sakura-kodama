package behavior

import (
	"testing"
	"time"

	"devcompanion/internal/types"
)

func TestInferrer_Infer(t *testing.T) {
	inf := NewInferrer(5 * time.Minute)
	now := time.Now()

	// 1. AI Pairing
	inf.AddSignal(types.Signal{Source: types.SourceAgent, Timestamp: now})
	inf.AddSignal(types.Signal{Source: types.SourceFS, Timestamp: now.Add(1 * time.Second)})
	
	b := inf.Infer()
	if b.Type != types.BehaviorAIPairing {
		t.Errorf("expected BehaviorAIPairing, got %v", b.Type)
	}

	// 2. Debugging
	inf = NewInferrer(5 * time.Minute)
	inf.AddSignal(types.Signal{Source: types.SourceFS, Message: "FAIL", Timestamp: now})
	
	b = inf.Infer()
	if b.Type != types.BehaviorDebugging {
		t.Errorf("expected BehaviorDebugging, got %v", b.Type)
	}

	// 3. Coding
	inf = NewInferrer(5 * time.Minute)
	inf.AddSignal(types.Signal{Source: types.SourceFS, Timestamp: now})
	
	b = inf.Infer()
	if b.Type != types.BehaviorCoding {
		t.Errorf("expected BehaviorCoding, got %v", b.Type)
	}
}
