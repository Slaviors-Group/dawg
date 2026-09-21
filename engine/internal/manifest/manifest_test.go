package manifest

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Slaviors-Group/dawg/engine/internal/dawgtypes"
)

func TestManifestRoundTripIsStable(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("testdata", "valid.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	value, err := Unmarshal(contents)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	first, err := Marshal(value)
	if err != nil {
		t.Fatalf("first serialization: %v", err)
	}
	secondValue, err := Unmarshal(first)
	if err != nil {
		t.Fatalf("decode first serialization: %v", err)
	}
	second, err := Marshal(secondValue)
	if err != nil {
		t.Fatalf("second serialization: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("manifest serialization was not stable\nfirst: %s\nsecond: %s", first, second)
	}
}

func TestValidateJSONAcceptsValidFixture(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("testdata", "valid.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := ValidateJSON(schemaPath(t), contents); err != nil {
		t.Fatalf("validate fixture: %v", err)
	}
}

func TestValidateJSONRejectsInvalidFixture(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("testdata", "invalid-missing-title.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	err = ValidateJSON(schemaPath(t), contents)
	if !errors.Is(err, dawgtypes.ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest error, got %v", err)
	}
}

func TestValidateJSONRejectsMalformedSchema(t *testing.T) {
	badSchema := filepath.Join(t.TempDir(), "schema.json")
	if err := os.WriteFile(badSchema, []byte("{"), 0o600); err != nil {
		t.Fatalf("write malformed schema: %v", err)
	}
	err := ValidateJSON(badSchema, []byte("{}"))
	if !errors.Is(err, dawgtypes.ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest error, got %v", err)
	}
}

func TestReadValidatedAcceptsSupportedVersions(t *testing.T) {
	contents := validManifestContents(t)
	for _, version := range []string{"0.2.3-naughty", "0.2.5-naughty", "0.2.7-naughty", SchemaVersion} {
		t.Run(version, func(t *testing.T) {
			versionedContents := contents
			if version != SchemaVersion {
				var legacy map[string]any
				if err := json.Unmarshal(contents, &legacy); err != nil {
					t.Fatalf("decode current fixture: %v", err)
				}
				delete(legacy, "diagnostics")
				legacy["schemaVersion"] = version
				legacy["layers"] = []any{map[string]any{
					"mediaType": string(dawgtypes.MediaTypeEnvironment),
					"digest":    "sha256:abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
					"size":      2048,
				}}
				versionedContents, _ = json.Marshal(legacy)
			}
			path := writeManifest(t, versionedContents)

			value, err := ReadValidated(path)
			if err != nil {
				t.Fatalf("read validated manifest: %v", err)
			}
			if value.SchemaVersion != version {
				t.Fatalf("expected schema version %q, got %q", version, value.SchemaVersion)
			}
		})
	}
}

func TestReadValidatedRejectsUnsupportedAndPathLikeVersions(t *testing.T) {
	contents := validManifestContents(t)
	for _, version := range []string{"0.2.4-naughty", "../../v0.2.5-naughty.json"} {
		t.Run(version, func(t *testing.T) {
			versionedContents := bytes.Replace(contents, []byte(SchemaVersion), []byte(version), 1)
			_, err := ReadValidated(writeManifest(t, versionedContents))
			if !errors.Is(err, dawgtypes.ErrInvalidManifest) {
				t.Fatalf("expected invalid manifest error, got %v", err)
			}
		})
	}
}

func TestReadValidatedRejectsAdditionalProperties(t *testing.T) {
	contents := validManifestContents(t)
	contents = bytes.Replace(contents, []byte(`  "title":`), []byte("  \"unexpected\": true,\n  \"title\":"), 1)

	_, err := ReadValidated(writeManifest(t, contents))
	if !errors.Is(err, dawgtypes.ErrInvalidManifest) {
		t.Fatalf("expected invalid manifest error, got %v", err)
	}
}

func TestSchemaPathUsesVersionSpecificSchemaWithOverride(t *testing.T) {
	currentPath := schemaPath(t)
	t.Setenv("DAWG_SCHEMA_PATH", currentPath)

	legacyPath, err := SchemaPath("0.2.3-naughty")
	if err != nil {
		t.Fatalf("locate legacy schema: %v", err)
	}
	if legacyPath == currentPath || filepath.Base(legacyPath) != "v0.2.3-naughty.json" {
		t.Fatalf("legacy version resolved to wrong schema: %s", legacyPath)
	}
}

func validManifestContents(t *testing.T) []byte {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("testdata", "valid.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return contents
}

func writeManifest(t *testing.T, contents []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dawg-manifest.json")
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func schemaPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "schema", "manifest", "v"+SchemaVersion+".json")
}
