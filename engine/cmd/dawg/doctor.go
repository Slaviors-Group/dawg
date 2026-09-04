package main

import (
	"encoding/json"
	"fmt"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgenv"
	"github.com/spf13/cobra"
)

func newDoctorCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose runtime dependencies and bundled components",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			report := dawgenv.RunDoctor(command.Context(), version)

			if outputFormat(command) == "json" {
				contents, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("doctor: serialize report: %w", err)
				}
				fmt.Fprintln(command.OutOrStdout(), string(contents))
				return nil
			}

			// Human-readable console output
			fmt.Printf("DAWG Engine Doctor Diagnostics (v%s)\n", version)
			fmt.Printf("Resource Root: %s (Bundled: %v)\n\n", report.ResourceDir, report.IsBundled)

			for _, c := range report.Components {
				statusIcon := "✅"
				if !c.Installed {
					statusIcon = "❌"
				}
				details := ""
				if c.Version != "" {
					details = fmt.Sprintf(" (v%s)", c.Version)
				}
				if c.Bundled {
					details += " [bundled]"
				}
				if c.Path != "" {
					details += fmt.Sprintf(" -> %s", c.Path)
				}
				if c.Error != "" {
					details += fmt.Sprintf(" [Error: %s]", c.Error)
				}
				fmt.Printf("%s %-20s%s\n", statusIcon, c.Name, details)
			}

			fmt.Println()
			if report.Status == "ready" {
				fmt.Println("All required runtime components are healthy and ready! 🚀")
			} else {
				fmt.Println("⚠️  Some components are missing or degraded. Check details above.")
			}

			return nil
		},
	}
}
