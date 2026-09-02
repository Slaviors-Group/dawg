package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/capture"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/spf13/cobra"
)

const defaultCaptureControlFile = ".dawg/capture-control.json"

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
	var proxyPort int
	var internalHosts []string
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
				ProxyPort:        proxyPort,
				InternalHosts:    internalHosts,
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
	command.Flags().StringVar(&sessionDirectory, "session-dir", ".dawg/captures", "Directory containing capture sessions")
	command.Flags().StringVar(&browserScript, "browser-script", defaultEngineScript("capture-browser.cjs"), "Playwright capture script")
	command.Flags().StringVar(&proxyAddon, "proxy-addon", defaultEngineScript("capture-proxy.py"), "mitmproxy capture addon")
	command.Flags().IntVar(&proxyPort, "proxy-port", 8081, "mitmproxy listen port")
	command.Flags().StringSliceVar(&internalHosts, "internal-host", nil, "Internal host captured as backend traffic")
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
	ProxyPort        int
	InternalHosts    []string
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
	process := exec.CommandContext(ctx, executable, arguments...)
	process.Stdout = io.Discard
	process.Stderr = io.Discard
	if err := process.Start(); err != nil {
		return captureStartResult{}, fmt.Errorf("capture: start daemon: %w", err)
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
	session, err := capture.NewSession(capture.SessionOptions{
		Directory: request.SessionDirectory,
		Metadata: dawgtypes.CaptureMetadata{
			SessionID:   sessionID,
			TargetURL:   request.TargetURL,
			ActionTrace: &dawgtypes.ActionTrace{Path: "actions/browser.jsonl", Version: "0.1.0"},
		},
		Components: []capture.SessionComponent{
			&browserComponent{recorder: &capture.BrowserRecorder{ScriptPath: request.BrowserScript}, targetURL: request.TargetURL},
			&proxyComponent{manager: &capture.ProxyManager{AddonPath: request.ProxyAddon}, port: request.ProxyPort, internalHosts: request.InternalHosts},
		},
	})
	if err != nil {
		return err
	}
	return capture.RunControlledSession(ctx, session, defaultCaptureControlFile)
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

func daemonArguments(request captureStartRequest, sessionPath string) []string {
	arguments := []string{"capture", "--daemon", "--url", request.TargetURL, "--session-dir", sessionPath, "--browser-script", request.BrowserScript, "--proxy-addon", request.ProxyAddon, "--proxy-port", fmt.Sprintf("%d", request.ProxyPort)}
	for _, host := range request.InternalHosts {
		arguments = append(arguments, "--internal-host", host)
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

func defaultEngineScript(name string) string {
	executable, err := os.Executable()
	if err != nil {
		return filepath.Join("scripts", name)
	}
	return filepath.Join(filepath.Dir(executable), "scripts", name)
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
