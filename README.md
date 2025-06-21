# csvlinter

A modern, streaming-first CSV validator with JSON Schema support. Validates structure, content, and encoding of CSV files — built for CI, CLI, and editor integration.

## Features

- **Streaming Validation**: Processes large CSV files efficiently without loading everything into memory
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
# Validate a CSV file
csvlinter validate data.csv

# Validate with custom delimiter
csvlinter validate data.csv --delimiter ";"

# Validate with JSON Schema
csvlinter validate data.csv --schema schema.json
```

### Output Options

```bash
# Pretty output (default)
csvlinter validate data.csv --format pretty

# JSON output
csvlinter validate data.csv --format json

# Save results to file
csvlinter validate data.csv --output results.json --format json
```

### CI/CD Integration

```bash
# Fail-fast mode for CI
csvlinter validate data.csv --fail-fast

# Exit with error code on validation failure
csvlinter validate data.csv || exit 1
```

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

## Examples

### Valid CSV
```csv
name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com
```

### Invalid CSV (missing column)
```csv
name,age
John Doe,30
Jane Smith,25,extra-column
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
