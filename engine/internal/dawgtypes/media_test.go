package dawgtypes

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCanonicalMediaTypesMatchSchema(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("..", "..", "..", "schema", "mediatypes.json"))
	if err != nil {
		t.Fatalf("read media types schema: %v", err)
	}

	var schemaTypes map[string]MediaType
	if err := json.Unmarshal(contents, &schemaTypes); err != nil {
		t.Fatalf("decode media types schema: %v", err)
	}
	if !reflect.DeepEqual(schemaTypes, CanonicalMediaTypes()) {
		t.Fatalf("Go media types differ from schema: got %#v, want %#v", CanonicalMediaTypes(), schemaTypes)
	}
}
