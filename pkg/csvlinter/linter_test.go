package csvlinter

import (
	"bytes"
	"os"
	"testing"

	"github.com/csvlinter/csvlinter/internal/schema"
)

func readFileToBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func TestLint(t *testing.T) {
	testCases := []struct {
		name          string
		filePath      string
		delimiter     string
		expectedErrs  int
		expectSuccess bool
	}{
		{
			name:          "Valid CSV",
			filePath:      "../../testdata/valid_sample.csv",
			delimiter:     ",",
			expectedErrs:  0,
			expectSuccess: true,
		},
		{
			name:          "Invalid CSV with mismatched columns",
			filePath:      "../../testdata/invalid_colons.csv",
			delimiter:     ",",
			expectedErrs:  1,
			expectSuccess: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := readFileToBytes(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}
			reader := bytes.NewReader(data)
			results, err := Lint(reader, tc.filePath, tc.delimiter)

			if tc.expectSuccess && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tc.expectSuccess && err != nil {
				t.Logf("Got expected operational error: %v", err)
				return
			}

			if results == nil && !tc.expectSuccess {
				t.Errorf("Expected results object even on validation failure, but got nil")
				return
			}

			if results != nil && len(results.Errors) != tc.expectedErrs {
				t.Errorf("Expected %d errors, but got %d. Errors: %v", tc.expectedErrs, len(results.Errors), results.Errors)
			}
		})
	}
}

func TestLintWithSchema(t *testing.T) {
	csvPath := "../../testdata/valid_sample.csv"
	schemaPath := "../../testdata/csvlinter.schema.json"
	csvData, err := readFileToBytes(csvPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}
	schemaData, err := readFileToBytes(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	schemaValidator, err := schema.NewValidatorFromReader(bytes.NewReader(schemaData))
	if err != nil {
		t.Fatalf("Failed to create schema validator: %v", err)
	}

	results, err := LintWithSchema(bytes.NewReader(csvData), csvPath, ",", schemaValidator)
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if results == nil {
		t.Fatalf("Expected results, got nil")
	}
	if !results.Valid {
		t.Errorf("Expected valid results, got errors: %v", results.Errors)
	}
}
