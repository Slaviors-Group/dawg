package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

func TestRootCommandPrintsHelp(t *testing.T) {
	command := newRootCommand()
	buffer := new(bytes.Buffer)
	command.SetOut(buffer)
	command.SetArgs(nil)

	if err := command.Execute(); err != nil {
		t.Fatalf("execute root command: %v", err)
	}
	if !strings.Contains(buffer.String(), "Package reproducible web-app bug reports") {
		t.Fatalf("help output did not contain command description: %q", buffer.String())
	}
}

func TestRunCommandRegistersInteractiveFlag(t *testing.T) {
	command := newRunCommand()
	flag := command.Flags().Lookup("interactive")
	if flag == nil || flag.DefValue != "false" {
		t.Fatalf("interactive flag was not registered correctly: %#v", flag)
	}
}

func TestRootCommandRejectsUnknownOutput(t *testing.T) {
	command := newRootCommand()
	command.SetArgs([]string{"--output", "xml"})

	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid output format") {
		t.Fatalf("expected invalid output error, got %v", err)
	}
}

func TestInitCreatesDefaultConfig(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DAWG_STATE_DIR", t.TempDir())
	command := newRootCommand()
	buffer := new(bytes.Buffer)
	command.SetOut(buffer)
	command.SetArgs([]string{"--output", "json", "init", directory})

	if err := command.Execute(); err != nil {
		t.Fatalf("execute init: %v", err)
	}
	contents, err := os.ReadFile(filepath.Join(directory, configFileName))
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	expected := fmt.Sprintf(defaultConfig, defaultDawgDir("captures"), defaultDawgDir("policies", "default.rego"))
	if string(contents) != expected {
		t.Fatalf("unexpected config contents: %q", contents)
	}
	var result initResult
	if err := json.Unmarshal(buffer.Bytes(), &result); err != nil {
		t.Fatalf("decode init result: %v", err)
	}
	if result.Status != "created" || result.Overwritten {
		t.Fatalf("unexpected init result: %#v", result)
	}
}

func TestInitRefusesToOverwriteWithoutForce(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DAWG_STATE_DIR", t.TempDir())
	configPath := filepath.Join(directory, configFileName)
	if err := os.WriteFile(configPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	command := newRootCommand()
	command.SetArgs([]string{"init", directory})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "use --force") {
		t.Fatalf("expected overwrite error, got %v", err)
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read existing config: %v", err)
	}
	if string(contents) != "existing" {
		t.Fatalf("config was unexpectedly overwritten: %q", contents)
	}
}

func TestInitForceOverwritesConfig(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("DAWG_STATE_DIR", t.TempDir())
	configPath := filepath.Join(directory, configFileName)
	if err := os.WriteFile(configPath, []byte("existing"), 0o600); err != nil {
		t.Fatalf("write existing config: %v", err)
	}

	command := newRootCommand()
	command.SetArgs([]string{"init", "--force", directory})
	if err := command.Execute(); err != nil {
		t.Fatalf("force init: %v", err)
	}
	contents, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read overwritten config: %v", err)
	}
	expected := fmt.Sprintf(defaultConfig, defaultDawgDir("captures"), defaultDawgDir("policies", "default.rego"))
	if string(contents) != expected {
		t.Fatalf("config was not overwritten: %q", contents)
	}
}

func TestCaptureStopReportsMissingControlFile(t *testing.T) {
	command := newRootCommand()
	controlFile := filepath.Join(t.TempDir(), "missing.json")
	command.SetArgs([]string{"capture", "stop", "--control-file", controlFile})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "read control state") {
		t.Fatalf("expected missing control-state error, got %v", err)
	}
}

func TestCaptureDaemonArgumentsUseExtensionCapture(t *testing.T) {
	arguments := daemonArguments(captureStartRequest{
		TargetURL:  "http://127.0.0.1:3000",
		ResultFile: filepath.Join(t.TempDir(), "capture-result.json"),
	}, t.TempDir())
	joined := strings.Join(arguments, " ")
	if strings.Contains(joined, "playwright") || strings.Contains(joined, "cdp-endpoint") || strings.Contains(joined, "proxy-addon") {
		t.Fatalf("extension capture inherited a legacy browser/proxy argument: %v", arguments)
	}
	if !strings.Contains(joined, "--url http://127.0.0.1:3000") || !strings.Contains(joined, "--result-file") {
		t.Fatalf("daemon arguments omitted extension capture state: %v", arguments)
	}
}

func TestWaitForControlStateReportsDaemonStartupFailure(t *testing.T) {
	directory := t.TempDir()
	resultPath := filepath.Join(directory, "result.json")
	contents, err := json.Marshal(captureResultState{
		Status:    "error",
		SessionID: "session-1",
		Error:     "browser extension did not connect",
	})
	if err != nil {
		t.Fatalf("serialize result: %v", err)
	}
	if err := os.WriteFile(resultPath, contents, 0o600); err != nil {
		t.Fatalf("write result: %v", err)
	}

	_, err = waitForControlState(context.Background(), filepath.Join(directory, "control.json"), resultPath, "session-1", time.Second)
	if err == nil || !strings.Contains(err.Error(), "browser extension did not connect") {
		t.Fatalf("expected daemon startup failure, got %v", err)
	}
	if _, statErr := os.Stat(resultPath); !os.IsNotExist(statErr) {
		t.Fatalf("startup result was not consumed: %v", statErr)
	}
}

func TestInspectReturnsValidatedManifestJSON(t *testing.T) {
	directory := t.TempDir()
	fixturePath := filepath.Join("..", "..", "internal", "manifest", "testdata", "valid.json")
	contents, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "dawg-manifest.json"), contents, 0o600); err != nil {
		t.Fatalf("write artifact manifest: %v", err)
	}
	command := newRootCommand()
	buffer := new(bytes.Buffer)
	command.SetOut(buffer)
	command.SetArgs([]string{"--output", "json", "inspect", directory})
	if err := command.Execute(); err != nil {
		t.Fatalf("inspect artifact: %v", err)
	}
	var value map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &value); err != nil {
		t.Fatalf("decode inspect output: %v", err)
	}
	if value["schemaVersion"] != manifest.SchemaVersion || value["title"] == "" {
		t.Fatalf("unexpected inspect output: %#v", value)
	}
}

func TestInspectAcceptsLegacyManifest(t *testing.T) {
	directory := t.TempDir()
	fixturePath := filepath.Join("..", "..", "internal", "manifest", "testdata", "valid.json")
	contents, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	var legacy map[string]any
	if err := json.Unmarshal(contents, &legacy); err != nil {
		t.Fatalf("decode manifest fixture: %v", err)
	}
	delete(legacy, "diagnostics")
	legacy["schemaVersion"] = "0.2.3-naughty"
	legacy["layers"] = []any{map[string]any{
		"mediaType": string(dawgtypes.MediaTypeEnvironment),
		"digest":    "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		"size":      2048,
	}}
	contents, err = json.Marshal(legacy)
	if err != nil {
		t.Fatalf("encode legacy manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(directory, "dawg-manifest.json"), contents, 0o600); err != nil {
		t.Fatalf("write artifact manifest: %v", err)
	}

	command := newRootCommand()
	buffer := new(bytes.Buffer)
	command.SetOut(buffer)
	command.SetArgs([]string{"--output", "json", "inspect", directory})
	if err := command.Execute(); err != nil {
		t.Fatalf("inspect legacy artifact: %v", err)
	}
	if !strings.Contains(buffer.String(), `"schemaVersion": "0.2.3-naughty"`) {
		t.Fatalf("inspect output did not preserve legacy version: %s", buffer.String())
	}
}

func TestInspectRejectsUnsupportedManifestVersion(t *testing.T) {
	directory := t.TempDir()
	fixturePath := filepath.Join("..", "..", "internal", "manifest", "testdata", "valid.json")
	contents, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	contents = bytes.Replace(contents, []byte(manifest.SchemaVersion), []byte("0.2.4-naughty"), 1)
	if err := os.WriteFile(filepath.Join(directory, "dawg-manifest.json"), contents, 0o600); err != nil {
		t.Fatalf("write artifact manifest: %v", err)
	}

	command := newRootCommand()
	command.SetArgs([]string{"inspect", directory})
	err = command.Execute()
	if err == nil || !strings.Contains(err.Error(), "unsupported schema version") {
		t.Fatalf("expected unsupported schema version error, got %v", err)
	}
}

func TestInspectRejectsAdditionalManifestProperties(t *testing.T) {
	directory := t.TempDir()
	fixturePath := filepath.Join("..", "..", "internal", "manifest", "testdata", "valid.json")
	contents, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("read manifest fixture: %v", err)
	}
	contents = bytes.Replace(contents, []byte(`  "title":`), []byte("  \"unexpected\": true,\n  \"title\":"), 1)
	if err := os.WriteFile(filepath.Join(directory, "dawg-manifest.json"), contents, 0o600); err != nil {
		t.Fatalf("write artifact manifest: %v", err)
	}

	command := newRootCommand()
	command.SetArgs([]string{"inspect", directory})
	err = command.Execute()
	if err == nil || !strings.Contains(err.Error(), "schema validation") {
		t.Fatalf("expected schema validation error, got %v", err)
	}
}

func TestPushRequiresRegistryReference(t *testing.T) {
	command := newRootCommand()
	command.SetArgs([]string{"push", t.TempDir()})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "--registry is required") {
		t.Fatalf("expected required registry error, got %v", err)
	}
}
