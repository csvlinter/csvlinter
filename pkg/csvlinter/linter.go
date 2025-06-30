package csvlinter

import (
	"os"

	"github.com/csvlinter/csvlinter/internal/schema"
	"github.com/csvlinter/csvlinter/internal/validator"
)

// Package csvlinter provides a public API for validating CSV files from any io.Reader, with or without a schema.
//
// Example usage:
//
//   results, err := csvlinter.Lint("file.csv", ",")
//   results, err := csvlinter.LintWithSchema("file.csv", "schema.json", ",")
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

// Lint validates a CSV file without a schema.
// filePath: path to the CSV file
// delimiter: field delimiter (e.g., ",", ";", "\t")
// Returns validation results and error if any.
func Lint(filePath string, delimiter string) (*validator.Results, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	v := validator.New(file, filePath, delimiter, nil, false)
	return v.Validate()
}

// LintWithSchema validates a CSV file with a schema file.
// filePath: path to the CSV file
// schemaPath: path to the JSON schema file
// delimiter: field delimiter (e.g., ",", ";", "\t")
// Returns validation results and error if any.
func LintWithSchema(filePath string, schemaPath string, delimiter string) (*validator.Results, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check schema file exists
	if _, err := os.Stat(schemaPath); err != nil {
		return nil, err
	}

	schemaValidator, err := schema.NewValidator(schemaPath)
	if err != nil {
		return nil, err
	}

	v := validator.New(file, filePath, delimiter, schemaValidator, false)
	return v.Validate()
}
