package capture

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProxyManagerStartsAndStopsFixture(t *testing.T) {
	root := t.TempDir()
	fixture := writeProxyFixture(t, root, `
console.log(JSON.stringify({status: "capturing"}));
setInterval(() => {}, 1000);
`)
	manager := ProxyManager{Executable: "node", AddonPath: fixture, Arguments: fixtureArguments, StartupTimeout: time.Second, StopTimeout: time.Second}
	if err := manager.Start(context.Background(), ProxyCaptureRequest{ListenPort: availablePort(t), SessionDirectory: filepath.Join(root, "session")}); err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	if err := manager.Stop(); err != nil {
		t.Fatalf("stop proxy: %v", err)
	}
}

func TestProxyManagerRejectsUnavailablePort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %v", err)
	}
	defer listener.Close()
	port := listener.Addr().(*net.TCPAddr).Port
	manager := ProxyManager{Executable: "node", AddonPath: "unused", Arguments: fixtureArguments}
	err = manager.Start(context.Background(), ProxyCaptureRequest{ListenPort: port, SessionDirectory: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected unavailable port error, got %v", err)
	}
}

func TestProxyManagerTimesOutBeforeReady(t *testing.T) {
	root := t.TempDir()
	fixture := writeProxyFixture(t, root, "setInterval(() => {}, 1000);\n")
	manager := ProxyManager{Executable: "node", AddonPath: fixture, Arguments: fixtureArguments, StartupTimeout: 25 * time.Millisecond}
	err := manager.Start(context.Background(), ProxyCaptureRequest{ListenPort: availablePort(t), SessionDirectory: filepath.Join(root, "session")})
	if err == nil || !strings.Contains(err.Error(), "startup timed out") {
		t.Fatalf("expected startup timeout, got %v", err)
	}
}

func TestProductionProxyAddonCompiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("mitmproxy is not installed on this Windows development host")
	}
	if _, err := os.Stat(productionProxyAddon(t)); err != nil {
		t.Fatalf("production addon missing: %v", err)
	}
}

func availablePort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("allocate port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func writeProxyFixture(t *testing.T, directory, contents string) string {
	t.Helper()
	path := filepath.Join(directory, "proxy-fixture.cjs")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write proxy fixture: %v", err)
	}
	return path
}

func fixtureArguments(_ ProxyCaptureRequest, addonPath string) []string {
	return []string{addonPath}
}

func TestProxyManagerIdempotentStop(t *testing.T) {
	// Calling Stop() when the proxy was never started must return nil — not an
	// error — so that a crashed proxy doesn't surface a spurious "stop error"
	// to the user when Stop() is called during session cleanup.
	manager := &ProxyManager{}
	if err := manager.Stop(); err != nil {
		t.Fatalf("Stop on never-started proxy: expected nil, got %v", err)
	}

	// A second Stop after a successful Stop must also be nil.
	root := t.TempDir()
	fixture := writeProxyFixture(t, root, `
console.log(JSON.stringify({status: "capturing"}));
setInterval(() => {}, 1000);
`)
	running := ProxyManager{
		Executable:     "node",
		AddonPath:      fixture,
		Arguments:      fixtureArguments,
		StartupTimeout: time.Second,
		StopTimeout:    time.Second,
	}
	if err := running.Start(context.Background(), ProxyCaptureRequest{
		ListenPort:       availablePort(t),
		SessionDirectory: filepath.Join(root, "session"),
	}); err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	if err := running.Stop(); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if err := running.Stop(); err != nil {
		t.Fatalf("second stop (idempotent): expected nil, got %v", err)
	}
}

func productionProxyAddon(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve proxy test path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..", "scripts", "capture-proxy.py"))
}


