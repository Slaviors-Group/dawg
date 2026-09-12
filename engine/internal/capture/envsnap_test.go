package capture

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func TestSnapshotEnvironmentPinsComposeImagesAndWritesLockfile(t *testing.T) {
	root := t.TempDir()
	composeFile := filepath.Join(root, "compose.yaml")
	composeContents := []byte("services:\n  backend:\n    image: example/backend:1.2.3\n  postgres:\n    image: postgres:16\n")
	if err := os.WriteFile(composeFile, composeContents, 0o600); err != nil {
		t.Fatalf("write Compose file: %v", err)
	}
	runner := scriptedRunner{responses: map[string]string{
		"docker version --format {{.Server.Version}}":                                  "27.0.0\n",
		"docker image inspect example/backend:1.2.3 --format {{index .RepoDigests 0}}": "example/backend@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\n",
		"docker image inspect postgres:16 --format {{index .RepoDigests 0}}":           "postgres@sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\n",
	}}

	snapshot, err := SnapshotEnvironment(context.Background(), runner, EnvironmentSnapshotRequest{
		ComposeFile:      composeFile,
		SessionDirectory: filepath.Join(root, "session"),
	})
	if err != nil {
		t.Fatalf("snapshot environment: %v", err)
	}
	if snapshot.CaptureEnvironment.DockerComposeFile != "env/compose.yaml" {
		t.Fatalf("unexpected Compose file: %q", snapshot.CaptureEnvironment.DockerComposeFile)
	}
	if snapshot.CaptureEnvironment.ImageDigests["backend"] != "example/backend@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("backend digest was not pinned: %#v", snapshot.CaptureEnvironment.ImageDigests)
	}
	var lockfile EnvironmentLockfile
	contents, err := os.ReadFile(snapshot.LockfilePath)
	if err != nil {
		t.Fatalf("read lockfile: %v", err)
	}
	if err := json.Unmarshal(contents, &lockfile); err != nil {
		t.Fatalf("decode lockfile: %v", err)
	}
	if lockfile.SchemaVersion != "0.2.3-naughty" || len(lockfile.ImageDigests) != 2 {
		t.Fatalf("unexpected lockfile: %#v", lockfile)
	}
	capturedCompose, err := os.ReadFile(snapshot.ComposePath)
	if err != nil {
		t.Fatalf("read captured Compose file: %v", err)
	}
	if string(capturedCompose) != string(composeContents) {
		t.Fatal("captured Compose file did not preserve source content")
	}
}

func TestSnapshotEnvironmentFailsWhenDockerIsUnavailable(t *testing.T) {
	root := t.TempDir()
	composeFile := filepath.Join(root, "compose.yaml")
	if err := os.WriteFile(composeFile, []byte("services:\n  backend:\n    image: example/backend:1.2.3\n"), 0o600); err != nil {
		t.Fatalf("write Compose file: %v", err)
	}
	runner := scriptedRunner{errors: map[string]error{
		"docker version --format {{.Server.Version}}": errors.New("daemon not running"),
	}}

	_, err := SnapshotEnvironment(context.Background(), runner, EnvironmentSnapshotRequest{
		ComposeFile:      composeFile,
		SessionDirectory: filepath.Join(root, "session"),
	})
	if !errors.Is(err, dawgtypes.ErrDockerUnavailable) || !IsDockerUnavailable(err) {
		t.Fatalf("expected Docker unavailable error, got %v", err)
	}
}

func TestSnapshotEnvironmentRejectsImageWithoutDigest(t *testing.T) {
	root := t.TempDir()
	composeFile := filepath.Join(root, "compose.yaml")
	if err := os.WriteFile(composeFile, []byte("services:\n  backend:\n    image: example/backend:1.2.3\n"), 0o600); err != nil {
		t.Fatalf("write Compose file: %v", err)
	}
	runner := scriptedRunner{responses: map[string]string{
		"docker version --format {{.Server.Version}}":                                  "27.0.0\n",
		"docker image inspect example/backend:1.2.3 --format {{index .RepoDigests 0}}": "<no value>\n",
	}}

	_, err := SnapshotEnvironment(context.Background(), runner, EnvironmentSnapshotRequest{
		ComposeFile:      composeFile,
		SessionDirectory: filepath.Join(root, "session"),
	})
	if err == nil {
		t.Fatal("expected unresolved image digest error")
	}
}

type scriptedRunner struct {
	responses map[string]string
	errors    map[string]error
}

func (runner scriptedRunner) Run(_ context.Context, name string, arguments ...string) ([]byte, error) {
	command := name + " " + joinArguments(arguments)
	if err := runner.errors[command]; err != nil {
		return nil, err
	}
	response, ok := runner.responses[command]
	if !ok {
		return nil, fmt.Errorf("unexpected command %q", command)
	}
	return []byte(response), nil
}

func joinArguments(arguments []string) string {
	if len(arguments) == 0 {
		return ""
	}
	result := arguments[0]
	for _, argument := range arguments[1:] {
		result += " " + argument
	}
	return result
}
