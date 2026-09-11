package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/capture"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgenv"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/packager"
	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
	"github.com/Slaviors-Group/dawg/engine/internal/sanitize"
	"github.com/spf13/cobra"
)

var defaultCaptureControlFile = defaultDawgDir("capture-control.json")
var defaultCaptureResultFile = defaultDawgDir("capture-result.json")

type captureStopResult struct {
	Status       string `json:"status"`
	ControlFile  string `json:"controlFile"`
	SessionID    string `json:"sessionId,omitempty"`
	ArtifactPath string `json:"artifactPath,omitempty"`
}

// captureResultState is written by the daemon process once the full
// capture->sanitize->package pipeline finishes (successfully or not), so
// that a separate `dawg capture stop` invocation - which has no handle to
// the daemon's os/exec.Cmd - can learn the real, absolute artifact path
// instead of guessing one.
type captureResultState struct {
	Status       string `json:"status"`
	SessionID    string `json:"sessionId"`
	ArtifactPath string `json:"artifactPath,omitempty"`
	Error        string `json:"error,omitempty"`
}

type captureStartResult struct {
	Status      string `json:"status"`
	SessionID   string `json:"sessionId"`
	SessionPath string `json:"sessionPath"`
	ControlFile string `json:"controlFile"`
	ResultFile  string `json:"resultFile"`
	DaemonPID   int    `json:"daemonPid"`
}

func newCaptureCommand() *cobra.Command {
	var targetURL string
	var sessionDirectory string
	var composeFile string
	var dbDiffFile string
	var logFile string
	var policyFile string
	var unsafeSkipSanitize bool
	var daemon bool
	var resultFile string

	command := &cobra.Command{
		Use:   "capture",
		Short: "Capture a web-app reproduction session",
		RunE: func(command *cobra.Command, _ []string) error {
			parsedURL, err := validateCaptureTarget(targetURL)
			if err != nil {
				return err
			}
			request := captureStartRequest{
				TargetURL:          targetURL,
				SessionDirectory:   sessionDirectory,
				ComposeFile:        composeFile,
				DBDiffFile:         dbDiffFile,
				LogFile:            logFile,
				PolicyFile:         policyFile,
				UnsafeSkipSanitize: unsafeSkipSanitize,
				ResultFile:         resultFile,
			}
			if unsafeSkipSanitize {
				hostname := parsedURL.Hostname()
				if hostname != "localhost" && hostname != "127.0.0.1" && hostname != "::1" {
					return fmt.Errorf("capture: --unsafe-skip-sanitize is only allowed for localhost targets")
				}
			}
			if daemon {
				return runCaptureDaemon(command.Context(), request)
			}
			result, err := launchCaptureDaemon(command.Context(), request)
			if err != nil {
				return err
			}
			return writeCaptureStartResult(command.OutOrStdout(), outputFormat(command), result)
		},
	}
	command.Flags().StringVar(&targetURL, "url", "", "Target URL to capture")
	command.Flags().StringVar(&sessionDirectory, "session-dir", defaultDawgDir("captures"), "Directory containing capture sessions")
	command.Flags().StringVar(&composeFile, "compose-file", "", "Path to docker-compose file for environment snapshot")
	command.Flags().StringVar(&dbDiffFile, "db-diff-file", "", "Path to DB diff stream or file")
	command.Flags().StringVar(&logFile, "log-file", "", "Path to structured application log file")
	command.Flags().StringVar(&policyFile, "policy-file", defaultEnginePolicy(), "Path to OPA sanitization policy")
	command.Flags().BoolVar(&unsafeSkipSanitize, "unsafe-skip-sanitize", false, "Skip sanitization (only allowed for localhost targets)")
	command.Flags().BoolVar(&daemon, "daemon", false, "Run the capture owner process")
	command.Flags().StringVar(&resultFile, "result-file", "", "Capture daemon result-state file")
	_ = command.Flags().MarkHidden("result-file")
	_ = command.Flags().MarkHidden("daemon")
	_ = command.MarkFlagRequired("url")
	command.AddCommand(newCaptureStopCommand())
	return command
}

type captureStartRequest struct {
	TargetURL          string
	SessionDirectory   string
	ComposeFile        string
	DBDiffFile         string
	LogFile            string
	PolicyFile         string
	UnsafeSkipSanitize bool
	ResultFile         string
}

func launchCaptureDaemon(ctx context.Context, request captureStartRequest) (captureStartResult, error) {
	if request.TargetURL == "" {
		return captureStartResult{}, fmt.Errorf("capture: --url is required")
	}
	if state, err := capture.ReadControlState(defaultCaptureControlFile); err == nil {
		healthCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
		active := capture.ControlledSessionActive(healthCtx, state)
		cancel()
		if active {
			return captureStartResult{}, dawgtypes.ErrCaptureAlreadyRunning
		}
		if err := os.Remove(defaultCaptureControlFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			return captureStartResult{}, fmt.Errorf("capture: remove stale control state: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return captureStartResult{}, fmt.Errorf("capture: validate existing control state: %w", err)
	}

	sessionID, err := newSessionID()
	if err != nil {
		return captureStartResult{}, err
	}
	sessionPath := filepath.Join(request.SessionDirectory, sessionID)
	absoluteSessionPath, err := filepath.Abs(sessionPath)
	if err != nil {
		return captureStartResult{}, fmt.Errorf("capture: resolve session directory: %w", err)
	}
	if err := os.MkdirAll(absoluteSessionPath, 0o700); err != nil {
		return captureStartResult{}, fmt.Errorf("capture: create session directory: %w", err)
	}
	resultPath := defaultDawgDir("capture-results", sessionID+".json")
	absoluteResultPath, err := filepath.Abs(resultPath)
	if err != nil {
		return captureStartResult{}, fmt.Errorf("capture: resolve result-state file: %w", err)
	}
	request.ResultFile = absoluteResultPath
	executable, err := os.Executable()
	if err != nil {
		return captureStartResult{}, fmt.Errorf("capture: locate executable: %w", err)
	}
	_ = os.Remove(absoluteResultPath)
	arguments := daemonArguments(request, absoluteSessionPath)
	process := exec.Command(executable, arguments...)
	procutil.HideWindow(process)
	procutil.Detach(process)

	logFile, err := os.Create(filepath.Join(absoluteSessionPath, "daemon.log"))
	if err == nil {
		process.Stdout = logFile
		process.Stderr = logFile
	}

	if err := process.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return captureStartResult{}, fmt.Errorf("capture: start daemon: %w", err)
	}
	if logFile != nil {
		_ = logFile.Close()
	}
	controlPath, err := filepath.Abs(defaultCaptureControlFile)
	if err != nil {
		_ = process.Process.Kill()
		_ = process.Wait()
		return captureStartResult{}, fmt.Errorf("capture: resolve control-state file: %w", err)
	}
	state, err := waitForControlState(ctx, controlPath, absoluteResultPath, sessionID, 20*time.Second)
	if err != nil {
		_ = process.Process.Kill()
		_ = process.Wait()
		return captureStartResult{}, err
	}
	if err := process.Process.Release(); err != nil {
		_ = process.Process.Kill()
		_ = process.Wait()
		return captureStartResult{}, fmt.Errorf("capture: release daemon process: %w", err)
	}
	return captureStartResult{
		Status:      "capturing",
		SessionID:   sessionID,
		SessionPath: absoluteSessionPath,
		ControlFile: controlPath,
		ResultFile:  state.ResultFile,
		DaemonPID:   state.PID,
	}, nil
}

func runCaptureDaemon(ctx context.Context, request captureStartRequest) error {
	if _, err := validateCaptureTarget(request.TargetURL); err != nil {
		return err
	}
	sessionID := filepath.Base(request.SessionDirectory)
	resultFile := request.ResultFile
	if resultFile == "" {
		resultFile = defaultCaptureResultFile
	}

	var artifactPath string
	var pipelineErr error
	defer func() {
		state := captureResultState{SessionID: sessionID}
		if pipelineErr != nil {
			state.Status = "error"
			state.Error = pipelineErr.Error()
		} else {
			state.Status = "packaged"
			state.ArtifactPath = artifactPath
		}
		_ = writeCaptureResultState(resultFile, state)
	}()

	components := []capture.SessionComponent{}
	if request.ComposeFile != "" {
		components = append(components, &envComponent{runner: capture.ExecRunner{}, composeFile: request.ComposeFile})
	}
	if request.DBDiffFile != "" {
		components = append(components, &dbComponent{sourcePath: request.DBDiffFile})
	}
	if request.LogFile != "" {
		components = append(components, &logComponent{sourcePath: request.LogFile})
	}
	// The extension is last so Session.Stop asks it to flush and acknowledge
	// before any optional environment streams are finalized and packaged.
	components = append(components, &capture.ExtensionServer{
		ListenAddr:       "127.0.0.1:8082",
		TargetURL:        request.TargetURL,
		RequireHandshake: true,
		StartupTimeout:   13 * time.Second,
	})

	session, err := capture.NewSession(capture.SessionOptions{
		Directory:  request.SessionDirectory,
		ResultFile: resultFile,
		Metadata: dawgtypes.CaptureMetadata{
			SessionID:   sessionID,
			TargetURL:   request.TargetURL,
			ActionTrace: &dawgtypes.ActionTrace{Path: "actions/browser.jsonl", Version: version},
		},
		Components: components,
	})
	if err != nil {
		pipelineErr = err
		return err
	}

	if err := capture.RunControlledSession(ctx, session, defaultCaptureControlFile); err != nil {
		pipelineErr = err
		return err
	}

	// Pipeline stage 2: Sanitize
	if !request.UnsafeSkipSanitize {
		_, err = sanitize.SanitizeDirectory(ctx, request.SessionDirectory, request.PolicyFile, "1.0.0")
		if err != nil {
			pipelineErr = err
			return err // ErrExportBlocked is naturally propagated here
		}
	} else {
		// Write dummy report to satisfy the packager
		dummyReport := dawgtypes.SanitizeReport{
			PolicyVersion:  "skipped",
			PolicyFile:     "skipped",
			OPAResult:      "allow",
			ExportAllowed:  true,
			FieldsScanned:  0,
			FieldsRedacted: 0,
			Redactions:     []dawgtypes.Redaction{},
			BlockedFields:  []string{},
		}
		reportPath := filepath.Join(request.SessionDirectory, "sanitize-report.json")
		reportBytes, _ := json.MarshalIndent(dummyReport, "", "  ")
		_ = os.WriteFile(reportPath, append(reportBytes, '\n'), 0o600)
	}

	// Pipeline stage 3: Package
	packageRequest := dawgtypes.PackageRequest{
		SessionDirectory: request.SessionDirectory,
		OutputDirectory:  defaultDawgDir("artifacts", sessionID),
		Title:            "Captured Session " + sessionID,
		Source: dawgtypes.ManifestSource{
			Reporter:    "local",
			Environment: "dev",
			RepoCommit:  "unknown",
		},
		ExpectedOutcome: dawgtypes.ExpectedOutcome{
			Type:          "assertion",
			Description:   "Manual reproduction",
			AssertionFile: "none",
		},
	}
	artifact, err := packager.Package(packageRequest)
	if err != nil {
		pipelineErr = err
		return err
	}
	artifactPath = artifact.Directory
	return nil
}

type envComponent struct {
	runner      capture.CommandRunner
	composeFile string
}

func (c *envComponent) Name() string { return "environment" }
func (c *envComponent) Start(ctx context.Context, directory string) error {
	_, err := capture.SnapshotEnvironment(ctx, c.runner, capture.EnvironmentSnapshotRequest{
		ComposeFile:      c.composeFile,
		SessionDirectory: directory,
	})
	return err
}
func (c *envComponent) Stop() error { return nil }

type dbComponent struct {
	sourcePath string
	file       *os.File
	cancel     context.CancelFunc
	done       chan struct{}
}

func (c *dbComponent) Name() string { return "db" }
func (c *dbComponent) Start(ctx context.Context, directory string) error {
	file, err := os.Open(c.sourcePath)
	if err != nil {
		return fmt.Errorf("capture: open DB diff source: %w", err)
	}
	c.file = file
	componentCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})
	go func() {
		defer close(c.done)
		_, _ = capture.CaptureDBDiffs(componentCtx, file, directory, nil)
	}()
	return nil
}
func (c *dbComponent) Stop() error {
	var closeErr error
	if c.file != nil {
		closeErr = c.file.Close()
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.done != nil {
		<-c.done
	}
	return closeErr
}

type logComponent struct {
	sourcePath string
	cancel     context.CancelFunc
	done       chan struct{}
}

func (c *logComponent) Name() string { return "logs" }
func (c *logComponent) Start(ctx context.Context, directory string) error {
	componentCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.done = make(chan struct{})
	go func() {
		defer close(c.done)
		_, _ = capture.TailStructuredLogFile(componentCtx, c.sourcePath, directory, 0, nil)
	}()
	return nil
}
func (c *logComponent) Stop() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.done != nil {
		<-c.done
	}
	return nil
}

func daemonArguments(request captureStartRequest, sessionPath string) []string {
	arguments := []string{"capture", "--daemon", "--url", request.TargetURL, "--session-dir", sessionPath}
	if request.ResultFile != "" {
		arguments = append(arguments, "--result-file", request.ResultFile)
	}
	if request.ComposeFile != "" {
		arguments = append(arguments, "--compose-file", request.ComposeFile)
	}
	if request.DBDiffFile != "" {
		arguments = append(arguments, "--db-diff-file", request.DBDiffFile)
	}
	if request.LogFile != "" {
		arguments = append(arguments, "--log-file", request.LogFile)
	}
	if request.PolicyFile != "" {
		arguments = append(arguments, "--policy-file", request.PolicyFile)
	}
	if request.UnsafeSkipSanitize {
		arguments = append(arguments, "--unsafe-skip-sanitize")
	}
	return arguments
}

func validateCaptureTarget(rawURL string) (*url.URL, error) {
	if strings.TrimSpace(rawURL) == "" {
		return nil, fmt.Errorf("capture: --url is required")
	}
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, fmt.Errorf("capture: invalid target URL: %w", err)
	}
	if (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return nil, fmt.Errorf("capture: target URL must use http:// or https://")
	}
	return parsedURL, nil
}

func writeCaptureStartResult(writer io.Writer, format string, result captureStartResult) error {
	if format == "json" {
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			return fmt.Errorf("capture: write JSON start result: %w", err)
		}
		return nil
	}
	_, err := fmt.Fprintf(writer, "Capture started: %s\n", result.SessionPath)
	return err
}

func waitForControlState(ctx context.Context, path string, resultPath string, expectedSessionID string, timeout time.Duration) (capture.ControlState, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		state, err := capture.ReadControlState(path)
		if err == nil {
			if expectedSessionID == "" || state.SessionID == expectedSessionID {
				return state, nil
			}
		}
		if contents, resultErr := os.ReadFile(resultPath); resultErr == nil {
			var result captureResultState
			if decodeErr := json.Unmarshal(contents, &result); decodeErr == nil &&
				(expectedSessionID == "" || result.SessionID == expectedSessionID) {
				_ = os.Remove(resultPath)
				if result.Error != "" {
					return capture.ControlState{}, fmt.Errorf("capture: daemon startup failed: %s", result.Error)
				}
				return capture.ControlState{}, fmt.Errorf("capture: daemon exited before publishing control state")
			}
		}
		select {
		case <-ctx.Done():
			return capture.ControlState{}, fmt.Errorf("capture: wait for daemon: %w", ctx.Err())
		case <-deadline.C:
			return capture.ControlState{}, fmt.Errorf("capture: daemon did not create matching control state within %s", timeout)
		case <-ticker.C:
		}
	}
}

func writeCaptureResultState(path string, state captureResultState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("capture: create result directory: %w", err)
	}
	contents, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("capture: serialize result state: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".dawg-result-*")
	if err != nil {
		return fmt.Errorf("capture: create temporary result state: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("capture: secure result state: %w", err)
	}
	if _, err := temporary.Write(append(contents, '\n')); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("capture: write result state %s: %w", path, err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("capture: sync result state %s: %w", path, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("capture: close result state %s: %w", path, err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("capture: write result state %s: %w", path, err)
	}
	return nil
}

// waitForCaptureResult polls for the daemon's result file, which is only
// written once the full capture->sanitize->package pipeline has finished.
// This lets `capture stop` return the real, absolute artifact path instead
// of a guessed one.
func waitForCaptureResult(ctx context.Context, path string, timeout time.Duration) (captureResultState, error) {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		if contents, err := os.ReadFile(path); err == nil {
			var state captureResultState
			if err := json.Unmarshal(contents, &state); err != nil {
				return captureResultState{}, fmt.Errorf("capture: decode result state %s: %w", path, err)
			}
			_ = os.Remove(path)
			return state, nil
		}
		select {
		case <-ctx.Done():
			return captureResultState{}, fmt.Errorf("capture: wait for packaging: %w", ctx.Err())
		case <-deadline.C:
			return captureResultState{}, fmt.Errorf("capture: packaging did not complete within %s", timeout)
		case <-ticker.C:
		}
	}
}

func newSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("capture: generate session ID: %w", err)
	}
	return fmt.Sprintf("%x", bytes), nil
}

func defaultEnginePolicy() string {
	return dawgenv.ResolvePolicy()
}

func newCaptureStopCommand() *cobra.Command {
	var controlFile string
	command := &cobra.Command{
		Use:   "stop",
		Short: "Stop the active capture session",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			absoluteControlFile, err := filepath.Abs(controlFile)
			if err != nil {
				return fmt.Errorf("capture: resolve control file: %w", err)
			}
			state, err := capture.ReadControlState(absoluteControlFile)
			if err != nil {
				return err
			}
			stopErr := capture.StopControlledSession(command.Context(), absoluteControlFile)
			daemonGone := errors.Is(stopErr, capture.ErrSessionAlreadyStopped)
			if stopErr != nil && !daemonGone {
				return stopErr
			}

			// Block until the daemon finishes sanitizing and packaging so the
			// caller gets back the real artifact path, not a guess.
			// 5 minutes: a long real-world session (rrweb traces + HTTP cassette +
			// OPA policy sanitization) can legitimately exceed 90s on large sites.
			// The "daemon already gone" fast path keeps 5s since in that case the
			// result file either already exists on disk or it never will.
			resultTimeout := 5 * time.Minute
			if daemonGone {
				resultTimeout = 5 * time.Second
			}
			resultFile := state.ResultFile
			if resultFile == "" {
				resultFile = defaultCaptureResultFile
			}
			result, err := waitForCaptureResult(command.Context(), resultFile, resultTimeout)
			if err != nil {
				if daemonGone {
					return fmt.Errorf("capture: session ended abnormally (daemon was not running)")
				}
				return err
			}
			if state.SessionID != "" && result.SessionID != "" && result.SessionID != state.SessionID {
				return fmt.Errorf("capture: result belongs to session %s, expected %s", result.SessionID, state.SessionID)
			}
			if result.Status == "error" {
				return fmt.Errorf("capture: packaging failed: %s", result.Error)
			}
			return writeCaptureStopResult(command.OutOrStdout(), outputFormat(command), captureStopResult{
				Status:       "packaged",
				ControlFile:  absoluteControlFile,
				SessionID:    result.SessionID,
				ArtifactPath: result.ArtifactPath,
			})
		},
	}
	command.Flags().StringVar(&controlFile, "control-file", defaultCaptureControlFile, "Capture daemon control-state file")
	return command
}

func writeCaptureStopResult(writer io.Writer, format string, result captureStopResult) error {
	if format == "json" {
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			return fmt.Errorf("capture: write JSON stop result: %w", err)
		}
		return nil
	}
	if result.ArtifactPath != "" {
		_, err := fmt.Fprintf(writer, "Capture packaged: %s\n", result.ArtifactPath)
		return err
	}
	_, err := fmt.Fprintln(writer, "Stopping capture session")
	return err
}
