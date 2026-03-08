package monitor

import (
	"context"
	"testing"
	"time"

	"devcompanion/internal/config"
	"devcompanion/internal/types"
)

func TestMonitor_Pipeline(t *testing.T) {
	cfg := config.DefaultAppConfig()
	m, _ := New(cfg, ".")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go m.Run(ctx)

	// シグナルを注入してイベントを確認
	sig := types.Signal{
		Type:      types.SigGitCommit,
		Source:    types.SourceGit,
		Timestamp: time.Now(),
	}
	m.signals <- sig
	m.signals <- sig

	select {
	case ev := <-m.Events():
		if ev.State != types.StateCoding {
			t.Errorf("expected StateCoding, got %v", ev.State)
		}
	case <-time.After(1 * time.Second):
		t.Error("expected event, got none")
	}
}
