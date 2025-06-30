package csvlinter

import (
	"io"

	"github.com/csvlinter/csvlinter/internal/schema"
	"github.com/csvlinter/csvlinter/internal/validator"
)

// Package csvlinter provides a public API for validating CSV files with or without a schema.
//
// Example usage:
//
//   results, err := csvlinter.Lint("file.csv", ",")
//   results, err := csvlinter.LintWithSchema("file.csv", "schema.json", ",")
//

// Lint validates a CSV stream without a schema.
// r: CSV data stream
// name: label for reporting (e.g., filename)
// delimiter: field delimiter (e.g., ",", ";", "\t")
// Returns validation results and error if any.
func Lint(r io.Reader, name string, delimiter string) (*validator.Results, error) {
	v := validator.New(r, name, delimiter, nil, false)
	return v.Validate()
}

// LintWithSchema validates a CSV stream with a provided schema validator.
// r: CSV data stream
// name: label for reporting (e.g., filename)
// delimiter: field delimiter (e.g., ",", ";", "\t")
// schemaValidator: compiled schema validator
// Returns validation results and error if any.
func LintWithSchema(r io.Reader, name string, delimiter string, schemaValidator *schema.Validator) (*validator.Results, error) {
	v := validator.New(r, name, delimiter, schemaValidator, false)
	return v.Validate()
}
