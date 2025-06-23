# csvlinter

A modern, streaming-first CSV validator with JSON Schema support. Validates structure, content, and encoding of CSV files — built for CI, CLI, and editor integration.

## Features

- **Streaming Validation**: Processes large CSV files efficiently without loading everything into memory
- **STDIN Support**: Process data directly from standard input
- **JSON Schema Support**: Validate CSV data against JSON Schema specifications
- **UTF-8 Encoding Validation**: Ensures proper character encoding
- **Flexible Delimiters**: Support for custom delimiter characters
- **Multiple Output Formats**: Pretty terminal output and structured JSON
- **Fail-Fast Mode**: Stop validation on first error for CI/CD integration
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Installation

### From Source

```bash
git clone https://github.com/yourusername/csvlinter.git
cd csvlinter
go build -o csvlinter
```

### Using Go

```bash
go install github.com/yourusername/csvlinter@latest
```

## Usage

### Basic Validation

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

> **Schema Fallback:**
> If `--schema`/`-s` is not set, csvlinter will look for `<csv>.schema.json` or `csvlinter.schema.json` in the same or parent directories automatically.

### Output Options

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

### CI/CD Integration

```bash
# Fail-fast mode for CI
csvlinter validate data.csv --fail-fast

# Exit with error code on validation failure
csvlinter validate data.csv || exit 1
```

### STDIN Support

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

> **Size Limit:**
> STDIN input is limited to 10MB by default. Use `--max-size` flag to adjust this limit (e.g., `--max-size 50MB`).

> **Logical Filename:**
> Use `--filename` to provide a logical filename for schema resolution and reporting when reading from STDIN. This enables automatic schema lookup as if you were validating a file with that name.

## JSON Schema Support

Create a JSON Schema file to validate your CSV data:

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

### Schema Resolution:

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

## Output Formats

### Pretty Output
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

### JSON Output
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

## Error Types

- **structure**: CSV format issues (wrong column count, malformed rows)
- **schema**: JSON Schema validation failures
- **encoding**: UTF-8 encoding problems

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Commit Message Guidelines (Conventional Commits)

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