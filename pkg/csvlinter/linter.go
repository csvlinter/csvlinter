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

// LintWithSchema validates a CSV stream with a schema loaded from a file.
// r: CSV data stream
// name: label for reporting (e.g., filename)
// delimiter: field delimiter (e.g., ",", ";", "\t")
// schemaPath: path to the JSON schema file
func LintWithSchema(r io.Reader, name string, delimiter string, schemaPath string) (*validator.Results, error) {
	schemaValidator, err := schema.NewValidator(schemaPath)
	if err != nil {
		return nil, err
	}
	v := validator.New(r, name, delimiter, schemaValidator, false)
	return v.Validate()
}
