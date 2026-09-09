package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if string(contents) != defaultConfig {
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
	if string(contents) != defaultConfig {
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
	if value["schemaVersion"] != "0.1.3-alpha" || value["title"] == "" {
		t.Fatalf("unexpected inspect output: %#v", value)
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
