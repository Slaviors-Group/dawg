package capture

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestBrowserRecorderWritesContractedTraceFiles(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, `
const fs = require("node:fs");
const path = require("node:path");
const options = Object.fromEntries(process.argv.slice(2).reduce((pairs, value, index, values) => index % 2 === 0 ? [...pairs, [value.slice(2), values[index + 1]]] : pairs, []));
for (const key of ["rrweb-output", "http-output", "actions-output"]) fs.mkdirSync(path.dirname(options[key]), { recursive: true });
fs.writeFileSync(options["rrweb-output"], "{\"timestamp\":1}\n");
fs.writeFileSync(options["http-output"], "{\"id\":\"request-1\",\"request\":{},\"response\":{}}\n");
fs.writeFileSync(options["actions-output"], "{\"type\":\"click\"}\n");
console.log(JSON.stringify({ status: "capturing" }));
setInterval(() => {}, 1000);
`)
	recorder := BrowserRecorder{NodeBinary: "node", ScriptPath: scriptPath, StartupTimeout: time.Second}
	sessionDirectory := filepath.Join(root, "session")
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: sessionDirectory}); err != nil {
		t.Fatalf("start browser recorder: %v", err)
	}
	if err := recorder.Stop(); err != nil {
		t.Fatalf("stop browser recorder: %v", err)
	}
	for _, path := range []string{"traces/rrweb.jsonl", "http/frontend.jsonl", "actions/browser.jsonl"} {
		contents, err := os.ReadFile(filepath.Join(sessionDirectory, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var value map[string]any
		if err := json.Unmarshal(contents, &value); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
	}
}

func TestBrowserRecorderTimesOutBeforeReady(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, "setInterval(() => {}, 1000);\n")
	recorder := BrowserRecorder{NodeBinary: "node", ScriptPath: scriptPath, StartupTimeout: 25 * time.Millisecond}
	err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: filepath.Join(root, "session")})
	if err == nil || !strings.Contains(err.Error(), "startup timed out") {
		t.Fatalf("expected startup timeout, got %v", err)
	}
}

func TestBrowserRecorderCapturesProductionScriptAgainstHTTPServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte("<html><body><button id=\"checkout\">Checkout</button></body></html>"))
	}))
	defer server.Close()

	root := t.TempDir()
	recorder := BrowserRecorder{
		NodeBinary:     "node",
		ScriptPath:     productionBrowserScript(t),
		StartupTimeout: 30 * time.Second,
	}
	sessionDirectory := filepath.Join(root, "session")
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: server.URL, SessionDirectory: sessionDirectory}); err != nil {
		t.Fatalf("start production browser recorder: %v", err)
	}
	waitForNonEmptyFile(t, filepath.Join(sessionDirectory, "traces", "rrweb.jsonl"), 5*time.Second)
	if err := recorder.Stop(); err != nil {
		t.Fatalf("stop production browser recorder: %v", err)
	}
	assertJSONLFile(t, filepath.Join(sessionDirectory, "traces", "rrweb.jsonl"))
	assertJSONLFile(t, filepath.Join(sessionDirectory, "http", "frontend.jsonl"))
}

func writeBrowserFixture(t *testing.T, directory, contents string) string {
	t.Helper()
	path := filepath.Join(directory, "capture-fixture.cjs")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write browser fixture: %v", err)
	}
	return path
}

func productionBrowserScript(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve browser test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "scripts", "capture-browser.cjs"))
}

func assertJSONLFile(t *testing.T, path string) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatalf("expected JSONL output in %s", path)
	}
	for _, line := range lines {
		var value map[string]any
		if err := json.Unmarshal([]byte(line), &value); err != nil {
			t.Fatalf("decode JSONL %s: %v", path, err)
		}
	}
}

func waitForNonEmptyFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		contents, err := os.ReadFile(path)
		if err == nil && len(strings.TrimSpace(string(contents))) > 0 {
			return
		}
		select {
		case <-timer.C:
			t.Fatalf("timed out waiting for %s", path)
		case <-ticker.C:
		}
	}
}
