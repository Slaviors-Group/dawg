package replay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

// CassetteReplayer manages the lifecycle of the mitmproxy instance used to replay third-party APIs.
type CassetteReplayer struct {
	Runner      CommandRunner
	Executable  string
	StopTimeout time.Duration
	cmd         *exec.Cmd
	// done receives the result of cmd.Wait() exactly once, then is set to nil
	// in Stop() after draining. The goroutine never closes the channel, so a
	// second Stop() call (double-click) sees done == nil and returns safely
	// without blocking or panicking on a second receive from a drained channel.
	done chan error
	job  *ReplayJob
}

// NewCassetteReplayer creates a new CassetteReplayer.
func NewCassetteReplayer(runner CommandRunner) *CassetteReplayer {
	if runner == nil {
		runner = ExecRunner{}
	}
	return &CassetteReplayer{
		Runner: runner,
	}
}

// Start launches mitmdump in server-replay mode.
// Start returns an error if called while a previous replayer instance is still running.
func (c *CassetteReplayer) Start(ctx context.Context, listenPort int, cassetteFile string) error {
	// Idempotency guard: a second Start() while mitmdump is still up would
	// leave the old process holding the listen port, causing the new one to
	// fail with "address already in use" — surfaced as "start replay error".
	if c.cmd != nil {
		return fmt.Errorf("replay: cassette replayer already running; call Stop() first")
	}

	executable := c.Executable
	if executable == "" {
		executable = "mitmdump"
	}
	c.cmd = exec.CommandContext(ctx, executable,
		"--listen-port", fmt.Sprintf("%d", listenPort),
		"--server-replay", cassetteFile,
		"--server-replay-kill-extra", // Drop requests not in cassette
		"--ssl-insecure",             // Allow self-signed/staging certs
	)
	procutil.HideWindow(c.cmd)

	// Route output for debugging if needed, but default to discard to avoid spam
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr

	// Use a 1-buffered channel. The goroutine sends exactly once and never
	// closes the channel — this lets Stop() drain it safely, then nil it out,
	// so a second Stop() call (done == nil) returns immediately without
	// blocking or panicking.
	c.done = make(chan error, 1)

	if err := c.cmd.Start(); err != nil {
		c.cmd = nil
		c.done = nil
		return fmt.Errorf("replay: failed to start mitmdump: %w", err)
	}

	job, err := NewReplayJob()
	if err == nil && job != nil {
		c.job = job
		_ = c.job.AssignProcess(c.cmd)
	}

	done := c.done // capture for goroutine; avoids race with Stop() nil-ing c.done
	cmd := c.cmd
	go func() {
		done <- cmd.Wait()
		// Intentionally NOT close(done). Stop() drains and nils the channel,
		// preventing a second receive from blocking forever on a closed channel.
	}()

	return nil
}

// Stop gracefully shuts down the mitmdump instance.
// Stop is idempotent: calling it when no replayer is running is a no-op.
func (c *CassetteReplayer) Stop() error {
	if c.job != nil {
		defer c.job.Close()
		c.job = nil
	}
	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	// Send SIGINT for graceful shutdown so it flushes buffers. Windows does not
	// implement os.Interrupt for arbitrary child processes, so fall back to a
	// direct kill there instead of leaking the replay process indefinitely.
	if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
		if killErr := c.cmd.Process.Kill(); killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
			return fmt.Errorf("replay: stop mitmdump after interrupt failed: %w", errors.Join(err, killErr))
		}
	}

	stopTimeout := c.StopTimeout
	if stopTimeout <= 0 {
		stopTimeout = 5 * time.Second
	}
	timer := time.NewTimer(stopTimeout)
	select {
	case <-c.done:
	case <-timer.C:
		if err := c.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("replay: force stop mitmdump: %w", err)
		}
		<-c.done
	}
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	c.cmd = nil
	c.done = nil
	return nil
}
