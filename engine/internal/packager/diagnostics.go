package packager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func diagnosticSummary(sessionDirectory string, metadata dawgtypes.CaptureMetadata, policyVersion string) (dawgtypes.DiagnosticsSummary, error) {
	summary := metadata.Diagnostics
	if summary.FormatVersion == "" {
		summary = dawgtypes.EmptyDiagnosticsSummary(dawgtypes.DiagnosticProfileSafe, policyVersion)
	}
	if summary.PolicyVersion == "" {
		summary.PolicyVersion = policyVersion
	}
	if summary.Limits.MaxConsoleRecordBytes == 0 {
		summary.Limits = dawgtypes.DefaultDiagnosticLimits()
	}
	profilePath := filepath.Join(sessionDirectory, "diagnostics", "profile.json")
	if contents, err := os.ReadFile(profilePath); err == nil {
		var profile struct {
			Profile dawgtypes.DiagnosticProfile `json:"profile"`
		}
		if json.Unmarshal(contents, &profile) == nil && (profile.Profile == dawgtypes.DiagnosticProfileSafe || profile.Profile == dawgtypes.DiagnosticProfileEnhanced) {
			summary.Profile = profile.Profile
		}
	}
	summary.ConsoleEvents, summary.NetworkEvents, summary.ErrorEvents = 0, 0, 0
	summary.RetainedRequestBodies, summary.RetainedResponseBodies = 0, 0
	summary.RedactedBodies, summary.BlockedBodies, summary.TruncatedBodies = 0, 0, 0
	sources := map[string]struct{}{}
	for _, relative := range []struct {
		path string
		kind string
	}{
		{"diagnostics/console.jsonl", "console"},
		{"diagnostics/network.jsonl", "network"},
		{"diagnostics/errors.jsonl", "error"},
	} {
		path := filepath.Join(sessionDirectory, filepath.FromSlash(relative.path))
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return dawgtypes.DiagnosticsSummary{}, fmt.Errorf("packager: open %s: %w", relative.path, err)
		}
		scanner := bufio.NewScanner(file)
		recordLimit := dawgtypes.MaxJSONLRecordBytes
		if relative.kind == "console" {
			recordLimit = int(summary.Limits.MaxConsoleRecordBytes)
		}
		scanner.Buffer(make([]byte, 64*1024), recordLimit)
		for scanner.Scan() {
			var record map[string]any
			if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
				_ = file.Close()
				return dawgtypes.DiagnosticsSummary{}, fmt.Errorf("packager: decode %s: %w", relative.path, err)
			}
			switch relative.kind {
			case "console":
				summary.ConsoleEvents++
			case "network":
				summary.NetworkEvents++
				countBodyStates(record, &summary)
			case "error":
				summary.ErrorEvents++
			}
			if source, ok := record["source"].(string); ok && source != "" {
				sources[source] = struct{}{}
			}
		}
		if err := scanner.Err(); err != nil {
			_ = file.Close()
			return dawgtypes.DiagnosticsSummary{}, fmt.Errorf("packager: scan %s: %w", relative.path, err)
		}
		if err := file.Close(); err != nil {
			return dawgtypes.DiagnosticsSummary{}, fmt.Errorf("packager: close %s: %w", relative.path, err)
		}
	}
	summary.Sources = summary.Sources[:0]
	for source := range sources {
		summary.Sources = append(summary.Sources, source)
	}
	sort.Strings(summary.Sources)
	return summary, nil
}

func countBodyStates(record map[string]any, summary *dawgtypes.DiagnosticsSummary) {
	for _, side := range []string{"request", "response"} {
		container, _ := record[side].(map[string]any)
		body, _ := container["body"].(map[string]any)
		state, _ := body["state"].(string)
		switch dawgtypes.EvidenceState(state) {
		case dawgtypes.EvidenceCaptured:
			if side == "request" {
				summary.RetainedRequestBodies++
			} else {
				summary.RetainedResponseBodies++
			}
		case dawgtypes.EvidenceRedacted:
			summary.RedactedBodies++
		case dawgtypes.EvidenceBlocked:
			summary.BlockedBodies++
		case dawgtypes.EvidenceTruncated:
			summary.TruncatedBodies++
		}
	}
}
