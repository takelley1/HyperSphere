// Path: internal/config/schema_test.go
// Description: Validate main config schema checks and failing-field path reporting.
package config

import "testing"

func TestSchemaValidationErrorStringIncludesFieldPath(t *testing.T) {
	err := SchemaValidationError{FieldPath: "ui.theme"}
	if err.Error() != "invalid config field: ui.theme" {
		t.Fatalf("unexpected schema error string: %q", err.Error())
	}
}

func TestValidateMainConfigAcceptsKnownFields(t *testing.T) {
	input := map[string]any{
		"mode":      "mark",
		"execute":   true,
		"threshold": 85,
		"ui": map[string]any{
			"theme": "default",
		},
	}
	if err := ValidateMainConfig(input); err != nil {
		t.Fatalf("expected valid config schema, got %v", err)
	}
}

func TestValidateMainConfigRejectsUnknownTopLevelFieldWithPath(t *testing.T) {
	input := map[string]any{
		"mode":          "mark",
		"unknown_field": true,
	}
	err := ValidateMainConfig(input)
	if err == nil {
		t.Fatalf("expected schema error for unknown field")
	}
	schemaErr, ok := err.(SchemaValidationError)
	if !ok {
		t.Fatalf("expected SchemaValidationError, got %T", err)
	}
	if schemaErr.FieldPath != "unknown_field" {
		t.Fatalf("expected failing path unknown_field, got %q", schemaErr.FieldPath)
	}
}

func TestValidateMainConfigRejectsUnknownNestedFieldWithPath(t *testing.T) {
	input := map[string]any{
		"ui": map[string]any{
			"theme":         "default",
			"unknown_child": "x",
		},
	}
	err := ValidateMainConfig(input)
	if err == nil {
		t.Fatalf("expected schema error for unknown nested field")
	}
	schemaErr, ok := err.(SchemaValidationError)
	if !ok {
		t.Fatalf("expected SchemaValidationError, got %T", err)
	}
	if schemaErr.FieldPath != "ui.unknown_child" {
		t.Fatalf("expected failing path ui.unknown_child, got %q", schemaErr.FieldPath)
	}
}

func TestValidateMainConfigRejectsTypeMismatchWithPath(t *testing.T) {
	input := map[string]any{
		"ui": "not-an-object",
	}
	err := ValidateMainConfig(input)
	if err == nil {
		t.Fatalf("expected schema error for type mismatch")
	}
	schemaErr, ok := err.(SchemaValidationError)
	if !ok {
		t.Fatalf("expected SchemaValidationError, got %T", err)
	}
	if schemaErr.FieldPath != "ui" {
		t.Fatalf("expected failing path ui, got %q", schemaErr.FieldPath)
	}
}
