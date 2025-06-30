package csvlinter

import (
	"path/filepath"
	"testing"
)

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
			absPath, err := filepath.Abs(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to get absolute path: %v", err)
			}

			results, err := Lint(absPath, tc.delimiter)

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
	absCSV, err := filepath.Abs(csvPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}
	absSchema, err := filepath.Abs(schemaPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	results, err := LintWithSchema(absCSV, absSchema, ",")
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
