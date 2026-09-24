package packager

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReplayTraceRequiresEventsAndFullSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		wantErr  string
	}{
		{name: "empty", contents: "", wantErr: "no rrweb events"},
		{name: "no timestamp", contents: "{\"type\":2}\n", wantErr: "no valid timestamps"},
		{name: "no full snapshot", contents: "{\"type\":3,\"timestamp\":1000}\n", wantErr: "no FullSnapshot"},
		{name: "valid", contents: "{\"type\":2,\"timestamp\":1000}\n{\"type\":3,\"timestamp\":1100}\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			traceDirectory := filepath.Join(directory, "traces")
			if err := os.MkdirAll(traceDirectory, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(traceDirectory, "rrweb.jsonl"), []byte(test.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			err := validateReplayTrace(directory)
			if test.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("expected %q error, got %v", test.wantErr, err)
			}
		})
	}
}
