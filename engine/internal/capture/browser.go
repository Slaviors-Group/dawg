package capture

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

const defaultBrowserStartupTimeout = 20 * time.Second

// BrowserRecorder manages the canonical Node.js Playwright capture process for PRD §6.1.
type BrowserRecorder struct {
	NodeBinary     string
	ScriptPath     string
	StartupTimeout time.Duration
	CaptureTimeout time.Duration

	mu   sync.Mutex
	cmd  *exec.Cmd
	done chan error
}

// BrowserCaptureRequest identifies a browser capture session and its output files.
type BrowserCaptureRequest struct {
	TargetURL        string
	SessionDirectory string
}

// Start launches the browser recorder and waits until it declares itself ready.
func (recorder *BrowserRecorder) Start(ctx context.Context, request BrowserCaptureRequest) error {
	if request.TargetURL == "" || request.SessionDirectory == "" {
		return fmt.Errorf("capture: target URL and session directory are required")
	}
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.cmd != nil {
		return fmt.Errorf("capture: browser recorder already running")
	}
	if recorder.NodeBinary == "" {
		recorder.NodeBinary = "node"
	}
	if recorder.ScriptPath == "" {
		return fmt.Errorf("capture: browser recorder script path is required")
	}
	if recorder.StartupTimeout <= 0 {
		recorder.StartupTimeout = defaultBrowserStartupTimeout
	}
	if recorder.CaptureTimeout <= 0 {
		recorder.CaptureTimeout = 30 * time.Minute
	}
	for _, directory := range []string{"traces", "http", "actions"} {
		if err := os.MkdirAll(filepath.Join(request.SessionDirectory, directory), 0o700); err != nil {
			return fmt.Errorf("capture: create browser output directory: %w", err)
		}
	}
	captureContext, cancel := context.WithTimeout(ctx, recorder.CaptureTimeout)
	command := exec.CommandContext(captureContext, recorder.NodeBinary, recorder.ScriptPath,
		"--url", request.TargetURL,
		"--rrweb-output", filepath.Join(request.SessionDirectory, "traces", "rrweb.jsonl"),
		"--http-output", filepath.Join(request.SessionDirectory, "http", "frontend.jsonl"),
		"--actions-output", filepath.Join(request.SessionDirectory, "actions", "browser.jsonl"),
	)
	procutil.HideWindow(command)
	stdout, err := command.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("capture: open browser stdout: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("capture: open browser stderr: %w", err)
	}
	if err := command.Start(); err != nil {
		cancel()
		return fmt.Errorf("capture: start Playwright recorder: %w", err)
	}
	recorder.cmd = command
	recorder.done = make(chan error, 1)
	ready := make(chan error, 1)
	go recorder.watch(command, cancel, stdout, stderr, ready)

	startupTimer := time.NewTimer(recorder.StartupTimeout)
	defer startupTimer.Stop()
	select {
	case err := <-ready:
		if err != nil {
			recorder.clearLocked()
			return err
		}
		return nil
	case <-startupTimer.C:
		_ = command.Process.Kill()
		<-recorder.done
		recorder.clearLocked()
		return fmt.Errorf("capture: Playwright recorder startup timed out after %s", recorder.StartupTimeout)
	case <-ctx.Done():
		_ = command.Process.Kill()
		<-recorder.done
		recorder.clearLocked()
		return fmt.Errorf("capture: start Playwright recorder: %w", ctx.Err())
	}
}

// Stop terminates the active browser capture process and waits for it to exit.
// Stop is idempotent: calling it when the recorder is not running is a no-op.
func (recorder *BrowserRecorder) Stop() error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.cmd == nil {
		return nil
	}
	if err := recorder.cmd.Process.Kill(); err != nil {
		return fmt.Errorf("capture: stop Playwright recorder: %w", err)
	}
	<-recorder.done
	recorder.clearLocked()
	return nil
}

func (recorder *BrowserRecorder) watch(command *exec.Cmd, cancel context.CancelFunc, stdout io.Reader, stderr io.Reader, ready chan<- error) {
	defer cancel()
	stderrDone := make(chan []byte, 1)
	go func() {
		contents, _ := io.ReadAll(stderr)
		stderrDone <- contents
	}()
	readyReported := false
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		var status struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			continue
		}
		if status.Status == "capturing" {
			ready <- nil
			readyReported = true
		}
	}
	if err := scanner.Err(); err != nil {
		if !readyReported {
			ready <- fmt.Errorf("capture: read Playwright status: %w", err)
		}
	}
	err := command.Wait()
	stderrContents := <-stderrDone
	if err != nil && len(stderrContents) > 0 {
		err = fmt.Errorf("%w: %s", err, string(stderrContents))
	}
	if !readyReported {
		if err == nil {
			err = fmt.Errorf("capture: Playwright recorder exited before readiness")
		}
		ready <- err
	}
	recorder.done <- err
}

func (recorder *BrowserRecorder) clearLocked() {
	recorder.cmd = nil
	recorder.done = nil
}
