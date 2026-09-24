package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/packager"
	"github.com/spf13/cobra"
)

func newDiagnosticsCommand() *cobra.Command {
	command := &cobra.Command{Use: "diagnostics", Short: "Inspect and export sanitized diagnostic evidence"}
	command.AddCommand(newDiagnosticsInspectCommand())
	command.AddCommand(newDiagnosticsHARCommand())
	command.AddCommand(newDiagnosticsCurlCommand())
	command.AddCommand(newDiagnosticsRemoveCommand())
	return command
}

func newDiagnosticsRemoveCommand() *cobra.Command {
	var output string
	var categories []string
	var bodyRefs []string
	command := &cobra.Command{Use: "remove <artifact-directory>", Short: "Create a reviewed copy without selected diagnostic evidence", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		if output == "" {
			return fmt.Errorf("diagnostics remove: --output-dir is required")
		}
		artifact, err := packager.RepackageDiagnostics(packager.DiagnosticRemovalRequest{InputDirectory: args[0], OutputDirectory: output, Categories: categories, BodyRefs: bodyRefs})
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(artifact)
	}}
	command.Flags().StringVar(&output, "output-dir", "", "New artifact directory")
	command.Flags().StringSliceVar(&categories, "remove-category", nil, "Diagnostic category to remove: console, network, errors, device, or bodies")
	command.Flags().StringSliceVar(&bodyRefs, "remove-body-ref", nil, "Retained diagnostic body path to remove, such as diagnostics/bodies/response_request-id.json")
	return command
}

func newDiagnosticsInspectCommand() *cobra.Command {
	return &cobra.Command{Use: "inspect <artifact-directory>", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		evidence, err := packager.ReadDiagnostics(args[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(command.OutOrStdout()).Encode(evidence)
	}}
}

func newDiagnosticsHARCommand() *cobra.Command {
	var output string
	command := &cobra.Command{Use: "export-har <artifact-directory>", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		if output == "" {
			return fmt.Errorf("diagnostics export-har: --output is required")
		}
		evidence, err := packager.ReadDiagnostics(args[0])
		if err != nil {
			return err
		}
		contents, err := json.MarshalIndent(sanitizedHAR(evidence), "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(output, append(contents, '\n'), 0o600); err != nil {
			return fmt.Errorf("diagnostics export-har: write output: %w", err)
		}
		return nil
	}}
	command.Flags().StringVar(&output, "output", "", "Destination HAR file")
	return command
}

func newDiagnosticsCurlCommand() *cobra.Command {
	var requestID string
	command := &cobra.Command{Use: "copy-curl <artifact-directory>", Args: cobra.ExactArgs(1), RunE: func(command *cobra.Command, args []string) error {
		if requestID == "" {
			return fmt.Errorf("diagnostics copy-curl: --request-id is required")
		}
		evidence, err := packager.ReadDiagnostics(args[0])
		if err != nil {
			return err
		}
		for _, record := range evidence.Network {
			if fmt.Sprint(record["requestId"]) == requestID {
				_, err := fmt.Fprintln(command.OutOrStdout(), curlForRecord(record, evidence.Bodies))
				return err
			}
		}
		return fmt.Errorf("diagnostics copy-curl: request %q not found", requestID)
	}}
	command.Flags().StringVar(&requestID, "request-id", "", "Captured request ID")
	return command
}

func sanitizedHAR(evidence packager.DiagnosticEvidence) map[string]any {
	entries := make([]any, 0, len(evidence.Network))
	for _, record := range evidence.Network {
		request, _ := record["request"].(map[string]any)
		response, _ := record["response"].(map[string]any)
		timing, _ := record["timing"].(map[string]any)
		entries = append(entries, map[string]any{
			"startedDateTime": record["timestamp"], "time": timing["durationMs"],
			"request":  map[string]any{"method": record["method"], "url": record["url"], "headers": harHeaders(request["headers"]), "_dawg": bodyAnnotation(request["body"])},
			"response": map[string]any{"status": response["status"], "statusText": response["statusText"], "headers": harHeaders(response["headers"]), "content": map[string]any{"_dawg": bodyAnnotation(response["body"])}},
			"cache":    map[string]any{}, "timings": map[string]any{"wait": timing["durationMs"]},
		})
	}
	return map[string]any{"log": map[string]any{"version": "1.2", "creator": map[string]string{"name": "DAWG", "version": version}, "entries": entries}}
}

func harHeaders(value any) []map[string]string {
	headers, _ := value.(map[string]any)
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]map[string]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, map[string]string{"name": key, "value": fmt.Sprint(headers[key])})
	}
	return result
}

func bodyAnnotation(value any) map[string]any {
	body, _ := value.(map[string]any)
	if body == nil {
		return map[string]any{"state": "unavailable"}
	}
	result := map[string]any{}
	for _, key := range []string{"state", "contentType", "size", "sha256"} {
		if body[key] != nil {
			result[key] = body[key]
		}
	}
	return result
}

func curlForRecord(record map[string]any, bodies map[string]string) string {
	request, _ := record["request"].(map[string]any)
	parts := []string{"curl", "-X", shellQuote(fmt.Sprint(record["method"])), shellQuote(fmt.Sprint(record["url"]))}
	headers, _ := request["headers"].(map[string]any)
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lower := strings.ToLower(key)
		if lower == "authorization" || lower == "proxy-authorization" || lower == "cookie" || lower == "set-cookie" || lower == "x-api-key" || lower == "x-auth-token" || lower == "x-csrf-token" || lower == "x-amz-security-token" {
			continue
		}
		parts = append(parts, "-H", shellQuote(key+": "+fmt.Sprint(headers[key])))
	}
	body, _ := request["body"].(map[string]any)
	if state, _ := body["state"].(string); state == "captured" || state == "redacted" {
		if ref, _ := body["ref"].(string); ref != "" {
			if value, ok := bodies[ref]; ok {
				parts = append(parts, "--data-raw", shellQuote(value))
			}
		}
	}
	return "# Review before running: this request may mutate data.\n" + strings.Join(parts, " ")
}

func shellQuote(value string) string { return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'" }
