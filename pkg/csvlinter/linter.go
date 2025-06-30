package csvlinter

import (
	"io"

	"github.com/csvlinter/csvlinter/internal/schema"
	"github.com/csvlinter/csvlinter/internal/validator"
)

// Package csvlinter provides a public API for validating CSV files from any io.Reader, with or without a schema.
//
// Example usage:
//
//   // Validate CSV without schema
//   f, _ := os.Open("file.csv")
//   results, err := csvlinter.Lint(f, "file.csv", ",")
//
//   // Validate CSV with schema (both as streams)
//   csvFile, _ := os.Open("file.csv")
//   schemaFile, _ := os.Open("schema.json")
//   schemaValidator, err := schema.NewValidatorFromReader(schemaFile)
//   results, err := csvlinter.LintWithSchema(csvFile, "file.csv", ",", schemaValidator)

// Lint validates a CSV stream without a schema.
// r: CSV data stream
// name: label for reporting (e.g., filename)
// delimiter: field delimiter (e.g., ",", ";", "\t")
func Lint(r io.Reader, name string, delimiter string) (*validator.Results, error) {
	v := validator.New(r, name, delimiter, nil, false)
	return v.Validate()
}

// LintWithSchema validates a CSV stream with a provided schema validator.
// r: CSV data stream
// name: label for reporting (e.g., filename)
// delimiter: field delimiter (e.g., ",", ";", "\t")
// schemaValidator: compiled schema validator
func LintWithSchema(r io.Reader, name string, delimiter string, schemaValidator *schema.Validator) (*validator.Results, error) {
	v := validator.New(r, name, delimiter, schemaValidator, false)
	return v.Validate()
}
