package pipeline

import (
	"log"
	"runtime/debug"
)

// SafeExecute wraps a function with panic recovery to ensure pipeline continuity.
func SafeExecute(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC-GUARD] recovered in %s: %v\nStack trace:\n%s", name, r, debug.Stack())
		}
	}()
	fn()
}
