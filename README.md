# csvlinter
## ⚡ The fastest streaming CSV validator (up to 270x faster than similar tools)

Blazing-fast, streaming-first CSV validator with JSON Schema support. Validates structure, content, and encoding of CSV files — built for CI, CLI, API and editor integration.

> **Performance:** On a 1,000,000-row CSV:
> - **Structure-only validation:** csvlinter completes in **0.23s**, compared to **8.83s** for csvkit (`csvstat`) and **30s** for csvlint — that's **38x faster than csvkit** and **130x faster than csvlint**.
> - **Schema validation:** csvlinter is **60–270x faster** than csvkit or csvlint, while using significantly less memory.
[See Benchmarks & Tool Comparison](./benchmark.md)

## Features

- **Streaming validation**: Processes large CSV files efficiently without loading everything into memory
- **STDIN support**: Process data directly from standard input
- **JSON schema support**: Validate CSV data against JSON Schema specifications
- **UTF-8 encoding validation**: Ensures proper character encoding
- **Flexible delimiters**: Support for custom delimiter characters
- **Multiple output formats**: Pretty terminal output and structured JSON
- **Fail-fast mode**: Stop validation on first error for CI/CD integration
- **Cross-platform**: Works on Windows, macOS, and Linux
- **RFC 4180 compliance**: Uses Go's standard CSV parser for robust, standards-based CSV handling

## Installation

### Using Homebrew

```bash
brew tap csvlinter/tap && brew install csvlinter
```

### From source

```bash
git clone https://github.com/csvlinter/csvlinter.git
cd csvlinter
go build -o csvlinter
```

### Using Go

```bash
go install github.com/csvlinter/csvlinter@latest
```

## Usage

### Basic validation

```bash
# Validate a CSV file (auto-detects schema if not provided)
csvlinter validate data.csv

# Validate data from STDIN (use "-" as input)
cat data.csv | csvlinter validate -

# Validate STDIN and provide a logical filename for schema resolution
cat data.csv | csvlinter validate - --filename data.csv

# Validate with custom delimiter (short flag)
csvlinter validate data.csv -d ";"

# Validate STDIN with custom delimiter
cat data.csv | csvlinter validate - -d ";"

# Validate with JSON Schema (short flag)
csvlinter validate data.csv -s schema.json

# Validate STDIN with JSON Schema
cat data.csv | csvlinter validate - -s schema.json
```

> **Schema fallback:**
> If `--schema`/`-s` is not set, csvlinter will look for `<csv>.schema.json` or `csvlinter.schema.json` in the same or parent directories automatically.

### Output options

```bash
# Pretty output (default)
csvlinter validate data.csv --format pretty

# JSON output (short flag)
csvlinter validate data.csv -f json

# Save results to file (short flag)
csvlinter validate data.csv -o results.json -f json
```

> **Output File:**
> If `--output`/`-o` is set, results are written to the specified file. Otherwise, output is printed to the terminal.

### CI/CD integration

```bash
# Fail-fast mode for CI
csvlinter validate data.csv --fail-fast

# csvlinter exits with code 1 when validation fails, so CI pipelines fail automatically
csvlinter validate data.csv
```

### STDIN support

csvlinter supports reading data from standard input using `-` as the input file:

```bash
# Basic STDIN validation
cat data.csv | csvlinter validate -

# Provide a logical filename for schema resolution and reporting
cat data.csv | csvlinter validate - --filename data.csv

# Pipeline integration
generate-csv | csvlinter validate - --fail-fast

# Process with custom options
cat data.csv | csvlinter validate - -d ";" -s schema.json
```

> **Size limit:**
> STDIN input is limited to 10MB by default. Use `--max-size` flag to adjust this limit (e.g., `--max-size 50MB`).

> **Logical filename:**
> Use `--filename` to provide a logical filename for schema resolution and reporting when reading from STDIN. This enables automatic schema lookup as if you were validating a file with that name.

## JSON schema support

Create a JSON schema file to validate your CSV data:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1
    },
    "age": {
      "type": "string",
      "pattern": "^[0-9]+$"
    },
    "email": {
      "type": "string",
      "format": "email"
    }
  },
  "required": ["name", "age", "email"]
}
```

Then validate your CSV:

```bash
csvlinter validate users.csv --schema user-schema.json
```

### Schema resolution:

When you do not specify a schema file with `--schema` or `-s`, csvlinter will attempt to automatically resolve the schema by searching for a file named `<csv>.schema.json` (where `<csv>` is your CSV filename) in the same directory as your CSV file. If not found, it will look for a file named `csvlinter.schema.json` in the same directory and then recursively in each parent directory until it reaches the root.

> **Note for STDIN:**
> When using STDIN input (`-`), automatic schema resolution is disabled unless you provide a logical filename with `--filename`. In that case, schema resolution works as if you were validating a file with that name. You must still explicitly provide a schema file using the `--schema` or `-s` flag if no schema is found.

## Examples

### Valid CSV
```csv
name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com
```

Example validation:
```bash
# File input
csvlinter validate users.csv

# STDIN input
echo "name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com" | csvlinter validate -
```

### Invalid CSV (missing column)
```csv
name,age
John Doe,30
Jane Smith,25,extra-column
```

Example validation:
```bash
# File input
csvlinter validate invalid.csv

# STDIN input
cat invalid.csv | csvlinter validate - --fail-fast
```

## Output formats

### Pretty output
```
CSV Validation Results
=====================
File: data.csv
Total Rows: 100
Duration: 15.2ms
Schema Used: true

Status: ✗ INVALID

Errors (2):
  1. Line 3 (email): invalid email format (value: "invalid-email") [schema]
  2. Line 5: column count mismatch: expected 3, got 4 [structure]

✗ Found 2 error(s)
```

### JSON output
```json
{
  "file": "data.csv",
  "total_rows": 100,
  "errors": [
    {
      "line_number": 3,
      "field": "email",
      "message": "invalid email format",
      "value": "invalid-email",
      "type": "schema"
    }
  ],
  "warnings": [],
  "duration": "15.2ms",
  "valid": false,
  "schema_used": true
}
```

## Error types

- **structure**: CSV format issues (wrong column count, malformed rows)
- **schema**: JSON Schema validation failures
- **encoding**: UTF-8 encoding problems

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for how to run tests, open PRs, and use Conventional Commits. By participating, you agree to the [Code of Conduct](CODE_OF_CONDUCT.md).

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Commit message guidelines (conventional commits)

This project uses [semantic-release](https://semantic-release.gitbook.io/) for automated versioning and changelog generation. To ensure your contributions are included in releases, **please follow the [Conventional Commits](https://www.conventionalcommits.org/) specification** for your commit messages:

- **feat:** A new feature
- **fix:** A bug fix
- **docs:** Documentation only changes
- **style:** Changes that do not affect the meaning of the code (white-space, formatting, etc)
- **refactor:** A code change that neither fixes a bug nor adds a feature
- **perf:** A code change that improves performance
- **test:** Adding or correcting tests
- **chore:** Changes to the build process or auxiliary tools

**Examples:**
```
feat: add streaming support for stdin
fix: handle empty CSV rows gracefully
docs: update usage examples in README
```

See the [Conventional Commits documentation](https://www.conventionalcommits.org/) for more details.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Public API

The public API is available in `pkg/csvlinter` and supports validating CSV data from any `io.Reader` (no need to write files to disk).

### Example usage

```go
import (
    "os"
    "bytes"
    "strings"
    "github.com/csvlinter/csvlinter/pkg/csvlinter"
)

// Validate CSV without schema (basic)
f, _ := os.Open("file.csv")
results, err := csvlinter.Lint(f, "file.csv", ",")

// Validate CSV with schema (basic)
csvFile, _ := os.Open("file.csv")
results, err := csvlinter.LintWithSchema(csvFile, "file.csv", ",", "schema.json")

// Validate CSV with schema from memory/stream (no schema file on disk)
schemaJSON := `{"type":"object","properties":{"id":{"type":"integer"},"name":{"type":"string"}},"required":["id","name"]}`
optsInMem := csvlinter.Options{
    SchemaReader: strings.NewReader(schemaJSON),
    Delimiter:    ",",
    Format:       "json",
    Filename:     "data.csv",
}
err = csvlinter.LintAdvanced(strings.NewReader(csvContent), optsInMem, &buf)

// Advanced: Use Options struct for full control
opts := csvlinter.Options{
    Delimiter:   ";",                // Custom delimiter
    FailFast:    true,               // Stop after first error
    Format:      "json",             // Output format: "pretty" or "json"
    Output:      "results.json",     // Output file (leave empty for stdout/writer)
    Filename:    "data.csv",         // Logical filename for schema resolution
    SchemaPath:  "schema.json",      // Optional: explicit schema file path
}
f2, _ := os.Open("file.csv")
var buf bytes.Buffer
err = csvlinter.LintAdvanced(f2, opts, &buf) // Output written to buf if Output is empty
// If opts.Output is set, results are written to that file.
```

- `Lint(r io.Reader, name string, delimiter string)`
- `LintWithSchema(r io.Reader, name string, delimiter string, schemaPath string)`
- `LintAdvanced(r io.Reader, opts Options, writer io.Writer)`

CSV input can be any stream (file, network, in-memory, etc.). Schema can be supplied the same way via `Options.SchemaReader` (e.g. `strings.NewReader(schemaJSON)`), or from a file path with `Options.SchemaPath` or automatic resolution from `Options.Filename`.

## RFC 4180 Compliance

csvlinter uses Go's standard `encoding/csv` parser, which is designed to be compatible with [RFC 4180](https://datatracker.ietf.org/doc/html/rfc4180), the common format for CSV files. This ensures robust handling of quoted fields, embedded newlines, and delimiter rules as described in the RFC.


