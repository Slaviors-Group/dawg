package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"time"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgenv"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/Slaviors-Group/dawg/engine/internal/packager"
	"github.com/Slaviors-Group/dawg/engine/internal/replay"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	var outputFormat string
	command := &cobra.Command{
		Use:   "run <artifact-directory>",
		Short: "Replay a captured DAWG artifact",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			layoutDir := arguments[0]
			ctx := command.Context()

			tmpDir, err := os.MkdirTemp("", "dawg-replay-*")
			if err != nil {
				return fmt.Errorf("replay: create temp directory: %w", err)
			}
			// Retain tmpDir for inspection if needed, but PRD implies ephemeral. We'll clean it up.
			defer os.RemoveAll(tmpDir)

			out, err := ExecuteReplay(ctx, layoutDir, tmpDir)
			if err != nil {
				return err
			}

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(out, "", "  ")
				fmt.Fprintln(os.Stdout, string(b))
			} else {
				fmt.Printf("Replay completed successfully.\nFinal screenshot: %s\n", out.Outcomes.Screenshots[0])
			}

			return nil
		},
	}
	command.Flags().StringVar(&outputFormat, "output", "text", "Output format (text, json)")
	return command
}

// ExecuteReplay runs the replay pipeline for an artifact.
func ExecuteReplay(ctx context.Context, layoutDir, tmpDir string) (dawgtypes.ReplayOutput, error) {
	var out dawgtypes.ReplayOutput

	if err := packager.Unpack(layoutDir, tmpDir); err != nil {
		return out, fmt.Errorf("replay: unpack artifact: %w", err)
	}

	manifestPath := filepath.Join(layoutDir, "dawg-manifest.json")
	m, err := manifest.Read(manifestPath)
	if err != nil {
		return out, fmt.Errorf("replay: read manifest: %w", err)
	}

	out.ArtifactID = m.ID
	out.ReplayedAt = time.Now()
	out.Status = "started"

	// Determinism
	envVars := replay.ConfigureDeterminism(m.Determinism)
	for _, ev := range envVars {
		parts := strings.SplitN(ev, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}

	sandbox := replay.Sandbox{
		Runner: replay.ExecRunner{},
	}

	// projectName is derived from the immutable artifact ID so two concurrent
	// replays of different artifacts never collide on the same Docker project.
	projectName := "dawg-" + strings.ReplaceAll(m.ID[7:15], ":", "")
	composeFile := filepath.Join(tmpDir, "env", "compose.yaml")

	out.Sandbox.ComposeProject = projectName

	sandboxStarted := false
	if _, err := os.Stat(composeFile); err == nil {
		if err := sandbox.Start(ctx, composeFile, projectName); err != nil {
			return out, err
		}
		sandboxStarted = true
	}

	containerName := projectName + "-db-1"
	out.Sandbox.ContainerID = containerName
	fixturePath := filepath.Join(tmpDir, "db", "fixture.sql")
	if runtime.GOOS != "windows" {
		if _, err := os.Stat(fixturePath); err == nil {
			dbRestorer := replay.NewDBRestorer(sandbox.Runner, "postgres")
			if err := dbRestorer.Restore(ctx, containerName, fixturePath); err != nil {
				if sandboxStarted {
					_ = sandbox.Teardown(context.Background(), composeFile, projectName)
				}
				return out, err
			}
		}
	}

	cassetteFile := filepath.Join(tmpDir, "cassettes", "cassette.yaml")
	var replayer *replay.CassetteReplayer
	if _, err := os.Stat(cassetteFile); err == nil {
		replayer = replay.NewCassetteReplayer(sandbox.Runner)
		replayer.Executable = dawgenv.ResolveMitmdump()
		if err := replayer.Start(ctx, 8080, cassetteFile); err != nil {
			if sandboxStarted {
				_ = sandbox.Teardown(context.Background(), composeFile, projectName)
			}
			return out, err
		}
	}

	player := &replay.EventPlayer{
		NodeBinary:   dawgenv.ResolveNode(),
		ScriptPath:   dawgenv.ResolveScript("replay-browser.cjs"),
		BrowsersDir:  dawgenv.ResolveBrowsersDir(),
		ChromiumPath: dawgenv.ResolveChromiumExecutable(),
	}

	outcome, replayErr := player.Replay(ctx, tmpDir)

	// Explicit sequential cleanup: stop the cassette replayer and tear down
	// the Docker sandbox BEFORE the caller removes tmpDir. Stacked defers run
	// in LIFO order, but because the caller defers os.RemoveAll(tmpDir) before
	// calling us, relying on defers here would let Docker / mitmdump hold file
	// handles into an already-deleted directory — corrupting the project state
	// and causing "start replay error" on the next replay attempt.
	if replayer != nil {
		_ = replayer.Stop()
	}
	if sandboxStarted {
		_ = sandbox.Teardown(context.Background(), composeFile, projectName)
	}

	if replayErr != nil {
		return out, replayErr
	}

	out.Status = "completed"
	out.Outcomes.ExitCode = 0 // Assuming success if it reached here
	out.Outcomes.AppLogs = outcome.Output
	if outcome.ScreenshotPath != "" {
		out.Outcomes.Screenshots = []string{outcome.ScreenshotPath}
	}

	// Check if frontend HTTP responses were captured
	httpResponsesPath := filepath.Join(tmpDir, "http", "frontend.jsonl")
	if _, err := os.Stat(httpResponsesPath); err == nil {
		out.Outcomes.HTTPResponses = httpResponsesPath
	}

	return out, nil
}
