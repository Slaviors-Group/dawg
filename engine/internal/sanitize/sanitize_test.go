package sanitize

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/dawg-placeholder/dawg/engine/internal/dawgtypes"
)

func TestSanitizeFilesRedactsHeadersAndJSONBodies(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "http", "frontend.jsonl")

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create capture directory: %v", err)
	}
	line := `{"request":{"headers":{"authorization":"Bearer private","cookie":"session=private"},"body":"{\"email\":\"person@example.test\",\"name\":\"Ada\"}"},"response":{"body":"{}"}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatalf("write capture: %v", err)
	}

	report, err := SanitizeFiles(directory, "policy.rego", "1.2.0")
	if err != nil {
		t.Fatalf("sanitize files: %v", err)
	}
	if report.FieldsRedacted != 4 || report.OPAResult != "not-evaluated" {
		t.Fatalf("unexpected report: %#v", report)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read sanitized capture: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(contents, &document); err != nil {
		t.Fatalf("decode sanitized capture: %v", err)
	}
	request := document["request"].(map[string]any)
	headers := request["headers"].(map[string]any)
	if headers["authorization"] != redactedValue || headers["cookie"] != redactedValue {
		t.Fatalf("headers were not redacted: %#v", headers)
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(request["body"].(string)), &body); err != nil {
		t.Fatalf("decode sanitized body: %v", err)
	}
	if body["email"] == "person@example.test" || !emailPattern.MatchString(body["email"]) {
		t.Fatalf("email was not replaced: %q", body["email"])
	}
	if body["name"] == "Ada" {
		t.Fatalf("name was not replaced: %q", body["name"])
	}
}

func TestSanitizeFilesPreservesOriginalOnInvalidJSONL(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "logs", "structured.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create capture directory: %v", err)
	}
	contents := []byte("not-json\n")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write capture: %v", err)
	}

	if _, err := SanitizeFiles(directory, "policy.rego", "1.2.0"); err == nil {
		t.Fatal("expected invalid JSONL error")
	}
	actual, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read original capture: %v", err)
	}
	if string(actual) != string(contents) {
		t.Fatalf("invalid capture was modified: %q", actual)
	}
}

func TestSanitizeDirectoryAllowsExportAndWritesReport(t *testing.T) {
	directory := t.TempDir()
	policyPath := writePolicy(t, directory, `
package dawg.sanitizer

import rego.v1

default allow := false

allow if {
  input.fieldsRedacted >= 0
  count(input.blockedFields) == 0
}
`)

	report, err := SanitizeDirectory(context.Background(), directory, policyPath, "1.2.0")
	if err != nil {
		t.Fatalf("sanitize directory: %v", err)
	}
	if !report.ExportAllowed || report.OPAResult != "allow" {
		t.Fatalf("expected allowed export report, got %#v", report)
	}
	contents, err := os.ReadFile(filepath.Join(directory, sanitizeReportFileName))
	if err != nil {
		t.Fatalf("read sanitization report: %v", err)
	}
	var written dawgtypes.SanitizeReport
	if err := json.Unmarshal(contents, &written); err != nil {
		t.Fatalf("decode sanitization report: %v", err)
	}
	if !written.ExportAllowed || written.OPAResult != "allow" {
		t.Fatalf("unexpected written report: %#v", written)
	}
}

func TestSanitizeDirectoryBlocksExportAndWritesReport(t *testing.T) {
	directory := t.TempDir()
	policyPath := writePolicy(t, directory, `
package dawg.sanitizer

default allow := false
`)

	report, err := SanitizeDirectory(context.Background(), directory, policyPath, "1.2.0")
	if !errors.Is(err, dawgtypes.ErrExportBlocked) {
		t.Fatalf("expected export blocked error, got %v", err)
	}
	if report.ExportAllowed || report.OPAResult != "deny" || len(report.BlockedFields) == 0 {
		t.Fatalf("expected denied export report, got %#v", report)
	}
	contents, err := os.ReadFile(filepath.Join(directory, sanitizeReportFileName))
	if err != nil {
		t.Fatalf("read denial report: %v", err)
	}
	var written dawgtypes.SanitizeReport
	if err := json.Unmarshal(contents, &written); err != nil {
		t.Fatalf("decode denial report: %v", err)
	}
	if written.ExportAllowed || written.OPAResult != "deny" {
		t.Fatalf("unexpected written denial report: %#v", written)
	}
}

func TestSanitizeDirectoryFailsWhenPolicyIsUnavailable(t *testing.T) {
	directory := t.TempDir()
	_, err := SanitizeDirectory(context.Background(), directory, filepath.Join(directory, "missing.rego"), "1.2.0")
	if err == nil {
		t.Fatal("expected unavailable policy error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing policy error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(directory, sanitizeReportFileName)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("policy failure unexpectedly wrote a report: %v", err)
	}
}

func writePolicy(t *testing.T, directory, policy string) string {
	t.Helper()
	policyPath := filepath.Join(directory, "policy.rego")
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("write policy: %v", err)
	}
	return policyPath
}
