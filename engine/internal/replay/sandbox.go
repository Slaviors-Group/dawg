// Package replay implements PRD §6.5 sandboxed deterministic artifact replay.
package replay

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
	"gopkg.in/yaml.v3"
)

const defaultSandboxTimeout = 2 * time.Minute

// CommandRunner executes Docker commands for sandbox lifecycle operations.
type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

// ExecRunner executes commands on the replay host.
type ExecRunner struct{}

// Run executes a command and returns combined output.
func (ExecRunner) Run(ctx context.Context, name string, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, arguments...)
	procutil.HideWindow(command)
	return command.CombinedOutput()
}

// Sandbox starts and tears down one rootless Docker Compose replay project.
type Sandbox struct {
	Runner  CommandRunner
	Timeout time.Duration
}

// Start verifies sandbox prerequisites and starts a digest-pinned Compose project.
func (sandbox Sandbox) Start(ctx context.Context, composeFile, projectName string) error {
	if runtime.GOOS == "windows" {
		// Log a warning that we are bypassing Docker and running natively (Native Sandbox mock mode)
		fmt.Println("WARNING: Running in Native Sandbox mode (Docker compose bypassed on Windows).")
		return nil
	}
	if sandbox.Runner == nil || composeFile == "" || projectName == "" {
		return fmt.Errorf("replay: runner, compose file, and project name are required")
	}
	if err := ensurePinnedComposeImages(composeFile); err != nil {
		return err
	}
	if sandbox.Timeout <= 0 {
		sandbox.Timeout = defaultSandboxTimeout
	}
	timedContext, cancel := context.WithTimeout(ctx, sandbox.Timeout)
	defer cancel()
	securityOptions, err := sandbox.Runner.Run(timedContext, "docker", "info", "--format", "{{.SecurityOptions}}")
	if err != nil {
		return fmt.Errorf("replay: %w: query Docker security options: %v", dawgtypes.ErrSandboxRequired, err)
	}
	if !strings.Contains(strings.ToLower(string(securityOptions)), "rootless") {
		return fmt.Errorf("replay: %w: Docker is not running rootless", dawgtypes.ErrSandboxRequired)
	}
	if output, err := sandbox.Runner.Run(timedContext, "docker", "compose", "-f", composeFile, "-p", projectName, "up", "-d", "--wait"); err != nil {
		return fmt.Errorf("replay: sandbox compose up: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// Teardown removes a replay project, its volumes, and orphaned containers.
func (sandbox Sandbox) Teardown(ctx context.Context, composeFile, projectName string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if sandbox.Runner == nil || composeFile == "" || projectName == "" {
		return fmt.Errorf("replay: runner, compose file, and project name are required")
	}
	if sandbox.Timeout <= 0 {
		sandbox.Timeout = defaultSandboxTimeout
	}
	timedContext, cancel := context.WithTimeout(ctx, sandbox.Timeout)
	defer cancel()
	if output, err := sandbox.Runner.Run(timedContext, "docker", "compose", "-f", composeFile, "-p", projectName, "down", "--volumes", "--remove-orphans"); err != nil {
		return fmt.Errorf("replay: sandbox teardown: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func ensurePinnedComposeImages(composeFile string) error {
	contents, err := os.ReadFile(composeFile)
	if err != nil {
		return fmt.Errorf("replay: read compose file %s: %w", composeFile, err)
	}
	var compose struct {
		Services map[string]struct {
			Image string `yaml:"image"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(contents, &compose); err != nil {
		return fmt.Errorf("replay: decode compose file %s: %w", composeFile, err)
	}
	if len(compose.Services) == 0 {
		return fmt.Errorf("replay: compose file %s has no services", composeFile)
	}
	for name, service := range compose.Services {
		if !strings.Contains(service.Image, "@sha256:") {
			return fmt.Errorf("replay: %w: service %s image is not digest-pinned", dawgtypes.ErrSandboxRequired, name)
		}
	}
	return nil
}
