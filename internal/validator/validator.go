package validator

import (
	"fmt"
	"time"

	"csvlinter/internal/parser"
	"csvlinter/internal/schema"
)

// Error represents a validation error
type Error struct {
	LineNumber int    `json:"line_number"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
	Value      string `json:"value,omitempty"`
	Type       string `json:"type"`
}

// Warning represents a validation warning
type Warning struct {
	LineNumber int    `json:"line_number"`
	Field      string `json:"field,omitempty"`
	Message    string `json:"message"`
	Value      string `json:"value,omitempty"`
	Type       string `json:"type"`
}

// Results contains the validation results
type Results struct {
	File       string    `json:"file"`
	TotalRows  int       `json:"total_rows"`
	Errors     []Error   `json:"errors"`
	Warnings   []Warning `json:"warnings"`
	Duration   string    `json:"duration"`
	Valid      bool      `json:"valid"`
	SchemaUsed bool      `json:"schema_used"`
}

// Validator represents the main validation engine
type Validator struct {
	filePath        string
	delimiter       string
	schemaValidator *schema.Validator
	failFast        bool
}

// New creates a new validator
func New(filePath, delimiter string, schemaValidator *schema.Validator, failFast bool) *Validator {
	return &Validator{
		filePath:        filePath,
		delimiter:       delimiter,
		schemaValidator: schemaValidator,
		failFast:        failFast,
	}
}

// Validate performs the complete validation process
func (v *Validator) Validate() (*Results, error) {
	startTime := time.Now()

	// Create parser
	p, err := parser.NewParser(v.filePath, v.delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	defer p.Close()

	// Validate UTF-8 encoding
	if err := p.ValidateUTF8(); err != nil {
		return &Results{
			File:     v.filePath,
			Valid:    false,
			Errors:   []Error{{Message: err.Error(), Type: "encoding"}},
			Duration: time.Since(startTime).String(),
		}, nil
	}

	// Read headers
	headers, err := p.ReadHeaders()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	var errors []Error
	var warnings []Warning
	totalRows := 0

	// Validate each row
	for {
		row, err := p.ReadRow()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		totalRows++

		// Basic structure validation
		if len(row.Data) != len(headers) {
			errors = append(errors, Error{
				LineNumber: row.LineNumber,
				Field:      "row",
				Message:    fmt.Sprintf("column count mismatch: expected %d, got %d", len(headers), len(row.Data)),
				Type:       "structure",
			})
		}

		// Schema validation if available
		if v.schemaValidator != nil {
			schemaErrors, err := v.schemaValidator.ValidateRow(headers, row.Data)
			if err != nil {
				return nil, fmt.Errorf("schema validation error on line %d: %w", row.LineNumber, err)
			}

			for _, schemaErr := range schemaErrors {
				errors = append(errors, Error{
					LineNumber: row.LineNumber,
					Field:      schemaErr.Field,
					Message:    schemaErr.Message,
					Value:      schemaErr.Value,
					Type:       "schema",
				})
			}
		}

		// Fail fast if requested
		if v.failFast && len(errors) > 0 {
			break
		}
	}

	duration := time.Since(startTime)
	valid := len(errors) == 0

	return &Results{
		File:       v.filePath,
		TotalRows:  totalRows,
		Errors:     errors,
		Warnings:   warnings,
		Duration:   duration.String(),
		Valid:      valid,
		SchemaUsed: v.schemaValidator != nil,
	}, nil
}
