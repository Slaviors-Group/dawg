package dawgenv

import (
	"context"
	"os"
	"path/filepath"
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
	report := RunDoctor(ctx, "0.1.1-test")

	if report.EnginePath == "" {
		t.Errorf("expected non-empty EnginePath")
	}
	if len(report.Components) == 0 {
		t.Errorf("expected non-empty Components in report")
	}
}
