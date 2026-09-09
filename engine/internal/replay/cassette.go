package replay

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

// CassetteReplayer manages the lifecycle of the mitmproxy instance used to replay third-party APIs.
type CassetteReplayer struct {
	Runner CommandRunner
	cmd    *exec.Cmd
	done   chan error
	job    *ReplayJob
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
func (c *CassetteReplayer) Start(ctx context.Context, listenPort int, cassetteFile string) error {
	c.cmd = exec.CommandContext(ctx, "mitmdump",
		"--listen-port", fmt.Sprintf("%d", listenPort),
		"--server-replay", cassetteFile,
		"--server-replay-kill-extra", // Drop requests not in cassette
		"--ssl-insecure",             // Allow self-signed/staging certs
	)
	procutil.HideWindow(c.cmd)

	// Route output for debugging if needed, but default to discard to avoid spam
	c.cmd.Stdout = os.Stdout
	c.cmd.Stderr = os.Stderr

	c.done = make(chan error, 1)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("replay: failed to start mitmdump: %w", err)
	}

	job, err := NewReplayJob()
	if err == nil && job != nil {
		c.job = job
		_ = c.job.AssignProcess(c.cmd)
	}

	go func() {
		c.done <- c.cmd.Wait()
		close(c.done)
	}()

	return nil
}

// Stop gracefully shuts down the mitmdump instance.
func (c *CassetteReplayer) Stop() error {
	if c.job != nil {
		defer c.job.Close()
	}
	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	// Send SIGINT for graceful shutdown so it flushes buffers
	if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
		return fmt.Errorf("replay: failed to signal mitmdump: %w", err)
	}

	// Wait for process to exit or kill after timeout
	<-c.done
	return nil
}
