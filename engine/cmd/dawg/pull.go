package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Slaviors-Group/dawg/engine/internal/registry"
	"github.com/spf13/cobra"
)

type pullResult struct {
	Status   string `json:"status"`
	Output   string `json:"output"`
	Registry string `json:"registry"`
}

func newPullCommand() *cobra.Command {
	var outputDirectory string
	command := &cobra.Command{
		Use:   "pull <registry-reference>",
		Short: "Pull a DAWG artifact from an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			ref := arguments[0]
			if outputDirectory == "" {
				return fmt.Errorf("pull: --output directory is required")
			}
			
			dest, err := filepath.Abs(outputDirectory)
			if err != nil {
				return fmt.Errorf("pull: resolve output directory: %w", err)
			}
			
			if err := (registry.Client{}).Pull(context.Background(), ref, dest); err != nil {
				return err
			}
			return writePullResult(command.OutOrStdout(), outputFormat(command), pullResult{Status: "pulled", Output: dest, Registry: ref})
		},
	}
	command.Flags().StringVar(&outputDirectory, "output", "", "Output directory to extract the artifact")
	return command
}

func writePullResult(writer io.Writer, format string, result pullResult) error {
	if format == "json" {
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			return fmt.Errorf("pull: write JSON result: %w", err)
		}
		return nil
	}
	_, err := fmt.Fprintf(writer, "Pulled %s to %s\n", result.Registry, result.Output)
	return err
}
