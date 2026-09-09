package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

type captureStopResult struct {
	Status      string `json:"status"`
	ControlFile string `json:"controlFile"`
}

type captureStartResult struct {
	Status      string `json:"status"`
	SessionID   string `json:"sessionId"`
	SessionPath string `json:"sessionPath"`
}

func newCaptureCommand() *cobra.Command {
	var targetURL string
	var sessionDirectory string
	var browserScript string
	var proxyAddon string
	var mitmproxyPath string
	var proxyPort int
	var internalHosts []string
	var composeFile string
	var dbDiffFile string
	var logFile string
	var policyFile string
	var unsafeSkipSanitize bool
	var daemon bool

	command := &cobra.Command{
		Use:   "capture",
		Short: "Capture a web-app reproduction session",
		RunE: func(command *cobra.Command, _ []string) error {
			request := captureStartRequest{
				TargetURL:        targetURL,
				SessionDirectory: sessionDirectory,
				BrowserScript:    browserScript,
				ProxyAddon:       proxyAddon,
				MitmproxyPath:    mitmproxyPath,
				ProxyPort:        proxyPort,
				InternalHosts:    internalHosts,
				ComposeFile:        composeFile,
				DBDiffFile:         dbDiffFile,
				LogFile:            logFile,
				PolicyFile:         policyFile,
				UnsafeSkipSanitize: unsafeSkipSanitize,
			}
			if unsafeSkipSanitize {
				parsedURL, err := url.Parse(targetURL)
				if err != nil {
					return fmt.Errorf("capture: invalid target URL: %w", err)
				}
				hostname := parsedURL.Hostname()
				if hostname != "localhost" && hostname != "127.0.0.1" && hostname != "[::1]" {
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
	command.Flags().StringVar(&browserScript, "browser-script", defaultEngineScript("capture-browser.cjs"), "Playwright capture script")
	command.Flags().StringVar(&proxyAddon, "proxy-addon", defaultEngineScript("capture-proxy.py"), "mitmproxy capture addon")
	command.Flags().StringVar(&mitmproxyPath, "mitmproxy-path", defaultMitmdumpPath(), "Path to the mitmdump executable")
	command.Flags().IntVar(&proxyPort, "proxy-port", 8081, "mitmproxy listen port")
	command.Flags().StringSliceVar(&internalHosts, "internal-host", nil, "Internal host captured as backend traffic")
	command.Flags().StringVar(&composeFile, "compose-file", "", "Path to docker-compose file for environment snapshot")
	command.Flags().StringVar(&dbDiffFile, "db-diff-file", "", "Path to DB diff stream or file")
	command.Flags().StringVar(&logFile, "log-file", "", "Path to structured application log file")
	command.Flags().StringVar(&policyFile, "policy-file", defaultEnginePolicy(), "Path to OPA sanitization policy")
	command.Flags().BoolVar(&unsafeSkipSanitize, "unsafe-skip-sanitize", false, "Skip sanitization (only allowed for localhost targets)")
	command.Flags().BoolVar(&daemon, "daemon", false, "Run the capture owner process")
	_ = command.MarkFlagRequired("url")
	command.AddCommand(newCaptureStopCommand())
	return command
}

type captureStartRequest struct {
	TargetURL        string
	SessionDirectory string
	BrowserScript    string
	ProxyAddon       string
	MitmproxyPath    string
	ProxyPort        int
	InternalHosts    []string
	ComposeFile        string
	DBDiffFile         string
	LogFile            string
	PolicyFile         string
	UnsafeSkipSanitize bool
}

func launchCaptureDaemon(ctx context.Context, request captureStartRequest) (captureStartResult, error) {
	if request.TargetURL == "" {
		return captureStartResult{}, fmt.Errorf("capture: --url is required")
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
	executable, err := os.Executable()
	if err != nil {
		return captureStartResult{}, fmt.Errorf("capture: locate executable: %w", err)
	}
	arguments := daemonArguments(request, absoluteSessionPath)
	process := exec.Command(executable, arguments...)
	procutil.HideWindow(process)

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
	controlPath := defaultCaptureControlFile
	if err := waitForControlFile(ctx, controlPath, 20*time.Second); err != nil {
		return captureStartResult{}, err
	}
	return captureStartResult{Status: "capturing", SessionID: sessionID, SessionPath: absoluteSessionPath}, nil
}

func runCaptureDaemon(ctx context.Context, request captureStartRequest) error {
	if request.TargetURL == "" {
		return fmt.Errorf("capture: --url is required")
	}
	if request.BrowserScript == "" || request.ProxyAddon == "" {
		return fmt.Errorf("capture: browser script and proxy addon are required")
	}
	sessionID := filepath.Base(request.SessionDirectory)

	components := []capture.SessionComponent{
		&browserComponent{recorder: &capture.BrowserRecorder{NodeBinary: dawgenv.ResolveNode(), ScriptPath: request.BrowserScript}, targetURL: request.TargetURL},
		&proxyComponent{manager: &capture.ProxyManager{Executable: request.MitmproxyPath, AddonPath: request.ProxyAddon}, port: request.ProxyPort, internalHosts: request.InternalHosts},
	}
	if request.ComposeFile != "" {
		components = append(components, &envComponent{runner: capture.ExecRunner{}, composeFile: request.ComposeFile})
	}
	if request.DBDiffFile != "" {
		components = append(components, &dbComponent{sourcePath: request.DBDiffFile})
	}
	if request.LogFile != "" {
		components = append(components, &logComponent{sourcePath: request.LogFile})
	}

	session, err := capture.NewSession(capture.SessionOptions{
		Directory: request.SessionDirectory,
		Metadata: dawgtypes.CaptureMetadata{
			SessionID:   sessionID,
			TargetURL:   request.TargetURL,
			ActionTrace: &dawgtypes.ActionTrace{Path: "actions/browser.jsonl", Version: "0.1.3-alpha"},
		},
		Components: components,
	})
	if err != nil {
		return err
	}

	if err := capture.RunControlledSession(ctx, session, defaultCaptureControlFile); err != nil {
		return err
	}

	// Pipeline stage 2: Sanitize
	if !request.UnsafeSkipSanitize {
		_, err = sanitize.SanitizeDirectory(ctx, request.SessionDirectory, request.PolicyFile, "1.0.0")
		if err != nil {
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
	_, err = packager.Package(packageRequest)
	return err
}

type browserComponent struct {
	recorder  *capture.BrowserRecorder
	targetURL string
}

func (component *browserComponent) Name() string { return "browser" }
func (component *browserComponent) Start(ctx context.Context, directory string) error {
	return component.recorder.Start(ctx, capture.BrowserCaptureRequest{TargetURL: component.targetURL, SessionDirectory: directory})
}
func (component *browserComponent) Stop() error { return component.recorder.Stop() }

type proxyComponent struct {
	manager       *capture.ProxyManager
	port          int
	internalHosts []string
}

func (component *proxyComponent) Name() string { return "proxy" }
func (component *proxyComponent) Start(ctx context.Context, directory string) error {
	return component.manager.Start(ctx, capture.ProxyCaptureRequest{ListenPort: component.port, SessionDirectory: directory, InternalHosts: component.internalHosts})
}
func (component *proxyComponent) Stop() error { return component.manager.Stop() }

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
	arguments := []string{"capture", "--daemon", "--url", request.TargetURL, "--session-dir", sessionPath, "--browser-script", request.BrowserScript, "--proxy-addon", request.ProxyAddon, "--mitmproxy-path", request.MitmproxyPath, "--proxy-port", fmt.Sprintf("%d", request.ProxyPort)}
	for _, host := range request.InternalHosts {
		arguments = append(arguments, "--internal-host", host)
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

func waitForControlFile(ctx context.Context, path string, timeout time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("capture: wait for daemon: %w", ctx.Err())
		case <-deadline.C:
			return fmt.Errorf("capture: daemon did not create control state within %s", timeout)
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

func defaultMitmdumpPath() string {
	return dawgenv.ResolveMitmdump()
}

func defaultEnginePolicy() string {
	return dawgenv.ResolvePolicy()
}

func defaultEngineScript(name string) string {
	return dawgenv.ResolveScript(name)
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
			if err := capture.StopControlledSession(context.Background(), absoluteControlFile); err != nil {
				return err
			}
			return writeCaptureStopResult(command.OutOrStdout(), outputFormat(command), captureStopResult{
				Status:      "stopping",
				ControlFile: absoluteControlFile,
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
	_, err := fmt.Fprintln(writer, "Stopping capture session")
	return err
}
