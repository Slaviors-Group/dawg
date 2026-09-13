package replay

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// TestCassetteReplayerIdempotentStop verifies that calling Stop() on a
// CassetteReplayer that was never started returns nil instead of panicking or
// blocking forever — the original double-close bug.
func TestCassetteReplayerIdempotentStop(t *testing.T) {
	c := NewCassetteReplayer(nil)
	if err := c.Stop(); err != nil {
		t.Fatalf("Stop on never-started replayer: expected nil, got %v", err)
	}
	// Calling Stop() a second time must also be safe (no channel panic, no block).
	if err := c.Stop(); err != nil {
		t.Fatalf("second Stop (idempotent): expected nil, got %v", err)
	}
}

// TestCassetteReplayerRejectsDoubleStart verifies that calling Start() while a
// previous instance is still running returns a clear error instead of silently
// leaking the old process and holding the listen port — the root cause of the
// "start replay error" on second replay.
func TestCassetteReplayerRejectsDoubleStart(t *testing.T) {
	c := NewCassetteReplayer(nil)

	// Simulate an already-started replayer by populating c.cmd directly.
	// The guard in Start() fires before exec.Command is constructed, so the
	// fake cmd value never needs a real PID.
	c.cmd = &exec.Cmd{}

	err := c.Start(context.Background(), 8080, "cassette.yaml")
	if err == nil {
		t.Fatal("expected double-Start error, got nil")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Fatalf("unexpected error message: %v", err)
	}

	// Clean up the fake cmd so the replayer doesn't try to signal it.
	c.cmd = nil
}
