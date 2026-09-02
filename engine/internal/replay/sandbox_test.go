package replay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func TestSandboxRejectsBareWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific security boundary")
	}
	err := (Sandbox{Runner: scriptedRunner{}}).Start(context.Background(), "unused", "project")
	if !errors.Is(err, dawgtypes.ErrSandboxRequired) {
		t.Fatalf("expected sandbox requirement error, got %v", err)
	}
}

func TestEnsurePinnedComposeImagesRejectsTag(t *testing.T) {
	path := writeCompose(t, "services:\n  app:\n    image: example/app:latest\n")
	err := ensurePinnedComposeImages(path)
	if !errors.Is(err, dawgtypes.ErrSandboxRequired) || !strings.Contains(err.Error(), "not digest-pinned") {
		t.Fatalf("expected pinning error, got %v", err)
	}
}

func writeCompose(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "compose.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write compose: %v", err)
	}
	return path
}

type scriptedRunner struct{}

func (scriptedRunner) Run(_ context.Context, name string, arguments ...string) ([]byte, error) {
	return nil, fmt.Errorf("unexpected command %s %s", name, strings.Join(arguments, " "))
}
