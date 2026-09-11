package capture

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBrowserRecorderDrainsBeforeGracefulStop(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, `
const fs = require("node:fs");
const readline = require("node:readline");
const options = Object.fromEntries(process.argv.slice(2).reduce((pairs, value, index, values) => index % 2 === 0 ? [...pairs, [value.slice(2), values[index + 1]]] : pairs, []));
for (const key of ["rrweb-output", "http-output", "actions-output"]) fs.writeFileSync(options[key], "{\"initial\":true}\n");
const input = readline.createInterface({ input: process.stdin });
input.on("line", line => {
    if (JSON.parse(line).command !== "stop") return;
    setTimeout(() => {
        fs.appendFileSync(options["rrweb-output"], "{\"flushed\":true}\n");
        console.log(JSON.stringify({ status: "stopped" }));
        input.close();
    }, 30);
});
console.log(JSON.stringify({ status: "capturing" }));
`)
	recorder := BrowserRecorder{
		NodeBinary:     "node",
		ScriptPath:     scriptPath,
		StartupTimeout: time.Second,
		StopTimeout:    time.Second,
	}
	sessionDirectory := filepath.Join(root, "session")
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: sessionDirectory}); err != nil {
		t.Fatalf("start browser recorder: %v", err)
	}
	if err := recorder.Stop(); err != nil {
		t.Fatalf("stop browser recorder: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(sessionDirectory, "traces", "rrweb.jsonl"))
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	if !strings.Contains(string(contents), `{"flushed":true}`) {
		t.Fatalf("graceful stop returned before final flush: %s", contents)
	}
	for _, relativePath := range []string{"traces/rrweb.jsonl", "http/frontend.jsonl", "actions/browser.jsonl"} {
		assertBrowserJSONLFile(t, filepath.Join(sessionDirectory, filepath.FromSlash(relativePath)))
	}
	if err := recorder.Stop(); err != nil {
		t.Fatalf("idempotent stop: %v", err)
	}
}

func TestBrowserRecorderPassesBundledBrowserFallbackOptions(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, `
const fs = require("node:fs");
const readline = require("node:readline");
const values = {};
for (let index = 2; index < process.argv.length; index += 2) values[process.argv[index].slice(2)] = process.argv[index + 1];
fs.writeFileSync(values["actions-output"], JSON.stringify({
    endpoint: values["cdp-endpoint"],
    chromium: values["chromium-path"],
    profile: values["browser-profile"],
    proxy: values["proxy-server"],
}));
const input = readline.createInterface({ input: process.stdin });
input.on("line", () => {
    input.close();
    process.stdin.destroy();
});
console.log(JSON.stringify({ status: "capturing" }));
`)
	recorder := BrowserRecorder{
		NodeBinary:              "node",
		ScriptPath:              scriptPath,
		CDPEndpoint:             "http://127.0.0.1:9333",
		ChromiumPath:            "/opt/dawg/chromium",
		BrowserProfileDirectory: filepath.Join(root, "profile"),
		ProxyServer:             "http://127.0.0.1:18881",
		StartupTimeout:          time.Second,
		StopTimeout:             time.Second,
	}
	sessionDirectory := filepath.Join(root, "session")
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: sessionDirectory}); err != nil {
		t.Fatalf("start browser recorder: %v", err)
	}
	if err := recorder.Stop(); err != nil {
		t.Fatalf("stop browser recorder: %v", err)
	}

	contents, err := os.ReadFile(filepath.Join(sessionDirectory, "actions", "browser.jsonl"))
	if err != nil {
		t.Fatalf("read fallback options: %v", err)
	}
	var options map[string]string
	if err := json.Unmarshal(contents, &options); err != nil {
		t.Fatalf("decode fallback options: %v", err)
	}
	if options["endpoint"] != recorder.CDPEndpoint || options["chromium"] != recorder.ChromiumPath || options["profile"] != recorder.BrowserProfileDirectory || options["proxy"] != recorder.ProxyServer {
		t.Fatalf("unexpected fallback options: %#v", options)
	}
}

func TestBrowserRecorderStopIsSafeForConcurrentCallers(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, `
const readline = require("node:readline");
const input = readline.createInterface({ input: process.stdin });
input.on("line", line => {
    if (JSON.parse(line).command === "stop") setTimeout(() => input.close(), 20);
});
console.log(JSON.stringify({ status: "capturing" }));
`)
	recorder := BrowserRecorder{NodeBinary: "node", ScriptPath: scriptPath, StartupTimeout: time.Second, StopTimeout: time.Second}
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: filepath.Join(root, "session")}); err != nil {
		t.Fatalf("start browser recorder: %v", err)
	}

	var waitGroup sync.WaitGroup
	errors := make(chan error, 4)
	for range 4 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errors <- recorder.Stop()
		}()
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent stop: %v", err)
		}
	}
}

func TestBrowserRecorderTimesOutBeforeReady(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, "setInterval(() => {}, 1000);\n")
	recorder := BrowserRecorder{
		NodeBinary:     "node",
		ScriptPath:     scriptPath,
		StartupTimeout: 25 * time.Millisecond,
		StopTimeout:    25 * time.Millisecond,
	}
	err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: filepath.Join(root, "session")})
	if err == nil || !strings.Contains(err.Error(), "startup timed out") {
		t.Fatalf("expected startup timeout, got %v", err)
	}
}

func TestBrowserRecorderUnexpectedExitRequestsSessionStop(t *testing.T) {
	root := t.TempDir()
	scriptPath := writeBrowserFixture(t, root, `
console.log(JSON.stringify({ status: "capturing" }));
setTimeout(() => process.exit(3), 30);
`)
	recorder := BrowserRecorder{NodeBinary: "node", ScriptPath: scriptPath, StartupTimeout: time.Second, StopTimeout: time.Second}
	if err := recorder.Start(context.Background(), BrowserCaptureRequest{TargetURL: "http://127.0.0.1:1", SessionDirectory: filepath.Join(root, "session")}); err != nil {
		t.Fatalf("start browser recorder: %v", err)
	}

	select {
	case <-recorder.StopRequested():
	case <-time.After(time.Second):
		t.Fatal("unexpected recorder exit did not request session stop")
	}
	if err := recorder.Stop(); err == nil {
		t.Fatal("expected unexpected recorder exit to fail capture")
	}
}

func writeBrowserFixture(t *testing.T, directory, contents string) string {
	t.Helper()
	path := filepath.Join(directory, "capture-fixture.cjs")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write browser fixture: %v", err)
	}
	return path
}

func assertBrowserJSONLFile(t *testing.T, path string) {
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
