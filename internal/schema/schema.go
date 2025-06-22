package schema

import (
	"fmt"
	"os"
	"strconv"
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

	// Convert row to a map and attempt to convert types based on schema
	rowData := make(map[string]interface{})
	for i, header := range headers {
		// Default to string
		var value interface{} = data[i]

		// Check schema for type information
		if prop, ok := v.schema.Properties[header]; ok {
			// A property can have multiple types, e.g., ["number", "null"]
			for _, t := range prop.Types {
				if t == "integer" {
					if v, err := strconv.Atoi(data[i]); err == nil {
						value = v
						break
					}
				} else if t == "number" {
					if v, err := strconv.ParseFloat(data[i], 64); err == nil {
						value = v
						break
					}
				}
			}
		}
		rowData[header] = value
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
		// Extract field name from instance location (e.g., "/email")
		field := strings.TrimPrefix(err.InstanceLocation, "/")

		// Find the original string value for reporting
		originalValue := ""
		if val, exists := data[field]; exists {
			originalValue = fmt.Sprintf("%v", val)
		}

		errors = append(errors, ValidationError{
			Field:   field,
			Message: err.Message,
			Value:   originalValue,
		})
	}

	// Recursively process nested validation errors
	for _, cause := range err.Causes {
		errors = append(errors, v.convertValidationErrors(cause, data)...)
	}

	return errors
}
