package replay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

// EventPlayer replays a recorded browser session through Playwright.
type EventPlayer struct {
	NodeBinary      string
	ScriptPath      string
	BrowsersDir     string
	ChromiumPath    string
	ColorScheme     string
	Interactive     bool
	ReplayTimeout   time.Duration
	ReplayEventSink func(json.RawMessage) error
}

// ReplayOutcome captures the results of a browser replay run.
type ReplayOutcome struct {
	ScreenshotPath string
	// Output holds the combined stdout+stderr of the browser replay script,
	// regardless of whether it succeeded. replay-browser.cjs writes
	// diagnostic detail here (event counts, in-page console/pageerror
	// output) that is otherwise invisible when the process exits 0 but the
	// replayed page rendered incorrectly (e.g. blank).
	Output string
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

	replayContext := ctx
	cancel := func() {}
	if !player.Interactive {
		replayContext, cancel = context.WithTimeout(ctx, timeout)
	}
	defer cancel()

	arguments := []string{
		scriptPath,
		"--rrweb-input", traceInput,
		"--screenshot-output", screenshotOutput,
	}
	if player.ColorScheme == "dark" || player.ColorScheme == "light" {
		arguments = append(arguments, "--color-scheme", player.ColorScheme)
	}
	if player.Interactive {
		arguments = append(arguments, "--interactive")
	}
	cmd := exec.CommandContext(replayContext, nodeBinary, arguments...)
	// Interactive stdin carries only versioned replay-control messages.
	if player.Interactive {
		cmd.Stdin = os.Stdin
	}
	procutil.HideWindow(cmd)
	if player.BrowsersDir != "" {
		cmd.Env = append(os.Environ(), "PLAYWRIGHT_BROWSERS_PATH="+player.BrowsersDir)
	}
	if player.ChromiumPath != "" {
		if cmd.Env == nil {
			cmd.Env = os.Environ()
		}
		cmd.Env = append(cmd.Env, "DAWG_CHROMIUM_EXECUTABLE_PATH="+player.ChromiumPath)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ReplayOutcome{}, fmt.Errorf("replay: open browser stdout: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return ReplayOutcome{}, fmt.Errorf("replay: open browser stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return ReplayOutcome{}, fmt.Errorf("replay: failed to start browser replay: %w", err)
	}

	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	readerErrors := make(chan error, 2)
	var readers sync.WaitGroup
	readers.Add(2)
	go func() {
		defer readers.Done()
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 64*1024), 16<<20)
		for scanner.Scan() {
			line := append([]byte(nil), scanner.Bytes()...)
			if player.Interactive && player.ReplayEventSink != nil {
				if !json.Valid(line) {
					readerErrors <- fmt.Errorf("replay: invalid browser event: %q", line)
					return
				}
				if err := player.ReplayEventSink(json.RawMessage(line)); err != nil {
					readerErrors <- fmt.Errorf("replay: forward browser event: %w", err)
					return
				}
				continue
			}
			stdoutBuf.Write(line)
			stdoutBuf.WriteByte('\n')
		}
		readerErrors <- scanner.Err()
	}()
	go func() {
		defer readers.Done()
		scanner := bufio.NewScanner(stderr)
		scanner.Buffer(make([]byte, 64*1024), 16<<20)
		for scanner.Scan() {
			stderrBuf.Write(scanner.Bytes())
			stderrBuf.WriteByte('\n')
		}
		readerErrors <- scanner.Err()
	}()

	job, jobErr := NewReplayJob()
	if jobErr == nil && job != nil {
		_ = job.AssignProcess(cmd)
		defer job.Close()
	}

	err = cmd.Wait()
	readers.Wait()
	var readerErr error
	for range 2 {
		if current := <-readerErrors; current != nil && readerErr == nil {
			readerErr = current
		}
	}
	output := append(stderrBuf.Bytes(), stdoutBuf.Bytes()...)
	if readerErr != nil && err == nil {
		err = readerErr
	}

	if err != nil {
		if replayContext.Err() == context.DeadlineExceeded {
			return ReplayOutcome{Output: string(output)}, fmt.Errorf("replay: browser replay timed out after %s: %s", timeout, string(output))
		}
		return ReplayOutcome{Output: string(output)}, fmt.Errorf("replay: browser replay failed: %w\n%s", err, string(output))
	}

	screenshotPath := screenshotOutput
	if player.Interactive {
		screenshotPath = ""
	}
	return ReplayOutcome{
		ScreenshotPath: screenshotPath,
		Output:         string(output),
	}, nil
}
