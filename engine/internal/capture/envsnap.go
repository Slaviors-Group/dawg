package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
	"gopkg.in/yaml.v3"
)

const environmentSnapshotTimeout = 30 * time.Second

// CommandRunner executes external commands for environment capture.
type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

// ExecRunner invokes commands through the host operating system.
type ExecRunner struct{}

// Run executes a command and returns its combined output.
func (ExecRunner) Run(ctx context.Context, name string, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, arguments...)
	procutil.HideWindow(command)
	return command.CombinedOutput()
}

// EnvironmentSnapshotRequest identifies the Compose file and capture session destination.
type EnvironmentSnapshotRequest struct {
	ComposeFile      string
	SessionDirectory string
}

// EnvironmentLockfile is the pinned environment definition stored in env/lockfile.json.
type EnvironmentLockfile struct {
	SchemaVersion string            `json:"schemaVersion"`
	ComposeFile   string            `json:"composeFile"`
	ImageDigests  map[string]string `json:"imageDigests"`
}

// EnvironmentSnapshot writes a copied Compose file and digest-pinned lockfile to a capture session.
type EnvironmentSnapshot struct {
	LockfilePath       string                       `json:"lockfilePath"`
	ComposePath        string                       `json:"composePath"`
	CaptureEnvironment dawgtypes.CaptureEnvironment `json:"captureEnvironment"`
}

type composeDocument struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Image string `yaml:"image"`
}

// SnapshotEnvironment resolves Compose image tags to immutable local Docker digests.
func SnapshotEnvironment(ctx context.Context, runner CommandRunner, request EnvironmentSnapshotRequest) (EnvironmentSnapshot, error) {
	if runner == nil {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: command runner is required")
	}
	if request.ComposeFile == "" || request.SessionDirectory == "" {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: compose file and session directory are required")
	}
	contents, err := os.ReadFile(request.ComposeFile)
	if err != nil {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: read Compose file %s: %w", request.ComposeFile, err)
	}
	var compose composeDocument
	if err := yaml.Unmarshal(contents, &compose); err != nil {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: decode Compose file %s: %w", request.ComposeFile, err)
	}
	if len(compose.Services) == 0 {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: Compose file %s has no services", request.ComposeFile)
	}

	timedContext, cancel := context.WithTimeout(ctx, environmentSnapshotTimeout)
	defer cancel()
	if _, err := runner.Run(timedContext, "docker", "version", "--format", "{{.Server.Version}}"); err != nil {
		return EnvironmentSnapshot{}, fmt.Errorf("%w: verify daemon: %v", dawgtypes.ErrDockerUnavailable, err)
	}

	digests := make(map[string]string, len(compose.Services))
	for serviceName, service := range compose.Services {
		if service.Image == "" {
			return EnvironmentSnapshot{}, fmt.Errorf("capture: service %s has no image", serviceName)
		}
		digest, err := resolveImageDigest(timedContext, runner, service.Image)
		if err != nil {
			return EnvironmentSnapshot{}, fmt.Errorf("capture: resolve image for service %s: %w", serviceName, err)
		}
		digests[serviceName] = digest
	}

	environmentDirectory := filepath.Join(request.SessionDirectory, "env")
	if err := os.MkdirAll(environmentDirectory, 0o700); err != nil {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: create environment directory: %w", err)
	}
	composePath := filepath.Join(environmentDirectory, "compose.yaml")
	if err := os.WriteFile(composePath, contents, 0o600); err != nil {
		return EnvironmentSnapshot{}, fmt.Errorf("capture: write captured Compose file: %w", err)
	}
	lockfilePath := filepath.Join(environmentDirectory, "lockfile.json")
	lockfile := EnvironmentLockfile{
		SchemaVersion: "0.1.4-alpha",
		ComposeFile:   "env/compose.yaml",
		ImageDigests:  digests,
	}
	if err := writeLockfile(lockfilePath, lockfile); err != nil {
		return EnvironmentSnapshot{}, err
	}
	return EnvironmentSnapshot{
		LockfilePath: lockfilePath,
		ComposePath:  composePath,
		CaptureEnvironment: dawgtypes.CaptureEnvironment{
			DockerComposeFile: "env/compose.yaml",
			ImageDigests:      digests,
		},
	}, nil
}

func resolveImageDigest(ctx context.Context, runner CommandRunner, image string) (string, error) {
	if strings.Contains(image, "@sha256:") {
		return image, nil
	}
	output, err := runner.Run(ctx, "docker", "image", "inspect", image, "--format", "{{index .RepoDigests 0}}")
	if err != nil {
		return "", fmt.Errorf("inspect image %s: %w", image, err)
	}
	digest := strings.TrimSpace(string(output))
	if !strings.Contains(digest, "@sha256:") {
		return "", fmt.Errorf("image %s did not resolve to a digest", image)
	}
	return digest, nil
}

func writeLockfile(path string, lockfile EnvironmentLockfile) error {
	contents, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return fmt.Errorf("capture: serialize environment lockfile: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		return fmt.Errorf("capture: write environment lockfile %s: %w", path, err)
	}
	return nil
}

// IsDockerUnavailable reports whether a snapshot failure was caused by Docker daemon availability.
func IsDockerUnavailable(err error) bool {
	return errors.Is(err, dawgtypes.ErrDockerUnavailable)
}
