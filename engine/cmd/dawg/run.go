package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/Slaviors-Group/dawg/engine/internal/packager"
	"github.com/Slaviors-Group/dawg/engine/internal/replay"
	"github.com/spf13/cobra"
)

type ReplayOutput struct {
	Status         string `json:"status"`
	Outcome        string `json:"outcome"`
	ScreenshotPath string `json:"screenshotPath"`
}

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

			if err := packager.Unpack(layoutDir, tmpDir); err != nil {
				return fmt.Errorf("replay: unpack artifact: %w", err)
			}

			manifestPath := filepath.Join(layoutDir, "dawg-manifest.json")
			m, err := manifest.Read(manifestPath)
			if err != nil {
				return fmt.Errorf("replay: read manifest: %w", err)
			}

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
			
			// Docker compose project name must be lowercase alphanumeric
			projectName := "dawg-" + strings.ReplaceAll(m.ID[7:15], ":", "") 
			composeFile := filepath.Join(tmpDir, "env", "compose.yaml")
			
			// Start sandbox only if compose file exists (which it should in MVP)
			if _, err := os.Stat(composeFile); err == nil {
				if err := sandbox.Start(ctx, composeFile, projectName); err != nil {
					return err
				}
				defer sandbox.Teardown(context.Background(), composeFile, projectName)
			}

			// DB Restore
			containerName := projectName + "-db-1"
			fixturePath := filepath.Join(tmpDir, "db", "fixture.sql")
			if _, err := os.Stat(fixturePath); err == nil {
				dbRestorer := replay.NewDBRestorer(sandbox.Runner, "postgres") // default PG db
				if err := dbRestorer.Restore(ctx, containerName, fixturePath); err != nil {
					return err
				}
			}

			// Cassette Replayer
			cassetteFile := filepath.Join(tmpDir, "cassettes", "cassette.yaml") // mitmproxy format
			if _, err := os.Stat(cassetteFile); err == nil {
				replayer := replay.NewCassetteReplayer(sandbox.Runner)
				if err := replayer.Start(ctx, 8080, cassetteFile); err != nil {
					return err
				}
				defer replayer.Stop()
			}

			// Event Player
			executable, _ := os.Executable()
			player := &replay.EventPlayer{
				ScriptPath: filepath.Join(filepath.Dir(executable), "..", "scripts", "replay-browser.cjs"),
			}
			
			// Fallback for dev mode
			if _, err := os.Stat(player.ScriptPath); err != nil {
				player.ScriptPath = filepath.Join("scripts", "replay-browser.cjs")
			}

			outcome, err := player.Replay(ctx, tmpDir)
			if err != nil {
				return err
			}

			out := ReplayOutput{
				Status:         "success",
				Outcome:        "reproduced", // For MVP, we assume reproduction if script finished without error
				ScreenshotPath: outcome.ScreenshotPath,
			}

			if outputFormat == "json" {
				b, _ := json.Marshal(out)
				fmt.Fprintln(os.Stdout, string(b))
			} else {
				fmt.Printf("Replay completed successfully.\nFinal screenshot: %s\n", outcome.ScreenshotPath)
			}

			return nil
		},
	}
	command.Flags().StringVar(&outputFormat, "output", "text", "Output format (text, json)")
	return command
}
