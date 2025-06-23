package reporter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"csvlinter/internal/validator"
)

func TestReporter(t *testing.T) {
	// Create test results
	results := &validator.Results{
		File:      "test.csv",
		TotalRows: 3,
		Errors: []validator.Error{
			{
				LineNumber: 2,
				Field:      "email",
				Message:    "invalid email format",
				Value:      "not-an-email",
				Type:       "schema",
			},
			{
				LineNumber: 3,
				Field:      "row",
				Message:    "column count mismatch: expected 3, got 4",
				Type:       "structure",
			},
		},
		Warnings: []validator.Warning{
			{
				LineNumber: 1,
				Field:      "age",
				Message:    "value out of recommended range",
				Value:      "150",
				Type:       "schema",
			},
		},
		Duration:   "15.2ms",
		Valid:      false,
		SchemaUsed: true,
	}

	t.Run("JSON format to stdout", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("json", "")
		err := r.Report(results, &buf)
		if err != nil {
			t.Fatalf("Failed to generate JSON report: %v", err)
		}

		// Verify JSON output
		var decoded validator.Results
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		// Check fields
		if decoded.File != results.File {
			t.Errorf("Expected file %s, got %s", results.File, decoded.File)
		}
		if decoded.TotalRows != results.TotalRows {
			t.Errorf("Expected %d rows, got %d", results.TotalRows, decoded.TotalRows)
		}
		if len(decoded.Errors) != len(results.Errors) {
			t.Errorf("Expected %d errors, got %d", len(results.Errors), len(decoded.Errors))
		}
		if len(decoded.Warnings) != len(results.Warnings) {
			t.Errorf("Expected %d warnings, got %d", len(results.Warnings), len(decoded.Warnings))
		}
	})

	t.Run("Pretty format to stdout", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("pretty", "")
		err := r.Report(results, &buf)
		if err != nil {
			t.Fatalf("Failed to generate pretty report: %v", err)
		}

		output := buf.String()

		// Check for expected sections in order-independent way
		expectedSections := []string{
			"CSV Validation Results",
			"File: test.csv",
			"Total Rows: 3",
			"Duration: 15.2ms",
			"Schema Used: true",
			"Status: ✗ INVALID",
			"Errors (2):",
		}

		for _, section := range expectedSections {
			if !strings.Contains(output, section) {
				t.Errorf("Expected output to contain %q", section)
			}
		}

		// Check for error messages in any order
		errorMessages := []string{
			"Line 2 (email): invalid email format (value: \"not-an-email\") [schema]",
			"Line 3 (row): column count mismatch: expected 3, got 4 [structure]",
		}
		for _, msg := range errorMessages {
			if !strings.Contains(output, msg) {
				t.Errorf("Expected output to contain error message %q", msg)
			}
		}

		// Check for warning message
		warningMsg := "Line 1 (age): value out of recommended range (value: \"150\") [schema]"
		if !strings.Contains(output, warningMsg) {
			t.Errorf("Expected output to contain warning message %q", warningMsg)
		}

		// Check for summary
		if !strings.Contains(output, "✗ Found 2 error(s)") {
			t.Errorf("Expected output to contain error summary")
		}
	})

	t.Run("JSON format to file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := filepath.Join(tempDir, "results.json")

		r := New("json", outputPath)
		err := r.Report(results, nil)
		if err != nil {
			t.Fatalf("Failed to write JSON report to file: %v", err)
		}

		// Read and verify file contents
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}

		var decoded validator.Results
		if err := json.Unmarshal(content, &decoded); err != nil {
			t.Fatalf("Failed to parse JSON from file: %v", err)
		}

		// Check fields
		if decoded.File != results.File {
			t.Errorf("Expected file %s, got %s", results.File, decoded.File)
		}
		if decoded.TotalRows != results.TotalRows {
			t.Errorf("Expected %d rows, got %d", results.TotalRows, decoded.TotalRows)
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("invalid", "")
		err := r.Report(results, &buf)
		if err == nil {
			t.Error("Expected error for invalid format, got none")
		}
		if !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("Expected 'unsupported format' error, got: %v", err)
		}
	})

	t.Run("Invalid output file path", func(t *testing.T) {
		r := New("json", "/nonexistent/directory/results.json")
		err := r.Report(results, nil)
		if err == nil {
			t.Error("Expected error for invalid output path, got none")
		}
	})

	t.Run("Nil results", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("json", "")
		err := r.Report(nil, &buf)
		if err == nil {
			t.Error("Expected error for nil results, got none")
		}
	})
}

func TestReporterWithEmptyResults(t *testing.T) {
	results := &validator.Results{
		File:       "empty.csv",
		TotalRows:  0,
		Errors:     []validator.Error{},
		Warnings:   []validator.Warning{},
		Duration:   "1.2ms",
		Valid:      true,
		SchemaUsed: false,
	}

	t.Run("JSON format with empty results", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("json", "")
		err := r.Report(results, &buf)
		if err != nil {
			t.Fatalf("Failed to generate JSON report: %v", err)
		}

		var decoded validator.Results
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("Failed to parse JSON output: %v", err)
		}

		if !decoded.Valid {
			t.Error("Expected valid=true for empty results")
		}
		if len(decoded.Errors) != 0 {
			t.Errorf("Expected 0 errors, got %d", len(decoded.Errors))
		}
		if len(decoded.Warnings) != 0 {
			t.Errorf("Expected 0 warnings, got %d", len(decoded.Warnings))
		}
	})

	t.Run("Pretty format with empty results", func(t *testing.T) {
		var buf bytes.Buffer
		r := New("pretty", "")
		err := r.Report(results, &buf)
		if err != nil {
			t.Fatalf("Failed to generate pretty report: %v", err)
		}

		output := buf.String()
		expectedSections := []string{
			"CSV Validation Results",
			"File: empty.csv",
			"Total Rows: 0",
			"Duration: 1.2ms",
			"Schema Used: false",
			"Status: ✓ VALID",
			"✓ All validations passed!",
		}

		for _, section := range expectedSections {
			if !strings.Contains(output, section) {
				t.Errorf("Expected output to contain %q", section)
			}
		}

		unexpectedSections := []string{
			"Errors",
			"Warnings",
		}

		for _, section := range unexpectedSections {
			if strings.Contains(output, section) {
				t.Errorf("Output should not contain %q", section)
			}
		}
	})
}
