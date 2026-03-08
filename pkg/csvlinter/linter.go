package csvlinter

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/csvlinter/csvlinter/internal/parser"
	"github.com/csvlinter/csvlinter/internal/reporter"
	"github.com/csvlinter/csvlinter/internal/schema"
	"github.com/csvlinter/csvlinter/internal/validator"
)

// DefaultInferSchemaMaxRows is the default number of rows to use when inferring schema.
const DefaultInferSchemaMaxRows = 1000

// DefaultInferSchemaMaxBytes is the default max bytes to read when inferring schema (e.g. from STDIN).
const DefaultInferSchemaMaxBytes = 10 * 1024 * 1024

// Options configures CSV validation and output for LintAdvanced.
// Delimiter defaults to "," and Format to "pretty" when empty.
type Options struct {
	Delimiter           string    // Field delimiter (e.g., ",", ";", "\t")
	FailFast            bool     // Stop after first error
	Format              string   // Output format: "pretty" or "json"
	Output              string   // Output file path (if empty, write to writer)
	Filename            string   // Logical filename for schema resolution (used if reading from stream)
	SchemaPath          string   // Path to JSON schema file (optional)
	SchemaReader        io.Reader // Optional: read JSON schema from this stream; takes precedence over SchemaPath when set
	InferSchema         bool     // If true and no schema provided, infer schema from data
	InferSchemaOutput   string   // If non-empty, write inferred schema to this path
	InferSchemaMaxRows  int      // Max rows for inference (0 = DefaultInferSchemaMaxRows)
	InferSchemaMaxBytes int64    // Max bytes to buffer for inference (0 = DefaultInferSchemaMaxBytes)
}

// LintAdvanced validates a CSV stream with full control over schema, format, and output.
// CSV is read from r; formatted results are written to writer unless opts.Output is set,
// in which case output is written to that file. Schema may be provided via
// opts.SchemaReader (takes precedence) or opts.SchemaPath. If neither is set and
// opts.Filename is set, schema resolution looks for <filename>.schema.json or csvlinter.schema.json
// in the same directory, then csvlinter.schema.json in parent directories up to a project root
// (.git or package.json) or system root.
// Returns the validation results on success so callers can check results.Valid; on error results may be nil.
func LintAdvanced(r io.Reader, opts Options, writer io.Writer) (*validator.Results, error) {
	// Determine name for reporting
	name := opts.Filename
	if name == "" {
		name = "STDIN"
	}

	delimiter := opts.Delimiter
	if delimiter == "" {
		delimiter = ","
	}

	// Schema resolution logic: SchemaReader takes precedence over SchemaPath
	var schemaValidator *schema.Validator
	var schemaInferred bool
	var err error
	if opts.SchemaReader != nil {
		schemaValidator, err = schema.NewValidatorFromReader(opts.SchemaReader)
		if err != nil {
			return nil, err
		}
	} else {
		schemaPath := opts.SchemaPath
		if schemaPath != "" {
			if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("Schema file '%s' does not exist", schemaPath)
			}
		} else if opts.Filename != "" {
			schemaPath = schema.ResolveSchema(opts.Filename)
		}
		if schemaPath != "" {
			schemaValidator, err = schema.NewValidator(schemaPath)
			if err != nil {
				return nil, err
			}
		}
	}

	input := r
	if schemaValidator == nil && opts.InferSchema {
		maxBytes := opts.InferSchemaMaxBytes
		if maxBytes == 0 {
			maxBytes = DefaultInferSchemaMaxBytes
		}
		maxRows := opts.InferSchemaMaxRows
		if maxRows == 0 {
			maxRows = DefaultInferSchemaMaxRows
		}
		buf, readErr := io.ReadAll(io.LimitReader(r, maxBytes))
		if readErr != nil {
			return nil, fmt.Errorf("reading input for schema inference: %w", readErr)
		}
		headers, sample, sampleErr := parser.ReadSampleFromBytes(buf, delimiter, maxRows)
		if sampleErr != nil {
			return nil, sampleErr
		}
		schemaJSON, inferErr := schema.Infer(headers, sample)
		if inferErr != nil {
			return nil, inferErr
		}
		if opts.InferSchemaOutput != "" {
			if writeErr := os.WriteFile(opts.InferSchemaOutput, schemaJSON, 0644); writeErr != nil {
				return nil, fmt.Errorf("writing inferred schema: %w", writeErr)
			}
		}
		schemaValidator, err = schema.NewValidatorFromReader(bytes.NewReader(schemaJSON))
		if err != nil {
			return nil, err
		}
		schemaInferred = true
		input = bytes.NewReader(buf)
	}

	// Validate format
	format := opts.Format
	if format == "" {
		format = "pretty"
	}
	if format != "pretty" && format != "json" {
		return nil, fmt.Errorf("Format must be 'pretty' or 'json'")
	}

	// Create validator
	v := validator.New(input, name, delimiter, schemaValidator, opts.FailFast, schemaInferred)
	results, err := v.Validate()
	if err != nil {
		return nil, err
	}

	// Create reporter
	rep := reporter.New(format, opts.Output)
	if err := rep.Report(results, writer); err != nil {
		return nil, err
	}
	return results, nil
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
		Format:     "json",
	}
	var buf bytes.Buffer
	return LintAdvanced(r, opts, &buf)
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
	return LintAdvanced(r, opts, &buf)
}
