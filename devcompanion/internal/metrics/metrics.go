package metrics

import (
	"sync/atomic"
)

var (
	SignalsReceivedTotal  uint64
	ContextSwitchTotal    uint64
	PersonaMessagesTotal uint64
)

func IncrementSignalsReceived() {
	atomic.AddUint64(&SignalsReceivedTotal, 1)
}

func IncrementContextSwitch() {
	atomic.AddUint64(&ContextSwitchTotal, 1)
}

func IncrementPersonaMessages() {
	atomic.AddUint64(&PersonaMessagesTotal, 1)
}

// GetMetrics returns current values for future Prometheus integration
func GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"signals_received_total": atomic.LoadUint64(&SignalsReceivedTotal),
		"context_switch_total":   atomic.LoadUint64(&ContextSwitchTotal),
		"persona_messages_total": atomic.LoadUint64(&PersonaMessagesTotal),
	}
}
