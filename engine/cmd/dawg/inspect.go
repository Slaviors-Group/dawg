package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
	"github.com/spf13/cobra"
)

func newInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <artifact-directory>",
		Short: "Inspect a DAWG artifact without replaying it",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			value, err := inspectArtifact(arguments[0])
			if err != nil {
				return err
			}
			return writeInspectedManifest(command.OutOrStdout(), outputFormat(command), value)
		},
	}
}

func inspectArtifact(directory string) (manifest.Manifest, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("inspect: artifact directory %s: %w", directory, err)
	}
	if !info.IsDir() {
		return manifest.Manifest{}, fmt.Errorf("inspect: artifact path %s is not a directory", directory)
	}
	value, err := manifest.Read(filepath.Join(directory, "dawg-manifest.json"))
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("inspect: read artifact manifest: %w", err)
	}
	schemaPath, err := manifest.DefaultSchemaPath()
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("inspect: locate manifest schema: %w", err)
	}
	if err := manifest.Validate(schemaPath, value); err != nil {
		return manifest.Manifest{}, fmt.Errorf("inspect: validate artifact manifest: %w", err)
	}
	return value, nil
}

func writeInspectedManifest(writer io.Writer, _ string, value manifest.Manifest) error {
	contents, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("inspect: serialize artifact manifest: %w", err)
	}
	if _, err := fmt.Fprintln(writer, string(contents)); err != nil {
		return fmt.Errorf("inspect: write artifact manifest: %w", err)
	}
	return nil
}
