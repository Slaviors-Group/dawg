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

type pushResult struct {
	Status   string `json:"status"`
	Artifact string `json:"artifact"`
	Registry string `json:"registry"`
}

func newPushCommand() *cobra.Command {
	var registryReference string
	command := &cobra.Command{
		Use:   "push <artifact-directory>",
		Short: "Push a DAWG artifact to an OCI registry",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			artifact, err := filepath.Abs(arguments[0])
			if err != nil {
				return fmt.Errorf("push: resolve artifact directory: %w", err)
			}
			if registryReference == "" {
				return fmt.Errorf("push: --registry is required")
			}
			if err := (registry.Client{}).Push(context.Background(), artifact, registryReference); err != nil {
				return err
			}
			return writePushResult(command.OutOrStdout(), outputFormat(command), pushResult{Status: "pushed", Artifact: artifact, Registry: registryReference})
		},
	}
	command.Flags().StringVar(&registryReference, "registry", "", "OCI registry reference")
	return command
}

func writePushResult(writer io.Writer, format string, result pushResult) error {
	if format == "json" {
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			return fmt.Errorf("push: write JSON result: %w", err)
		}
		return nil
	}
	_, err := fmt.Fprintf(writer, "Pushed %s to %s\n", result.Artifact, result.Registry)
	return err
}
