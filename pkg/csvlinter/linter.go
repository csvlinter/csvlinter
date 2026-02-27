package csvlinter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/csvlinter/csvlinter/internal/reporter"
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
//   // Validate CSV with schema from stream (no file on disk)
//   csvReader := strings.NewReader(csvContent)
//   schemaReader := strings.NewReader(schemaJSON)
//   opts := csvlinter.Options{SchemaReader: schemaReader, Delimiter: ",", Format: "json", Filename: "data.csv"}
//   var buf bytes.Buffer
//   err := csvlinter.LintAdvanced(csvReader, opts, &buf)

// Options defines advanced options for CSV linting.
type Options struct {
	Delimiter    string   // Field delimiter (e.g., ",", ";", "\t")
	FailFast     bool     // Stop after first error
	Format       string   // Output format: "pretty" or "json"
	Output       string   // Output file path (if empty, write to writer)
	Filename     string   // Logical filename for schema resolution (used if reading from stream)
	SchemaPath   string   // Path to JSON schema file (optional)
	SchemaReader io.Reader // Optional: read JSON schema from this stream; takes precedence over SchemaPath when set
}

// LintAdvanced validates a CSV stream with advanced options, matching CLI capabilities.
// r: CSV data stream
// opts: advanced options (see Options struct)
// writer: where to write output (if opts.Output is empty)
func LintAdvanced(r io.Reader, opts Options, writer io.Writer) error {
	// Determine name for reporting
	name := opts.Filename
	if name == "" {
		name = "STDIN"
	}

	// Schema resolution logic: SchemaReader takes precedence over SchemaPath
	var schemaValidator *schema.Validator
	var err error
	if opts.SchemaReader != nil {
		schemaValidator, err = schema.NewValidatorFromReader(opts.SchemaReader)
		if err != nil {
			return err
		}
	} else {
		schemaPath := opts.SchemaPath
		if schemaPath != "" {
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				return fmt.Errorf("Schema file '%s' does not exist", schemaPath)
			}
		} else if opts.Filename != "" {
			schemaPath = schema.ResolveSchema(opts.Filename)
		}
		if schemaPath != "" {
			schemaValidator, err = schema.NewValidator(schemaPath)
			if err != nil {
				return err
			}
		}
	}

	// Validate format
	format := opts.Format
	if format == "" {
		format = "pretty"
	}
	if format != "pretty" && format != "json" {
		return fmt.Errorf("Format must be 'pretty' or 'json'")
	}

	delimiter := opts.Delimiter
	if delimiter == "" {
		delimiter = ","
	}

	// Create validator
	v := validator.New(r, name, delimiter, schemaValidator, opts.FailFast)
	results, err := v.Validate()
	if err != nil {
		return err
	}

	// Create reporter
	rep := reporter.New(format, opts.Output)
	return rep.Report(results, writer)
}

// Lint validates a CSV stream without a schema (backward compatible).
func Lint(r io.Reader, name string, delimiter string) (*validator.Results, error) {
	opts := Options{
		Delimiter: delimiter,
		Filename:  name,
		Format:    "json",
	}
	var buf bytes.Buffer
	err := LintAdvanced(r, opts, &buf)
	if err != nil {
		return nil, err
	}
	// Parse results from buffer (JSON is easiest for programmatic use)
	var results validator.Results
	jsonErr := json.Unmarshal(buf.Bytes(), &results)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return &results, nil
}

// LintWithSchema validates a CSV stream with a schema loaded from a file (backward compatible).
func LintWithSchema(r io.Reader, name string, delimiter string, schemaPath string) (*validator.Results, error) {
	opts := Options{
		Delimiter:  delimiter,
		Filename:   name,
		SchemaPath: schemaPath,
		Format:     "json",
	}
	var buf bytes.Buffer
	err := LintAdvanced(r, opts, &buf)
	if err != nil {
		return nil, err
	}
	var results validator.Results
	jsonErr := json.Unmarshal(buf.Bytes(), &results)
	if jsonErr != nil {
		return nil, jsonErr
	}
	return &results, nil
}
