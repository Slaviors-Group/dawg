package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/spf13/cobra"
)

const configFileName = "dawg.config.yaml"

const defaultConfig = `schemaVersion: "0.1.0"
capture:
  outputDir: ".dawg/captures"
  browser: "chromium"
sanitize:
  policyFile: ".dawg/policies/default.rego"
replay:
  composeFile: "compose.yaml"
`

type initResult struct {
	Status      string `json:"status"`
	ConfigPath  string `json:"configPath"`
	Overwritten bool   `json:"overwritten"`
}

func newInitCommand() *cobra.Command {
	var force bool

	command := &cobra.Command{
		Use:   "init [directory]",
		Short: "Create a DAWG configuration file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(command *cobra.Command, arguments []string) error {
			directory := "."
			if len(arguments) == 1 {
				directory = arguments[0]
			}
			result, err := initializeConfig(directory, force)
			if err != nil {
				return err
			}
			return writeInitResult(command.OutOrStdout(), outputFormat(command), result)
		},
	}
	command.Flags().BoolVar(&force, "force", false, "Overwrite an existing configuration file")
	return command
}

func initializeConfig(directory string, force bool) (initResult, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return initResult{}, fmt.Errorf("init: inspect target directory %s: %w", directory, err)
	}
	if !info.IsDir() {
		return initResult{}, fmt.Errorf("init: target %s is not a directory", directory)
	}

	configPath := filepath.Join(directory, configFileName)
	if !force {
		file, err := os.OpenFile(configPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return initResult{}, fmt.Errorf("init: %w: %s; use --force to overwrite", dawgtypes.ErrConfigExists, configPath)
			}
			return initResult{}, fmt.Errorf("init: create config %s: %w", configPath, err)
		}
		defer file.Close()
		if _, err := io.WriteString(file, defaultConfig); err != nil {
			return initResult{}, fmt.Errorf("init: write config %s: %w", configPath, err)
		}
		return initResult{Status: "created", ConfigPath: configPath}, nil
	}

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0o600); err != nil {
		return initResult{}, fmt.Errorf("init: overwrite config %s: %w", configPath, err)
	}
	return initResult{Status: "created", ConfigPath: configPath, Overwritten: true}, nil
}

func writeInitResult(writer io.Writer, format string, result initResult) error {
	if format == "json" {
		if err := json.NewEncoder(writer).Encode(result); err != nil {
			return fmt.Errorf("init: write JSON result: %w", err)
		}
		return nil
	}
	_, err := fmt.Fprintf(writer, "Created %s\n", result.ConfigPath)
	return err
}
