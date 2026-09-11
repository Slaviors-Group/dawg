package dawgenv

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResourceDirOverride(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("DAWG_RESOURCES_DIR", tempDir)

	if got := ResourceDir(); got != tempDir {
		t.Fatalf("ResourceDir() = %q, want %q", got, tempDir)
	}
}

func TestResolveMitmdump(t *testing.T) {
	tempDir := t.TempDir()
	mockBin := filepath.Join(tempDir, "mitmdump.exe")
	if err := os.WriteFile(mockBin, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DAWG_MITMDUMP_PATH", mockBin)
	if got := ResolveMitmdump(); got != mockBin {
		t.Fatalf("ResolveMitmdump() = %q, want %q", got, mockBin)
	}
}

func TestResolveNode(t *testing.T) {
	tempDir := t.TempDir()
	mockBin := filepath.Join(tempDir, "node.exe")
	if err := os.WriteFile(mockBin, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DAWG_NODE_PATH", mockBin)
	if got := ResolveNode(); got != mockBin {
		t.Fatalf("ResolveNode() = %q, want %q", got, mockBin)
	}
}

func TestResolveChromiumExecutableFromBundledBrowsers(t *testing.T) {
	resourceDirectory := t.TempDir()
	t.Setenv("DAWG_RESOURCES_DIR", resourceDirectory)
	t.Setenv("PLAYWRIGHT_BROWSERS_PATH", "")
	t.Setenv("DAWG_CHROMIUM_EXECUTABLE_PATH", "")

	var relativeExecutable string
	switch runtime.GOOS {
	case "windows":
		relativeExecutable = filepath.Join("chromium-1234", "chrome-win64", "chrome.exe")
	case "darwin":
		relativeExecutable = filepath.Join("chromium-1234", "chrome-mac", "Chromium.app", "Contents", "MacOS", "Chromium")
	default:
		relativeExecutable = filepath.Join("chromium-1234", "chrome-linux", "chrome")
	}
	executable := filepath.Join(resourceDirectory, "browsers", relativeExecutable)
	if err := os.MkdirAll(filepath.Dir(executable), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(executable, []byte("browser"), 0o700); err != nil {
		t.Fatal(err)
	}
	if got := ResolveChromiumExecutable(); got != executable {
		t.Fatalf("ResolveChromiumExecutable() = %q, want %q", got, executable)
	}
}

func TestResolveScript(t *testing.T) {
	tempDir := t.TempDir()
	scriptsDir := filepath.Join(tempDir, "scripts")
	if err := os.MkdirAll(scriptsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	mockScript := filepath.Join(scriptsDir, "test-script.cjs")
	if err := os.WriteFile(mockScript, []byte("// script\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("DAWG_RESOURCES_DIR", tempDir)
	if got := ResolveScript("test-script.cjs"); got != mockScript {
		t.Fatalf("ResolveScript() = %q, want %q", got, mockScript)
	}
}

func TestDoctorReport(t *testing.T) {
	ctx := context.Background()
	report := RunDoctor(ctx, "0.1.4-test")

	if report.EnginePath == "" {
		t.Errorf("expected non-empty EnginePath")
	}
	if len(report.Components) == 0 {
		t.Errorf("expected non-empty Components in report")
	}
}
