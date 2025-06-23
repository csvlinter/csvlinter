package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"csvlinter/internal/validator"

	"github.com/mattn/go-isatty"
)

// Reporter handles output formatting
type Reporter struct {
	format     string
	outputPath string
	isTerminal bool
}

// New creates a new reporter
func New(format, outputPath string) *Reporter {
	return &Reporter{
		format:     format,
		outputPath: outputPath,
		isTerminal: isatty.IsTerminal(os.Stdout.Fd()),
	}
}

// Report outputs the validation results
func (r *Reporter) Report(results *validator.Results, writer io.Writer) error {
	if results == nil {
		return fmt.Errorf("results cannot be nil")
	}

	var output string
	var err error

	switch r.format {
	case "json":
		output, err = r.formatJSON(results)
	case "pretty":
		output, err = r.formatPretty(results)
	default:
		return fmt.Errorf("unsupported format: %s", r.format)
	}

	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if writer == nil {
		writer = os.Stdout
	}

	// Write to file or stdout
	if r.outputPath != "" {
		if err := os.WriteFile(r.outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		if _, err := fmt.Fprint(writer, output); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	}

	return nil
}

// formatJSON formats results as JSON
func (r *Reporter) formatJSON(results *validator.Results) (string, error) {
	jsonBytes, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes) + "\n", nil
}

// formatPretty formats results for human reading
func (r *Reporter) formatPretty(results *validator.Results) (string, error) {
	var sb strings.Builder

	// Header
	if r.isTerminal {
		sb.WriteString("\033[1m") // Bold
	}
	sb.WriteString("CSV Validation Results\n")
	sb.WriteString("=====================\n")
	if r.isTerminal {
		sb.WriteString("\033[0m") // Reset
	}

	// File info
	sb.WriteString(fmt.Sprintf("File: %s\n", results.File))
	sb.WriteString(fmt.Sprintf("Total Rows: %d\n", results.TotalRows))
	sb.WriteString(fmt.Sprintf("Duration: %s\n", results.Duration))
	sb.WriteString(fmt.Sprintf("Schema Used: %t\n", results.SchemaUsed))

	// Status
	sb.WriteString("\nStatus: ")
	if results.Valid {
		if r.isTerminal {
			sb.WriteString("\033[32m") // Green
		}
		sb.WriteString("✓ VALID\n")
		if r.isTerminal {
			sb.WriteString("\033[0m") // Reset
		}
	} else {
		if r.isTerminal {
			sb.WriteString("\033[31m") // Red
		}
		sb.WriteString("✗ INVALID\n")
		if r.isTerminal {
			sb.WriteString("\033[0m") // Reset
		}
	}

	// Errors
	if len(results.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(results.Errors)))
		for i, err := range results.Errors {
			if r.isTerminal {
				sb.WriteString("\033[31m") // Red
			}
			sb.WriteString(fmt.Sprintf("  %d. Line %d", i+1, err.LineNumber))
			if err.Field != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", err.Field))
			}
			sb.WriteString(fmt.Sprintf(": %s", err.Message))
			if err.Value != "" {
				sb.WriteString(fmt.Sprintf(" (value: %q)", err.Value))
			}
			sb.WriteString(fmt.Sprintf(" [%s]", err.Type))
			sb.WriteString("\n")
			if r.isTerminal {
				sb.WriteString("\033[0m") // Reset
			}
		}
	}

	// Warnings
	if len(results.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("\nWarnings (%d):\n", len(results.Warnings)))
		for i, warning := range results.Warnings {
			if r.isTerminal {
				sb.WriteString("\033[33m") // Yellow
			}
			sb.WriteString(fmt.Sprintf("  %d. Line %d", i+1, warning.LineNumber))
			if warning.Field != "" && warning.Field != "row" {
				sb.WriteString(fmt.Sprintf(" (%s)", warning.Field))
			}
			sb.WriteString(fmt.Sprintf(": %s", warning.Message))
			if warning.Value != "" {
				sb.WriteString(fmt.Sprintf(" (value: %q)", warning.Value))
			}
			sb.WriteString(fmt.Sprintf(" [%s]", warning.Type))
			sb.WriteString("\n")
			if r.isTerminal {
				sb.WriteString("\033[0m") // Reset
			}
		}
	}

	// Summary
	sb.WriteString("\n")
	if results.Valid {
		if r.isTerminal {
			sb.WriteString("\033[32m") // Green
		}
		sb.WriteString("✓ All validations passed!\n")
		if r.isTerminal {
			sb.WriteString("\033[0m") // Reset
		}
	} else {
		if r.isTerminal {
			sb.WriteString("\033[31m") // Red
		}
		sb.WriteString(fmt.Sprintf("✗ Found %d error(s)\n", len(results.Errors)))
		if r.isTerminal {
			sb.WriteString("\033[0m") // Reset
		}
	}

	return sb.String(), nil
}
