// Command dawg is the CLI entry point for DAWG's capture and replay pipeline.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

const version = "0.1.0-dev"

func main() {
	rootCommand := newRootCommand()
	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		pauseIfStandalone()
		os.Exit(2)
	}

	if len(os.Args) == 1 {
		pauseIfStandalone()
	}
}

func pauseIfStandalone() {
	if runtime.GOOS == "windows" && len(os.Args) <= 1 {
		fmt.Println("\nPress Enter to exit...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}

func newRootCommand() *cobra.Command {
	var output string

	command := &cobra.Command{
		Use:           "dawg",
		Short:         "Package reproducible web-app bug reports",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	command.PersistentFlags().StringVar(&output, "output", "text", "Output format: text or json")
	command.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		if output != "text" && output != "json" {
			return fmt.Errorf("invalid output format %q: must be text or json", output)
		}
		return nil
	}
	command.Version = version
	command.SetVersionTemplate("{{.Version}}\n")
	command.RunE = func(command *cobra.Command, _ []string) error {
		if output == "json" {
			return json.NewEncoder(command.OutOrStdout()).Encode(map[string]string{"status": "ready", "version": version})
		}
		return command.Help()
	}
	command.AddCommand(newInitCommand())
	command.AddCommand(newCaptureCommand())
	command.AddCommand(newPushCommand())
	command.AddCommand(newInspectCommand())
	command.AddCommand(newRunCommand())
	command.AddCommand(newVerifyCommand())
	
	return command
}

func outputFormat(command *cobra.Command) string {
	format, err := command.Flags().GetString("output")
	if err != nil {
		return "text"
	}
	return format
}
