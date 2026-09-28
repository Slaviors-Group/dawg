package packager

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Slaviors-Group/dawg/engine/internal/artifact"
	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
	"github.com/Slaviors-Group/dawg/engine/internal/manifest"
)

// DiagnosticRemovalRequest creates a new artifact without selected diagnostic
// categories or retained body files. The input artifact is never modified.
type DiagnosticRemovalRequest struct {
	InputDirectory  string
	OutputDirectory string
	Categories      []string
	BodyRefs        []string
}

// RepackageDiagnostics removes selected evidence from a current diagnostic
// artifact and publishes a new OCI layout with recomputed digests and logical
// artifact ID. It validates the complete source OCI layout before staging.
func RepackageDiagnostics(request DiagnosticRemovalRequest) (dawgtypes.PackagedArtifact, error) {
	if request.InputDirectory == "" || request.OutputDirectory == "" || (len(request.Categories) == 0 && len(request.BodyRefs) == 0) {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: input, output, and diagnostic categories or body references are required")
	}
	input, err := filepath.Abs(request.InputDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	output, err := filepath.Abs(request.OutputDirectory)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	if input == output || strings.HasPrefix(output, input+string(os.PathSeparator)) {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: output must not be the input artifact or its descendant")
	}
	if _, err := os.Stat(output); err == nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: output directory already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return dawgtypes.PackagedArtifact{}, err
	}
	value, err := artifact.Validate(input)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: validate input OCI layout: %w", err)
	}
	if value.SchemaVersion != manifest.SchemaVersion || value.Diagnostics == nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: diagnostic removal requires a %s artifact with diagnostics", manifest.SchemaVersion)
	}
	categories, err := validatedDiagnosticCategories(request.Categories)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	bodyRefs, err := validatedBodyRefs(request.BodyRefs)
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	staging, err := os.MkdirTemp(filepath.Dir(output), ".dawg-diagnostic-removal-")
	if err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	defer os.RemoveAll(staging)
	if err := Unpack(input, staging); err != nil {
		return dawgtypes.PackagedArtifact{}, fmt.Errorf("packager: stage artifact: %w", err)
	}
	for category, relative := range map[string]string{
		"console": "diagnostics/console.jsonl",
		"network": "diagnostics/network.jsonl",
		"errors":  "diagnostics/errors.jsonl",
		"device":  "diagnostics/device.jsonl",
	} {
		if categories[category] {
			if err := os.Remove(filepath.Join(staging, filepath.FromSlash(relative))); err != nil && !os.IsNotExist(err) {
				return dawgtypes.PackagedArtifact{}, err
			}
		}
	}
	if categories["network"] || categories["bodies"] {
		if err := os.RemoveAll(filepath.Join(staging, "diagnostics", "bodies")); err != nil {
			return dawgtypes.PackagedArtifact{}, err
		}
		if !categories["network"] {
			if err := clearNetworkBodyReferences(filepath.Join(staging, "diagnostics", "network.jsonl"), nil); err != nil {
				return dawgtypes.PackagedArtifact{}, err
			}
		}
	} else if len(bodyRefs) > 0 {
		for ref := range bodyRefs {
			if err := os.Remove(filepath.Join(staging, filepath.FromSlash(ref))); err != nil && !os.IsNotExist(err) {
				return dawgtypes.PackagedArtifact{}, err
			}
		}
		if err := clearNetworkBodyReferences(filepath.Join(staging, "diagnostics", "network.jsonl"), bodyRefs); err != nil {
			return dawgtypes.PackagedArtifact{}, err
		}
	}
	metadata := dawgtypes.CaptureMetadata{
		SessionID:   value.ID,
		StartedAt:   value.CreatedAt,
		StoppedAt:   value.CreatedAt,
		Determinism: value.Determinism,
		Diagnostics: *value.Diagnostics,
	}
	if err := writeJSON(filepath.Join(staging, "meta.json"), metadata); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	report := dawgtypes.SanitizeReport{
		PolicyVersion:  value.Sanitize.PolicyVersion,
		PolicyFile:     "artifact-policy",
		OPAResult:      "allow",
		ExportAllowed:  true,
		FieldsRedacted: value.Sanitize.FieldsRedacted,
		Redactions:     []dawgtypes.Redaction{},
		BlockedFields:  []string{},
	}
	if err := writeJSON(filepath.Join(staging, "sanitize-report.json"), report); err != nil {
		return dawgtypes.PackagedArtifact{}, err
	}
	return Package(dawgtypes.PackageRequest{
		SessionDirectory: staging,
		OutputDirectory:  output,
		Title:            value.Title,
		Source:           value.Source,
		ExpectedOutcome:  value.ExpectedOutcome,
	})
}

func validatedDiagnosticCategories(categories []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, category := range categories {
		if category != "console" && category != "network" && category != "errors" && category != "device" && category != "bodies" {
			return nil, fmt.Errorf("packager: unknown diagnostic category %q", category)
		}
		if result[category] {
			return nil, fmt.Errorf("packager: duplicate diagnostic category %q", category)
		}
		result[category] = true
	}
	return result, nil
}

func validatedBodyRefs(refs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, ref := range refs {
		normalized := filepath.ToSlash(filepath.Clean(filepath.FromSlash(ref)))
		if !strings.HasPrefix(normalized, "diagnostics/bodies/") || normalized == "diagnostics/bodies/" || strings.Contains(normalized, "..") || filepath.IsAbs(ref) {
			return nil, fmt.Errorf("packager: invalid diagnostic body reference %q", ref)
		}
		if result[normalized] {
			return nil, fmt.Errorf("packager: duplicate diagnostic body reference %q", ref)
		}
		result[normalized] = true
	}
	return result, nil
}

// clearNetworkBodyReferences marks either all body references (removed == nil)
// or selected body references unavailable. It never invents replacement data.
func clearNetworkBodyReferences(path string, removed map[string]bool) error {
	contents, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(contents)), "\n")
	output := make([]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("packager: decode network record: %w", err)
		}
		for _, side := range []string{"request", "response"} {
			section, _ := record[side].(map[string]any)
			body, _ := section["body"].(map[string]any)
			if body == nil {
				continue
			}
			ref, _ := body["ref"].(string)
			if removed != nil && !removed[ref] {
				continue
			}
			body["state"] = "unavailable"
			for _, key := range []string{"ref", "sha256", "size", "contentType"} {
				delete(body, key)
			}
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			return err
		}
		output = append(output, string(encoded))
	}
	return os.WriteFile(path, []byte(strings.Join(output, "\n")+"\n"), 0o600)
}
