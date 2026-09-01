package capture

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const defaultProxyStartupTimeout = 15 * time.Second

// ProxyManager manages the mitmproxy capture sidecar required by PRD §6.1.
type ProxyManager struct {
	Executable     string
	AddonPath      string
	Arguments      func(ProxyCaptureRequest, string) []string
	StartupTimeout time.Duration
	StopTimeout    time.Duration

	mu   sync.Mutex
	cmd  *exec.Cmd
	done chan error
}

// ProxyCaptureRequest configures a mitmproxy capture session.
type ProxyCaptureRequest struct {
	ListenPort       int
	SessionDirectory string
	InternalHosts    []string
}

// Start launches mitmproxy and waits until its DAWG addon declares readiness.
func (manager *ProxyManager) Start(ctx context.Context, request ProxyCaptureRequest) error {
	if request.ListenPort < 1 || request.ListenPort > 65535 || request.SessionDirectory == "" {
		return fmt.Errorf("capture: valid proxy port and session directory are required")
	}
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.cmd != nil {
		return fmt.Errorf("capture: proxy is already running")
	}
	if manager.Executable == "" {
		manager.Executable = "mitmdump"
	}
	if manager.AddonPath == "" {
		return fmt.Errorf("capture: mitmproxy addon path is required")
	}
	if manager.StartupTimeout <= 0 {
		manager.StartupTimeout = defaultProxyStartupTimeout
	}
	if manager.StopTimeout <= 0 {
		manager.StopTimeout = 5 * time.Second
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(request.ListenPort)))
	if err != nil {
		return fmt.Errorf("capture: proxy port %d unavailable: %w", request.ListenPort, err)
	}
	if err := listener.Close(); err != nil {
		return fmt.Errorf("capture: release proxy port %d: %w", request.ListenPort, err)
	}
	if err := os.MkdirAll(filepath.Join(request.SessionDirectory, "http"), 0o700); err != nil {
		return fmt.Errorf("capture: create proxy HTTP directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(request.SessionDirectory, "cassettes"), 0o700); err != nil {
		return fmt.Errorf("capture: create proxy cassette directory: %w", err)
	}

	arguments := defaultProxyArguments(request, manager.AddonPath)
	if manager.Arguments != nil {
		arguments = manager.Arguments(request, manager.AddonPath)
	}
	command := exec.CommandContext(ctx, manager.Executable, arguments...)
	internalHosts, err := json.Marshal(request.InternalHosts)
	if err != nil {
		return fmt.Errorf("capture: encode proxy internal hosts: %w", err)
	}
	command.Env = append(os.Environ(), "DAWG_SESSION_DIR="+request.SessionDirectory, "DAWG_INTERNAL_HOSTS="+string(internalHosts))
	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("capture: open proxy stdout: %w", err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return fmt.Errorf("capture: open proxy stderr: %w", err)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("capture: start mitmproxy: %w", err)
	}
	manager.cmd = command
	manager.done = make(chan error, 1)
	ready := make(chan error, 1)
	go watchCaptureProcess(command, stdout, stderr, ready, manager.done, "mitmproxy")

	timer := time.NewTimer(manager.StartupTimeout)
	defer timer.Stop()
	select {
	case err := <-ready:
		if err != nil {
			manager.clearLocked()
			return err
		}
		return nil
	case <-timer.C:
		_ = command.Process.Kill()
		<-manager.done
		manager.clearLocked()
		return fmt.Errorf("capture: mitmproxy startup timed out after %s", manager.StartupTimeout)
	case <-ctx.Done():
		_ = command.Process.Kill()
		<-manager.done
		manager.clearLocked()
		return fmt.Errorf("capture: start mitmproxy: %w", ctx.Err())
	}
}

func defaultProxyArguments(request ProxyCaptureRequest, addonPath string) []string {
	return []string{
		"--listen-port", strconv.Itoa(request.ListenPort),
		"--set", "dawg_output_dir=" + request.SessionDirectory,
		"-s", addonPath,
		"--ssl-insecure",
	}
}

// Stop requests graceful proxy shutdown before killing a process that fails to exit.
func (manager *ProxyManager) Stop() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.cmd == nil {
		return fmt.Errorf("capture: proxy is not running")
	}
	_ = manager.cmd.Process.Signal(os.Interrupt)
	timer := time.NewTimer(manager.StopTimeout)
	defer timer.Stop()
	select {
	case <-manager.done:
		manager.clearLocked()
		return nil
	case <-timer.C:
		if err := manager.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("capture: force stop mitmproxy: %w", err)
		}
		<-manager.done
		manager.clearLocked()
		return nil
	}
}

func (manager *ProxyManager) clearLocked() {
	manager.cmd = nil
	manager.done = nil
}

func watchCaptureProcess(command *exec.Cmd, stdout io.Reader, stderr io.Reader, ready chan<- error, done chan<- error, name string) {
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
			if !readyReported {
				ready <- fmt.Errorf("capture: decode %s status: %w", name, err)
			}
			continue
		}
		if status.Status == "capturing" && !readyReported {
			ready <- nil
			readyReported = true
		}
	}
	if err := scanner.Err(); err != nil && !readyReported {
		ready <- fmt.Errorf("capture: read %s status: %w", name, err)
	}
	err := command.Wait()
	stderrContents := <-stderrDone
	if err != nil && len(stderrContents) > 0 {
		err = fmt.Errorf("%w: %s", err, string(stderrContents))
	}
	if !readyReported {
		if err == nil {
			err = fmt.Errorf("capture: %s exited before readiness", name)
		}
		ready <- err
	}
	done <- err
}
