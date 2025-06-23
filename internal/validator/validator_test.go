package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator(t *testing.T) {
	testCases := []struct {
		name          string
		filePath      string
		expectedErrs  int
		expectSuccess bool
	}{
		{
			name:          "Valid CSV",
			filePath:      "../../testdata/valid_sample.csv",
			expectedErrs:  0,
			expectSuccess: true,
		},
		{
			name:          "Invalid CSV with mismatched columns",
			filePath:      "../../testdata/invalid_colons.csv",
			expectedErrs:  1,
			expectSuccess: false,
		},
		{
			name:          "Invalid CSV with bad quotes",
			filePath:      "../../testdata/invalid_bad_quotes.csv",
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

			// Open the file
			file, err := os.Open(absPath)
			if err != nil {
				if tc.expectSuccess {
					t.Fatalf("Failed to open file: %v", err)
				}
				return
			}
			defer file.Close()

			validator := New(file, absPath, ",", nil, true)
			results, err := validator.Validate()

			if tc.expectSuccess && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}

			if !tc.expectSuccess && err != nil {
				// This covers operational errors like file not found, which is expected for the nonexistent file case.
				// For validation errors, `err` should be nil.
				t.Logf("Got expected operational error: %v", err)
				return
			}

			if results == nil && !tc.expectSuccess {
				t.Errorf("Expected results object even on validation failure, but got nil")
				return
			}

			if len(results.Errors) != tc.expectedErrs {
				t.Errorf("Expected %d errors, but got %d. Errors: %v", tc.expectedErrs, len(results.Errors), results.Errors)
			}
		})
	}
}
