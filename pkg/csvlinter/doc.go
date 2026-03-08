// Package csvlinter provides a streaming CSV validator with optional JSON Schema support.
//
// Validate CSV from any io.Reader (files, network, in-memory) without loading the whole
// file. Supports structure checks (column count, UTF-8), JSON Schema validation, and
// configurable output (pretty or JSON).
//
// Result errors have Type one of "structure" (column count, malformed row), "schema"
// (JSON Schema validation), or "encoding" (invalid UTF-8).
//
// Basic usage:
//
//	// No explicit schema (schema may still be auto-resolved from name if it looks like a path)
//	f, _ := os.Open("file.csv")
//	results, err := csvlinter.Lint(f, "file.csv", ",")
//
//	// With schema file
//	results, err := csvlinter.LintWithSchema(f, "file.csv", ",", "schema.json")
//
//	// Full control: schema from stream, format, fail-fast
//	opts := csvlinter.Options{
//	    SchemaReader: strings.NewReader(schemaJSON),
//	    Delimiter:    ",",
//	    Format:       "json",
//	    FailFast:     true,
//	    Filename:     "data.csv",
//	}
//	results, err := csvlinter.LintAdvanced(csvReader, opts, &buf)
//
// Lint and LintWithSchema return a results struct with File, TotalRows, Errors,
// Warnings, Duration, Valid, and SchemaUsed. LintAdvanced returns (*Results, error)
// so callers can check results.Valid; use it when you need to write formatted output
// to a writer or file. All functions read from r until EOF or error.
//
// Benchmarks and tool comparison (csvkit, csvlint): https://github.com/csvlinter/csvlinter
package csvlinter
