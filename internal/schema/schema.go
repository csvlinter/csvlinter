package schema

import (
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Validator represents a JSON Schema validator
type Validator struct {
	schema *jsonschema.Schema
}

// ValidationError represents a schema validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
}

// NewValidator creates a new schema validator from a JSON Schema file
func NewValidator(schemaPath string) (*Validator, error) {
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("schema.json", strings.NewReader(string(schemaBytes))); err != nil {
		return nil, fmt.Errorf("failed to add schema resource: %w", err)
	}

	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator{
		schema: schema,
	}, nil
}

// ValidateRow validates a CSV row against the JSON Schema
func (v *Validator) ValidateRow(headers []string, data []string) ([]ValidationError, error) {
	if len(headers) != len(data) {
		return []ValidationError{{
			Field:   "row",
			Message: fmt.Sprintf("mismatched columns: headers=%d, data=%d", len(headers), len(data)),
			Value:   "",
		}}, nil
	}

	// Convert row to JSON object
	rowData := make(map[string]interface{})
	for i, header := range headers {
		rowData[header] = data[i]
	}

	// Validate against schema
	if err := v.schema.Validate(rowData); err != nil {
		if validationErr, ok := err.(*jsonschema.ValidationError); ok {
			return v.convertValidationErrors(validationErr, rowData), nil
		}
		return nil, fmt.Errorf("schema validation error: %w", err)
	}

	return nil, nil
}

// convertValidationErrors converts jsonschema validation errors to our format
func (v *Validator) convertValidationErrors(err *jsonschema.ValidationError, data map[string]interface{}) []ValidationError {
	var errors []ValidationError

	if err.InstanceLocation != "" {
		// Extract field name from instance location
		field := err.InstanceLocation
		if field[0] == '/' {
			field = field[1:]
		}

		value := ""
		if val, exists := data[field]; exists {
			if str, ok := val.(string); ok {
				value = str
			}
		}

		errors = append(errors, ValidationError{
			Field:   field,
			Message: err.Message,
			Value:   value,
		})
	}

	// Recursively process nested errors
	for _, nestedErr := range err.Causes {
		errors = append(errors, v.convertValidationErrors(nestedErr, data)...)
	}

	return errors
}

// GetSchemaInfo returns basic information about the schema
func (v *Validator) GetSchemaInfo() (map[string]interface{}, error) {
	// This is a simplified version - in a real implementation you might want
	// to extract more detailed schema information
	return map[string]interface{}{
		"type": "json-schema",
	}, nil
}
