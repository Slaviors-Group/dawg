package packager

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	value, err := validateOCILayout(input)
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

func validateOCILayout(directory string) (manifest.Manifest, error) {
	value, err := manifest.ReadValidated(filepath.Join(directory, "dawg-manifest.json"))
	if err != nil {
		return manifest.Manifest{}, err
	}
	layoutContents, err := os.ReadFile(filepath.Join(directory, "oci-layout"))
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI layout: %w", err)
	}
	var layout struct {
		ImageLayoutVersion string `json:"imageLayoutVersion"`
	}
	if err := json.Unmarshal(layoutContents, &layout); err != nil || layout.ImageLayoutVersion != "1.0.0" {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI layout")
	}
	var index ociIndex
	if err := readOCIJSON(filepath.Join(directory, "index.json"), &index); err != nil {
		return manifest.Manifest{}, fmt.Errorf("read OCI index: %w", err)
	}
	if index.SchemaVersion != 2 || index.MediaType != ociImageIndexMediaType || len(index.Manifests) != 1 {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI index")
	}
	manifestContents, err := verifyOCIDescriptor(directory, index.Manifests[0])
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI manifest descriptor: %w", err)
	}
	var image ociImageManifest
	if err := json.Unmarshal(manifestContents, &image); err != nil {
		return manifest.Manifest{}, fmt.Errorf("decode OCI manifest: %w", err)
	}
	if image.SchemaVersion != 2 || image.MediaType != ociImageManifestMediaType || image.Config.MediaType != string(dawgtypes.MediaTypeManifest) {
		return manifest.Manifest{}, fmt.Errorf("invalid OCI manifest")
	}
	config, err := verifyOCIDescriptor(directory, image.Config)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("validate OCI config descriptor: %w", err)
	}
	manifestJSON, err := manifest.Marshal(value)
	if err != nil || string(config) != string(manifestJSON) {
		return manifest.Manifest{}, fmt.Errorf("OCI config does not match dawg-manifest.json")
	}
	if len(image.Layers) != len(value.Layers) {
		return manifest.Manifest{}, fmt.Errorf("OCI layer descriptors do not match DAWG manifest")
	}
	for index, descriptor := range image.Layers {
		layer := value.Layers[index]
		if descriptor.MediaType != string(layer.MediaType) || descriptor.Digest != layer.Digest || descriptor.Size != layer.Size {
			return manifest.Manifest{}, fmt.Errorf("OCI layer descriptor %d does not match DAWG manifest", index)
		}
		if _, err := verifyOCIDescriptor(directory, descriptor); err != nil {
			return manifest.Manifest{}, fmt.Errorf("validate OCI layer descriptor: %w", err)
		}
	}
	return value, nil
}

func readOCIJSON(path string, target any) error {
	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(contents, target)
}

func verifyOCIDescriptor(directory string, descriptor ociDescriptor) ([]byte, error) {
	if descriptor.MediaType == "" || descriptor.Size < 0 {
		return nil, fmt.Errorf("invalid descriptor")
	}
	digestPath, err := layerDigestPath(descriptor.Digest)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(directory, "blobs", "sha256", digestPath)
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("OCI blob is not a regular file")
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	actual := sha256.Sum256(contents)
	if descriptor.Digest != fmt.Sprintf("sha256:%x", actual) || descriptor.Size != int64(len(contents)) {
		return nil, fmt.Errorf("descriptor digest or size mismatch: %s", descriptor.Digest)
	}
	return contents, nil
}
