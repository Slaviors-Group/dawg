package replay

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

// EventPlayer orchestrates the playback of a recorded browser session using Playwright.
// Implements PRD §6.5 (Replay Engine) browser replay layer.
type EventPlayer struct {
	NodeBinary     string
	ScriptPath     string
	ReplayTimeout  time.Duration
}

// ReplayOutcome captures the results of a browser replay run.
type ReplayOutcome struct {
	ScreenshotPath string
}

// Replay executes the browser replay script synchronously, passing the required traces.
func (player *EventPlayer) Replay(ctx context.Context, sessionDirectory string) (ReplayOutcome, error) {
	if sessionDirectory == "" {
		return ReplayOutcome{}, fmt.Errorf("replay: session directory is required")
	}

	nodeBinary := player.NodeBinary
	if nodeBinary == "" {
		nodeBinary = "node"
	}
	scriptPath := player.ScriptPath
	if scriptPath == "" {
		return ReplayOutcome{}, fmt.Errorf("replay: event player script path is required")
	}
	timeout := player.ReplayTimeout
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	traceInput := filepath.Join(sessionDirectory, "traces", "rrweb.jsonl")
	screenshotOutput := filepath.Join(sessionDirectory, "outcome", "screenshot.png")

	replayContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(replayContext, nodeBinary, scriptPath,
		"--rrweb-input", traceInput,
		"--screenshot-output", screenshotOutput,
	)
	procutil.HideWindow(cmd)

	var outputBuf bytes.Buffer
	cmd.Stdout = &outputBuf
	cmd.Stderr = &outputBuf

	if err := cmd.Start(); err != nil {
		return ReplayOutcome{}, fmt.Errorf("replay: failed to start browser replay: %w", err)
	}

	job, jobErr := NewReplayJob()
	if jobErr == nil && job != nil {
		_ = job.AssignProcess(cmd)
		defer job.Close()
	}

	err := cmd.Wait()
	output := outputBuf.Bytes()

	if err != nil {
		if replayContext.Err() == context.DeadlineExceeded {
			return ReplayOutcome{}, fmt.Errorf("replay: browser replay timed out after %s: %s", timeout, string(output))
		}
		return ReplayOutcome{}, fmt.Errorf("replay: browser replay failed: %w\n%s", err, string(output))
	}

	return ReplayOutcome{
		ScreenshotPath: screenshotOutput,
	}, nil
}
