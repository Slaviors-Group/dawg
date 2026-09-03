package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/Slaviors-Group/dawg/engine/internal/verify"
	"github.com/spf13/cobra"
)

func newVerifyCommand() *cobra.Command {
	var against string
	var outputFormat string

	command := &cobra.Command{
		Use:   "verify <artifact-directory>",
		Short: "Verify that an artifact reproduces against the local environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			layoutDir := arguments[0]
			ctx := command.Context()

			tmpDir, err := os.MkdirTemp("", "dawg-verify-*")
			if err != nil {
				return fmt.Errorf("verify: create temp directory: %w", err)
			}
			defer os.RemoveAll(tmpDir)

			// Execute Replay Pipeline
			out, err := ExecuteReplay(ctx, layoutDir, tmpDir)
			if err != nil {
				return fmt.Errorf("verify: replay failed: %w", err)
			}

			// Run Verification Checks
			manifestPath := filepath.Join(layoutDir, "dawg-manifest.json")
			m, err := manifest.Read(manifestPath)
			if err != nil {
				return fmt.Errorf("verify: read manifest: %w", err)
			}

			v := &verify.Verifier{}
			result, err := v.Verify(ctx, out, tmpDir, m.ExpectedOutcome)
			if err != nil {
				return fmt.Errorf("verify: run checks: %w", err)
			}

			// If user specifies a branch, record it
			if against != "" {
				result.VerifiedAgainst = against
			}

			// Output
			if outputFormat == "json" {
				b, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(os.Stdout, string(b))
			} else {
				fmt.Printf("\n--- Verification Report ---\n")
				fmt.Printf("Artifact ID: %s\n", result.ArtifactID)
				fmt.Printf("Verified Against: %s\n", result.VerifiedAgainst)
				fmt.Printf("Result: %s\n", result.Result)
				fmt.Printf("Summary: %s\n\n", result.Summary)
				fmt.Printf("Checks:\n")
				for _, check := range result.Checks {
					status := "FAIL"
					if check.Passed {
						status = "PASS"
					}
					fmt.Printf("  [%s] %s\n", status, check.Type)
					fmt.Printf("    Expected: %v\n", check.Expected)
					fmt.Printf("    Actual:   %v\n", check.Actual)
					if !check.Passed {
						// Only show threshold details if it failed and applies
						if check.Threshold > 0 {
							fmt.Printf("    Diff Pixels: %d (Threshold: %d)\n", check.DiffPixels, check.Threshold)
						}
					}
				}
			}

			if result.Result == "fail" {
				os.Exit(1)
			}

			return nil
		},
	}
	
	command.Flags().StringVar(&against, "against", "local", "Branch or commit being verified against")
	command.Flags().StringVar(&outputFormat, "output", "text", "Output format (text, json)")
	return command
}
