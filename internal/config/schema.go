// Path: internal/config/schema.go
// Description: Enforce main-config schema validation with field-path reporting.
package config

import "fmt"

// SchemaValidationError reports the exact field path that violates schema rules.
type SchemaValidationError struct {
	FieldPath string
}

func (e SchemaValidationError) Error() string {
	return fmt.Sprintf("invalid config field: %s", e.FieldPath)
}

// ValidateMainConfig validates main config fields and nested objects against the schema.
func ValidateMainConfig(input map[string]any) error {
	return validateConfigMap(input, mainConfigSchema(), "")
}

func mainConfigSchema() map[string]any {
	return map[string]any{
		"mode":             nil,
		"execute":          nil,
		"threshold":        nil,
		"non_interactive":  nil,
		"config_dir":       nil,
		"ui":               map[string]any{"theme": nil},
		"hotkeys_file":     nil,
		"aliases_file":     nil,
		"plugins_file":     nil,
		"skin_file":        nil,
		"endpoint_overlay": nil,
	}
}

func validateConfigMap(input map[string]any, schema map[string]any, prefix string) error {
	for key, value := range input {
		path := schemaPath(prefix, key)
		expected, ok := schema[key]
		if !ok {
			return SchemaValidationError{FieldPath: path}
		}
		childSchema, child := expected.(map[string]any)
		if !child {
			continue
		}
		childValue, ok := value.(map[string]any)
		if !ok {
			return SchemaValidationError{FieldPath: path}
		}
		if err := validateConfigMap(childValue, childSchema, path); err != nil {
			return err
		}
	}
	return nil
}

func schemaPath(prefix string, field string) string {
	if prefix == "" {
		return field
	}
	return prefix + "." + field
}
