package capture

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

const (
	defaultBrowserStartupTimeout = 20 * time.Second
	defaultBrowserStopTimeout    = 5 * time.Second
)

// BrowserRecorder manages the canonical Node.js Playwright-over-CDP capture process for PRD §6.1.
type BrowserRecorder struct {
	NodeBinary              string
	ScriptPath              string
	CDPEndpoint             string
	ChromiumPath            string
	BrowserProfileDirectory string
	ProxyServer             string
	StartupTimeout          time.Duration
	CaptureTimeout          time.Duration
	StopTimeout             time.Duration

	mu           sync.Mutex
	cmd          *exec.Cmd
	stdin        io.WriteCloser
	done         chan struct{}
	exitErr      error
	stopping     bool
	stopRequests chan struct{}
}

// BrowserCaptureRequest identifies a browser capture session and its output files.
type BrowserCaptureRequest struct {
	TargetURL        string
	SessionDirectory string
}

// Start connects the browser recorder to an existing Chromium browser and waits for readiness.
func (recorder *BrowserRecorder) Start(ctx context.Context, request BrowserCaptureRequest) error {
	if request.TargetURL == "" || request.SessionDirectory == "" {
		return fmt.Errorf("capture: target URL and session directory are required")
	}
	recorder.mu.Lock()
	if recorder.cmd != nil {
		recorder.mu.Unlock()
		return fmt.Errorf("capture: browser recorder already running")
	}
	if recorder.NodeBinary == "" {
		recorder.NodeBinary = "node"
	}
	if recorder.ScriptPath == "" {
		recorder.mu.Unlock()
		return fmt.Errorf("capture: browser recorder script path is required")
	}
	if recorder.CDPEndpoint == "" {
		recorder.CDPEndpoint = "http://127.0.0.1:9222"
	}
	if recorder.StartupTimeout <= 0 {
		recorder.StartupTimeout = defaultBrowserStartupTimeout
	}
	if recorder.CaptureTimeout <= 0 {
		recorder.CaptureTimeout = 30 * time.Minute
	}
	if recorder.StopTimeout <= 0 {
		recorder.StopTimeout = defaultBrowserStopTimeout
	}
	recorder.mu.Unlock()

	for _, directory := range []string{"traces", "http", "actions"} {
		if err := os.MkdirAll(filepath.Join(request.SessionDirectory, directory), 0o700); err != nil {
			return fmt.Errorf("capture: create browser output directory: %w", err)
		}
	}

	captureContext, cancel := context.WithTimeout(ctx, recorder.CaptureTimeout)
	arguments := []string{recorder.ScriptPath,
		"--url", request.TargetURL,
		"--cdp-endpoint", recorder.CDPEndpoint,
		"--rrweb-output", filepath.Join(request.SessionDirectory, "traces", "rrweb.jsonl"),
		"--http-output", filepath.Join(request.SessionDirectory, "http", "frontend.jsonl"),
		"--actions-output", filepath.Join(request.SessionDirectory, "actions", "browser.jsonl"),
	}
	if recorder.ChromiumPath != "" {
		arguments = append(arguments, "--chromium-path", recorder.ChromiumPath)
	}
	if recorder.BrowserProfileDirectory != "" {
		arguments = append(arguments, "--browser-profile", recorder.BrowserProfileDirectory)
	}
	if recorder.ProxyServer != "" {
		arguments = append(arguments, "--proxy-server", recorder.ProxyServer)
	}
	command := exec.CommandContext(captureContext, recorder.NodeBinary, arguments...)
	procutil.HideWindow(command)
	stdin, err := command.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("capture: open browser stdin: %w", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		cancel()
		_ = stdin.Close()
		return fmt.Errorf("capture: open browser stdout: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		cancel()
		_ = stdin.Close()
		return fmt.Errorf("capture: open browser stderr: %w", err)
	}
	if err := command.Start(); err != nil {
		cancel()
		_ = stdin.Close()
		return fmt.Errorf("capture: start Playwright recorder: %w", err)
	}

	done := make(chan struct{})
	stopRequests := make(chan struct{}, 1)
	ready := make(chan error, 1)
	recorder.mu.Lock()
	recorder.cmd = command
	recorder.stdin = stdin
	recorder.done = done
	recorder.exitErr = nil
	recorder.stopping = false
	recorder.stopRequests = stopRequests
	recorder.mu.Unlock()
	go recorder.watch(command, cancel, stdout, stderr, ready, done, stopRequests)

	startupTimer := time.NewTimer(recorder.StartupTimeout)
	defer startupTimer.Stop()
	select {
	case err := <-ready:
		if err != nil {
			recorder.clear(command)
			return err
		}
		return nil
	case <-startupTimer.C:
		_ = recorder.Stop()
		return fmt.Errorf("capture: Playwright recorder startup timed out after %s", recorder.StartupTimeout)
	case <-ctx.Done():
		_ = recorder.Stop()
		return fmt.Errorf("capture: start Playwright recorder: %w", ctx.Err())
	}
}

// Stop asks the Node recorder to drain pending events and disconnect from CDP, then enforces a timeout.
// Stop is idempotent and safe for concurrent callers.
func (recorder *BrowserRecorder) Stop() error {
	recorder.mu.Lock()
	command := recorder.cmd
	if command == nil {
		recorder.mu.Unlock()
		return nil
	}
	stdin := recorder.stdin
	done := recorder.done
	stopTimeout := recorder.StopTimeout
	requestStop := !recorder.stopping
	recorder.stopping = true
	recorder.mu.Unlock()

	var requestErr error
	if requestStop {
		if _, err := io.WriteString(stdin, "{\"command\":\"stop\"}\n"); err != nil {
			requestErr = fmt.Errorf("capture: request Playwright recorder stop: %w", err)
			_ = command.Process.Kill()
		}
	}

	timer := time.NewTimer(stopTimeout)
	defer timer.Stop()
	select {
	case <-done:
		recorder.mu.Lock()
		exitErr := recorder.exitErr
		recorder.mu.Unlock()
		recorder.clear(command)
		if requestErr != nil && exitErr != nil {
			return errors.Join(requestErr, exitErr)
		}
		if requestErr != nil {
			return requestErr
		}
		if exitErr != nil {
			return fmt.Errorf("capture: Playwright recorder shutdown: %w", exitErr)
		}
		return nil
	case <-timer.C:
		_ = command.Process.Kill()
		select {
		case <-done:
			recorder.clear(command)
			return fmt.Errorf("capture: Playwright recorder did not stop within %s", stopTimeout)
		case <-time.After(2 * time.Second):
			return fmt.Errorf("capture: Playwright recorder did not exit after forced termination")
		}
	}
}

// StopRequested reports an unexpected browser-recorder exit to the owning capture daemon.
func (recorder *BrowserRecorder) StopRequested() <-chan struct{} {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return recorder.stopRequests
}

func (recorder *BrowserRecorder) watch(command *exec.Cmd, cancel context.CancelFunc, stdout io.Reader, stderr io.Reader, ready chan<- error, done chan struct{}, stopRequests chan<- struct{}) {
	defer cancel()
	stderrDone := make(chan []byte, 1)
	go func() {
		contents, _ := io.ReadAll(stderr)
		stderrDone <- contents
	}()

	readyReported := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		var status struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &status); err != nil {
			continue
		}
		if status.Status == "capturing" && !readyReported {
			ready <- nil
			readyReported = true
		}
	}

	waitErr := command.Wait()
	stderrContents := <-stderrDone
	if scanErr := scanner.Err(); scanErr != nil {
		waitErr = errors.Join(waitErr, fmt.Errorf("read Playwright status: %w", scanErr))
	}
	if waitErr != nil && len(stderrContents) > 0 {
		waitErr = fmt.Errorf("%w: %s", waitErr, string(stderrContents))
	}

	recorder.mu.Lock()
	requested := recorder.stopping
	if !requested && readyReported && waitErr == nil {
		waitErr = fmt.Errorf("Playwright recorder exited unexpectedly")
	}
	recorder.exitErr = waitErr
	recorder.mu.Unlock()

	if !readyReported {
		if waitErr == nil {
			waitErr = fmt.Errorf("capture: Playwright recorder exited before readiness")
		}
		ready <- waitErr
	}
	if !requested && readyReported {
		select {
		case stopRequests <- struct{}{}:
		default:
		}
	}
	close(done)
}

func (recorder *BrowserRecorder) clear(command *exec.Cmd) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.cmd != command {
		return
	}
	if recorder.stdin != nil {
		_ = recorder.stdin.Close()
	}
	recorder.cmd = nil
	recorder.stdin = nil
	recorder.done = nil
	recorder.exitErr = nil
	recorder.stopping = false
	recorder.stopRequests = nil
}
