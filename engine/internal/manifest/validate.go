package manifest

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dawg-placeholder/dawg/engine/internal/dawgtypes"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validate checks a manifest against the schema at schemaPath.
func Validate(schemaPath string, value Manifest) error {
	contents, err := Marshal(value)
	if err != nil {
		return err
	}
	return ValidateJSON(schemaPath, contents)
}

// ValidateJSON checks raw manifest JSON against the schema at schemaPath.
func ValidateJSON(schemaPath string, contents []byte) error {
	schemaContents, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("%w: read schema %s: %v", dawgtypes.ErrInvalidManifest, schemaPath, err)
	}

	var schemaDocument any
	if err := json.Unmarshal(schemaContents, &schemaDocument); err != nil {
		return fmt.Errorf("%w: decode schema %s: %v", dawgtypes.ErrInvalidManifest, schemaPath, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	if err := compiler.AddResource("dawg-manifest.json", schemaDocument); err != nil {
		return fmt.Errorf("%w: load schema %s: %v", dawgtypes.ErrInvalidManifest, schemaPath, err)
	}
	schema, err := compiler.Compile("dawg-manifest.json")
	if err != nil || schema == nil {
		if err == nil {
			err = fmt.Errorf("compiler returned no schema")
		}
		return fmt.Errorf("%w: compile schema %s: %v", dawgtypes.ErrInvalidManifest, schemaPath, err)
	}

	var document any
	if err := json.Unmarshal(contents, &document); err != nil {
		return fmt.Errorf("%w: decode JSON: %v", dawgtypes.ErrInvalidManifest, err)
	}
	if err := schema.Validate(document); err != nil {
		return fmt.Errorf("%w: schema validation: %v", dawgtypes.ErrInvalidManifest, err)
	}
	return nil
}
