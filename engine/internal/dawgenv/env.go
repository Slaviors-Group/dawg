// Package dawgenv manages runtime environment discovery, bundled asset resolution, and health diagnostics for DAWG.
package dawgenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/procutil"
)

// ResourceDir returns the root directory containing bundled DAWG resources.
func ResourceDir() string {
	if env := os.Getenv("DAWG_RESOURCES_DIR"); env != "" {
		if abs, err := filepath.Abs(env); err == nil {
			return abs
		}
		return env
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		// Check for embedded resources directory next to or above executable
		candidates := []string{
			filepath.Join(exeDir, "resources"),
			filepath.Join(exeDir, "..", "resources"),
			exeDir,
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				// Verify if this directory contains a signature DAWG resource (e.g. scripts or schema)
				if _, err := os.Stat(filepath.Join(candidate, "scripts")); err == nil {
					return candidate
				}
				if _, err := os.Stat(filepath.Join(candidate, "schema")); err == nil {
					return candidate
				}
			}
		}
		return exeDir
	}

	return "."
}

// ResolveMitmdump locates the standalone mitmdump executable.
func ResolveMitmdump() string {
	if env := os.Getenv("DAWG_MITMDUMP_PATH"); env != "" {
		return env
	}

	exeName := "mitmdump"
	if runtime.GOOS == "windows" {
		exeName = "mitmdump.exe"
	}

	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "binaries", "mitmdump", exeName),
		filepath.Join(res, "binaries", exeName),
		filepath.Join(res, "mitmdump", exeName),
		filepath.Join(res, exeName),
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "binaries", "mitmdump", exeName),
			filepath.Join(exeDir, "mitmdump", exeName),
			filepath.Join(exeDir, exeName),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	if path, err := exec.LookPath(exeName); err == nil {
		return path
	}

	return exeName
}

// ResolveNode locates the portable or system Node.js binary.
func ResolveNode() string {
	if env := os.Getenv("DAWG_NODE_PATH"); env != "" {
		return env
	}

	exeName := "node"
	if runtime.GOOS == "windows" {
		exeName = "node.exe"
	}

	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "binaries", "node", exeName),
		filepath.Join(res, "binaries", exeName),
		filepath.Join(res, "node", exeName),
		filepath.Join(res, exeName),
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "binaries", "node", exeName),
			filepath.Join(exeDir, "node", exeName),
			filepath.Join(exeDir, exeName),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	if path, err := exec.LookPath(exeName); err == nil {
		return path
	}

	return exeName
}

// ResolveScript locates an engine script (e.g. capture-browser.cjs, capture-proxy.py, replay-browser.cjs).
func ResolveScript(name string) string {
	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "scripts", name),
		filepath.Join(res, "..", "scripts", name),
		filepath.Join("scripts", name),
		filepath.Join("engine", "scripts", name),
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "scripts", name),
			filepath.Join(exeDir, "..", "scripts", name),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return filepath.Join("scripts", name)
}

// ResolvePolicy locates default.rego.
func ResolvePolicy() string {
	if env := os.Getenv("DAWG_POLICY_PATH"); env != "" {
		return env
	}

	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "schema", "policies", "default.rego"),
		filepath.Join(res, "..", "schema", "policies", "default.rego"),
		filepath.Join("schema", "policies", "default.rego"),
		filepath.Join("..", "schema", "policies", "default.rego"),
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "schema", "policies", "default.rego"),
			filepath.Join(exeDir, "..", "schema", "policies", "default.rego"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return "schema/policies/default.rego"
}

// ResolveSchema locates the manifest schema JSON.
func ResolveSchema(version string) (string, error) {
	if env := os.Getenv("DAWG_SCHEMA_PATH"); env != "" {
		return env, nil
	}

	filename := fmt.Sprintf("v%s.json", version)
	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "schema", "manifest", filename),
		filepath.Join(res, "..", "schema", "manifest", filename),
		filepath.Join("schema", "manifest", filename),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	// Ancestor search from current directory and executable directory
	starts := []string{}
	if wd, err := os.Getwd(); err == nil {
		starts = append(starts, wd)
	}
	if exe, err := os.Executable(); err == nil {
		starts = append(starts, filepath.Dir(exe))
	}

	for _, start := range starts {
		for dir := start; ; dir = filepath.Dir(dir) {
			target := filepath.Join(dir, "schema", "manifest", filename)
			if _, err := os.Stat(target); err == nil {
				return target, nil
			}
			if parent := filepath.Dir(dir); parent == dir {
				break
			}
		}
	}

	return "", fmt.Errorf("dawgenv: locate schema for version %s: %w", version, os.ErrNotExist)
}

// ResolveBrowsersDir returns custom Playwright browser storage path if provisioned in bundle.
func ResolveBrowsersDir() string {
	if env := os.Getenv("PLAYWRIGHT_BROWSERS_PATH"); env != "" {
		return env
	}

	res := ResourceDir()
	candidates := []string{
		filepath.Join(res, "browsers"),
		filepath.Join(res, "node_modules", "playwright-core", ".local-browsers"),
		filepath.Join(res, "node_modules", "playwright", ".local-browsers"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}

	return ""
}

// ComponentStatus captures diagnostics for a single runtime dependency.
type ComponentStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
	Bundled   bool   `json:"bundled"`
	Error     string `json:"error,omitempty"`
}

// DoctorReport contains the full system diagnostic result.
type DoctorReport struct {
	Status      string            `json:"status"` // "ready" or "degraded"
	EnginePath  string            `json:"enginePath"`
	ResourceDir string            `json:"resourceDir"`
	IsBundled   bool              `json:"isBundled"`
	Components  []ComponentStatus `json:"components"`
	GeneratedAt time.Time         `json:"generatedAt"`
}

// RunDoctor inspects all dependencies and returns a diagnostic report.
func RunDoctor(ctx context.Context, engineVersion string) DoctorReport {
	resDir := ResourceDir()
	exe, _ := os.Executable()

	report := DoctorReport{
		Status:      "ready",
		EnginePath:  exe,
		ResourceDir: resDir,
		GeneratedAt: time.Now().UTC(),
	}

	// 1. Engine
	report.Components = append(report.Components, ComponentStatus{
		Name:      "dawg-engine",
		Installed: true,
		Path:      exe,
		Version:   engineVersion,
		Bundled:   false,
	})

	// 2. mitmdump
	mitmPath := ResolveMitmdump()
	mitmStatus := ComponentStatus{Name: "mitmdump", Path: mitmPath}
	if strings.HasPrefix(mitmPath, resDir) {
		mitmStatus.Bundled = true
	}
	cmdCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	mitmCmd := exec.CommandContext(cmdCtx, mitmPath, "--version")
	procutil.HideWindow(mitmCmd)
	if out, err := mitmCmd.Output(); err == nil {
		mitmStatus.Installed = true
		lines := strings.Split(string(out), "\n")
		if len(lines) > 0 {
			mitmStatus.Version = strings.TrimSpace(lines[0])
		}
	} else {
		mitmStatus.Installed = false
		mitmStatus.Error = fmt.Sprintf("failed to execute mitmdump (%s): %v", mitmPath, err)
		report.Status = "degraded"
	}
	report.Components = append(report.Components, mitmStatus)

	// 3. Node.js
	nodePath := ResolveNode()
	nodeStatus := ComponentStatus{Name: "node", Path: nodePath}
	if strings.HasPrefix(nodePath, resDir) {
		nodeStatus.Bundled = true
	}
	cmdCtx2, cancel2 := context.WithTimeout(ctx, 15*time.Second)
	defer cancel2()
	nodeCmd := exec.CommandContext(cmdCtx2, nodePath, "--version")
	procutil.HideWindow(nodeCmd)
	if out, err := nodeCmd.Output(); err == nil {
		nodeStatus.Installed = true
		nodeStatus.Version = strings.TrimSpace(string(out))
	} else {
		nodeStatus.Installed = false
		nodeStatus.Error = fmt.Sprintf("failed to execute node (%s): %v", nodePath, err)
		report.Status = "degraded"
	}
	report.Components = append(report.Components, nodeStatus)

	// 4. Scripts
	scripts := []string{"capture-proxy.py", "replay-browser.cjs"}
	for _, scriptName := range scripts {
		scriptPath := ResolveScript(scriptName)
		scriptStatus := ComponentStatus{Name: "script:" + scriptName, Path: scriptPath}
		if strings.HasPrefix(scriptPath, resDir) {
			scriptStatus.Bundled = true
		}
		if _, err := os.Stat(scriptPath); err == nil {
			scriptStatus.Installed = true
		} else {
			scriptStatus.Installed = false
			scriptStatus.Error = fmt.Sprintf("script not found at %s", scriptPath)
			report.Status = "degraded"
		}
		report.Components = append(report.Components, scriptStatus)
	}

	// 5. Schema & Policies
	policyPath := ResolvePolicy()
	policyStatus := ComponentStatus{Name: "policy:default.rego", Path: policyPath}
	if _, err := os.Stat(policyPath); err == nil {
		policyStatus.Installed = true
	} else {
		policyStatus.Installed = false
		policyStatus.Error = fmt.Sprintf("policy not found at %s", policyPath)
		report.Status = "degraded"
	}
	report.Components = append(report.Components, policyStatus)

	if _, err := ResolveSchema("0.1.4-alpha"); err == nil {
		report.Components = append(report.Components, ComponentStatus{
			Name:      "schema:manifest",
			Installed: true,
			Version:   "0.1.4-alpha",
		})
	} else {
		report.Components = append(report.Components, ComponentStatus{
			Name:      "schema:manifest",
			Installed: false,
			Error:     "manifest schema v0.1.4-alpha not found",
		})
		report.Status = "degraded"
	}

	// Check if environment as a whole is running in bundled mode
	report.IsBundled = mitmStatus.Bundled && nodeStatus.Bundled

	return report
}
