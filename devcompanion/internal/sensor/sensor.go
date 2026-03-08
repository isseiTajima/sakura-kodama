package sensor

import (
	"context"
	"devcompanion/internal/types"
)

// Sensor はOSやアプリの行動を観測し、Signalsを生成する。
type Sensor interface {
	Run(ctx context.Context, signals chan<- types.Signal) error
	Name() string
}
