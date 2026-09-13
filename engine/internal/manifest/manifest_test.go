package manifest

import (
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

func schemaPath(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..", "..", "schema", "manifest", "v0.2.3-naughty.json")
}
