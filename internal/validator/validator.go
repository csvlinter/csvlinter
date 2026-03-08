package validator

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/csvlinter/csvlinter/internal/parser"
	"github.com/csvlinter/csvlinter/internal/schema"
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
	File            string    `json:"file"`
	TotalRows       int       `json:"total_rows"`
	Errors          []Error   `json:"errors"`
	Warnings        []Warning `json:"warnings"`
	Duration        string    `json:"duration"`
	Valid           bool      `json:"valid"`
	SchemaUsed      bool      `json:"schema_used"`
	SchemaInferred  bool      `json:"schema_inferred,omitempty"`
}

// Validator represents the main validation engine
type Validator struct {
	input            io.Reader
	name             string
	delimiter        string
	schemaValidator  *schema.Validator
	failFast         bool
	schemaInferred   bool
}

// New creates a new validator. schemaInferred should be true when the schema was inferred from data rather than loaded from file.
func New(input io.Reader, name string, delimiter string, schemaValidator *schema.Validator, failFast bool, schemaInferred bool) *Validator {
	return &Validator{
		input:           input,
		name:            name,
		delimiter:       delimiter,
		schemaValidator: schemaValidator,
		failFast:        failFast,
		schemaInferred:  schemaInferred,
	}
}

// Validate performs the complete validation process
func (v *Validator) Validate() (*Results, error) {
	startTime := time.Now()

	// Create parser
	p, err := parser.NewParser(v.input, v.delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}
	defer p.Close()

	// Read headers (UTF-8 validated inside ReadHeaders when streaming)
	headers, err := p.ReadHeaders()
	if err != nil {
		var encErr *parser.EncodingError
		if errors.As(err, &encErr) {
			return &Results{
				File:     v.name,
				Valid:    false,
				Errors:   []Error{{LineNumber: encErr.LineNumber, Message: "invalid UTF-8 encoding", Type: "encoding"}},
				Duration: time.Since(startTime).String(),
			}, nil
		}
		return nil, fmt.Errorf("failed to read headers: %w", err)
	}

	var errs []Error
	var warnings []Warning
	totalRows := 0

	// Validate each row
	for {
		row, err := p.ReadRow()
		if err != nil {
			if err == io.EOF {
				break
			}
			var encErr *parser.EncodingError
			errType := "structure"
			errMsg := err.Error()
			lineNum := p.GetLineNumber() + 1
			if errors.As(err, &encErr) {
				errType = "encoding"
				errMsg = "invalid UTF-8 encoding"
				lineNum = encErr.LineNumber
			}
			errs = append(errs, Error{
				LineNumber: lineNum,
				Message:    errMsg,
				Type:       errType,
			})
			break
		}

		// Skip empty rows, often caused by trailing newlines
		if row.IsEmpty() {
			continue
		}

		totalRows++

		// Basic structure validation
		if len(row.Data) != len(headers) {
			errs = append(errs, Error{
				LineNumber: row.LineNumber,
				Field:      "row",
				Message:    fmt.Sprintf("column count mismatch: expected %d, got %d", len(headers), len(row.Data)),
				Type:       "structure",
			})
			// Fail fast if requested
			if v.failFast {
				break
			}
			// Skip schema validation for this row
			continue
		}

		// Schema validation if available
		if v.schemaValidator != nil {
			schemaErrors, err := v.schemaValidator.ValidateRow(headers, row.Data)
			if err != nil {
				return nil, fmt.Errorf("schema validation error on line %d: %w", row.LineNumber, err)
			}

			for _, schemaErr := range schemaErrors {
				errs = append(errs, Error{
					LineNumber: row.LineNumber,
					Field:      schemaErr.Field,
					Message:    schemaErr.Message,
					Value:      schemaErr.Value,
					Type:       "schema",
				})
			}
		}

		// Fail fast if requested
		if v.failFast && len(errs) > 0 {
			break
		}
	}

	duration := time.Since(startTime)
	valid := len(errs) == 0

	return &Results{
		File:           v.name,
		TotalRows:      totalRows,
		Errors:         errs,
		Warnings:       warnings,
		Duration:       duration.String(),
		Valid:          valid,
		SchemaUsed:     v.schemaValidator != nil,
		SchemaInferred: v.schemaInferred,
	}, nil
}
