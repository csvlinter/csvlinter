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

// Options configures CSV validation and output for LintAdvanced.
// Delimiter defaults to "," and Format to "pretty" when empty.
type Options struct {
	Delimiter    string    // Field delimiter (e.g., ",", ";", "\t")
	FailFast     bool      // Stop after first error
	Format       string    // Output format: "pretty" or "json"
	Output       string    // Output file path (if empty, write to writer)
	Filename     string    // Logical filename for schema resolution (used if reading from stream)
	SchemaPath   string    // Path to JSON schema file (optional)
	SchemaReader io.Reader // Optional: read JSON schema from this stream; takes precedence over SchemaPath when set
}

// LintAdvanced validates a CSV stream with full control over schema, format, and output.
// CSV is read from r; formatted results are written to writer unless opts.Output is set,
// in which case output is written to that file. Schema may be provided via
// opts.SchemaReader (takes precedence) or opts.SchemaPath. If neither is set and
// opts.Filename is set, schema resolution looks for <filename>.schema.json or csvlinter.schema.json
// in the same directory, then csvlinter.schema.json in parent directories up to a project root
// (.git or package.json) or system root.
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

// Lint validates a CSV stream and returns structured results.
// It does not take an explicit schema path; a schema may still be auto-resolved from name
// when name looks like a file path (see LintAdvanced for resolution rules). name is used for
// reporting (e.g. filename or "STDIN"); delimiter is the field separator (e.g. "," or ";").
// Returns results with File, TotalRows, Errors, Warnings, Duration, Valid; caller can check results.Valid and results.Errors.
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

// LintWithSchema validates a CSV stream against a JSON Schema file at schemaPath.
// name is used for reporting; delimiter is the field separator. Returns the same
// results shape as Lint, with SchemaUsed true when the schema was applied.
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
