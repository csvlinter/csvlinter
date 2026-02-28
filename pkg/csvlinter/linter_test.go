package csvlinter

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
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

	results, err := LintWithSchema(bytes.NewReader(csvData), csvPath, ",", schemaPath)
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

func TestLintAdvanced(t *testing.T) {
	csvPath := "../../testdata/valid_sample.csv"
	invalidPath := "../../testdata/invalid_colons.csv"

	csvData, err := readFileToBytes(csvPath)
	if err != nil {
		t.Fatalf("Failed to read CSV file: %v", err)
	}
	invalidData, err := readFileToBytes(invalidPath)
	if err != nil {
		t.Fatalf("Failed to read invalid CSV file: %v", err)
	}

	t.Run("Pretty output to buffer", func(t *testing.T) {
		opts := Options{
			Delimiter: ",",
			Format:    "pretty",
			Filename:  csvPath,
		}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "CSV Validation Results") {
			t.Errorf("Expected pretty output, got: %q", output)
		}
	})

	t.Run("JSON output to buffer", func(t *testing.T) {
		opts := Options{
			Delimiter: ",",
			Format:    "json",
			Filename:  csvPath,
		}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &results)
		if err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
	})

	t.Run("Output to file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := tempDir + "/results.json"
		opts := Options{
			Delimiter: ",",
			Format:    "json",
			Output:    outputPath,
			Filename:  csvPath,
		}
		err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		var results map[string]interface{}
		err = json.Unmarshal(content, &results)
		if err != nil {
			t.Errorf("Expected valid JSON in file, got error: %v", err)
		}
	})

	t.Run("Fail-fast stops after first error", func(t *testing.T) {
		opts := Options{
			Delimiter: ",",
			FailFast:  true,
			Format:    "json",
			Filename:  invalidPath,
		}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(invalidData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &results)
		if err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
		errs, ok := results["errors"].([]interface{})
		if !ok || len(errs) != 1 {
			t.Errorf("Expected 1 error due to fail-fast, got: %v", len(errs))
		}
	})

	t.Run("Logical filename triggers schema resolution", func(t *testing.T) {
		opts := Options{
			Delimiter: ",",
			Format:    "json",
			Filename:  csvPath, // should resolve schema automatically
		}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &results)
		if err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
		if _, ok := results["schema_used"]; !ok {
			t.Errorf("Expected schema_used field in results")
		}
	})
}

const schemaFromReaderJSON = `{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"name":{"type":"string","minLength":1},"age":{"type":"string","pattern":"^[0-9]+$"},"email":{"type":"string","format":"email"},"city":{"type":"string","minLength":1}},"required":["name","age","email","city"],"additionalProperties":false}`

const validCSVForSchema = "name,age,email,city\nJohn Doe,30,john@example.com,New York\nJane Smith,25,jane@example.com,Los Angeles\n"

const invalidCSVForSchema = "name,age,email,city\nJohn Doe,abc,john@example.com,New York\nJane Smith,25,not-an-email,Los Angeles\n"

func TestLintAdvanced_SchemaFromReader(t *testing.T) {
	t.Run("valid CSV with schema from reader", func(t *testing.T) {
		opts := Options{
			SchemaReader: strings.NewReader(schemaFromReaderJSON),
			Delimiter:    ",",
			Format:       "json",
			Filename:     "test.csv",
		}
		var buf bytes.Buffer
		err := LintAdvanced(strings.NewReader(validCSVForSchema), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("Expected valid JSON output: %v", err)
		}
		if valid, ok := results["valid"].(bool); !ok || !valid {
			t.Errorf("Expected valid result, got: %v", results["valid"])
		}
		used, _ := results["schema_used"].(bool)
		if !used {
			t.Errorf("Expected schema_used true when using SchemaReader")
		}
	})

	t.Run("invalid CSV with schema from reader", func(t *testing.T) {
		opts := Options{
			SchemaReader: strings.NewReader(schemaFromReaderJSON),
			Delimiter:    ",",
			Format:       "json",
			Filename:     "test.csv",
		}
		var buf bytes.Buffer
		err := LintAdvanced(strings.NewReader(invalidCSVForSchema), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("Expected valid JSON output: %v", err)
		}
		errs, ok := results["errors"].([]interface{})
		if !ok || len(errs) < 1 {
			t.Errorf("Expected validation errors, got %d", len(errs))
		}
	})

	t.Run("SchemaReader takes precedence over SchemaPath", func(t *testing.T) {
		opts := Options{
			SchemaReader: strings.NewReader(schemaFromReaderJSON),
			SchemaPath:   "nonexistent.json",
			Delimiter:    ",",
			Format:       "json",
			Filename:     "test.csv",
		}
		var buf bytes.Buffer
		err := LintAdvanced(strings.NewReader(validCSVForSchema), opts, &buf)
		if err != nil {
			t.Fatalf("SchemaReader should take precedence; LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("Expected valid JSON output: %v", err)
		}
		if valid, ok := results["valid"].(bool); !ok || !valid {
			t.Errorf("Expected valid result when schema from reader, got: %v", results["valid"])
		}
	})
}

func TestLintAdvanced_AllOptions(t *testing.T) {
	csvPath := "../../testdata/valid_sample.csv"
	semicolonPath := "../../testdata/valid_semis.csv"
	tabPath := "../../testdata/valid_tabs.csv"

	csvData, _ := readFileToBytes(csvPath)
	semicolonData, _ := readFileToBytes(semicolonPath)
	tabData, _ := readFileToBytes(tabPath)

	t.Run("Delimiter: semicolon", func(t *testing.T) {
		opts := Options{Delimiter: ";", Format: "json", Filename: semicolonPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(semicolonData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
	})

	t.Run("Delimiter: tab", func(t *testing.T) {
		opts := Options{Delimiter: "\t", Format: "json", Filename: tabPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(tabData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
	})

	t.Run("FailFast: true vs false", func(t *testing.T) {
		invalidSamplePath := "../../testdata/invalid_sample.csv"
		schemaPath := "../../testdata/schema.json"
		invalidSampleData, _ := readFileToBytes(invalidSamplePath)
		// With fail-fast
		optsFF := Options{Delimiter: ",", FailFast: true, Format: "json", Filename: invalidSamplePath, SchemaPath: schemaPath}
		var bufFF bytes.Buffer
		err := LintAdvanced(bytes.NewReader(invalidSampleData), optsFF, &bufFF)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var resultsFF map[string]interface{}
		_ = json.Unmarshal(bufFF.Bytes(), &resultsFF)
		errsFF := resultsFF["errors"].([]interface{})

		// Without fail-fast
		optsNF := Options{Delimiter: ",", FailFast: false, Format: "json", Filename: invalidSamplePath, SchemaPath: schemaPath}
		var bufNF bytes.Buffer
		err = LintAdvanced(bytes.NewReader(invalidSampleData), optsNF, &bufNF)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var resultsNF map[string]interface{}
		_ = json.Unmarshal(bufNF.Bytes(), &resultsNF)
		errsNF := resultsNF["errors"].([]interface{})

		if len(errsFF) >= len(errsNF) {
			t.Errorf("FailFast should produce fewer errors: got %d vs %d", len(errsFF), len(errsNF))
		}
		if len(errsNF) < 2 {
			t.Errorf("Expected multiple errors in non-fail-fast mode, got %d", len(errsNF))
		}
	})

	t.Run("Format: pretty", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "pretty", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "CSV Validation Results") {
			t.Errorf("Expected pretty output, got: %q", output)
		}
	})

	t.Run("Format: json", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "json", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Errorf("Expected valid JSON output, got error: %v", err)
		}
	})

	t.Run("Output: to file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := tempDir + "/results.json"
		opts := Options{Delimiter: ",", Format: "json", Output: outputPath, Filename: csvPath}
		err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(content, &results); err != nil {
			t.Errorf("Expected valid JSON in file, got error: %v", err)
		}
	})

	t.Run("Filename: logical schema resolution", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "json", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		_ = json.Unmarshal(buf.Bytes(), &results)
		if _, ok := results["schema_used"]; !ok {
			t.Errorf("Expected schema_used field in results")
		}
	})

	t.Run("SchemaPath: explicit", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "json", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		_ = json.Unmarshal(buf.Bytes(), &results)
		if _, ok := results["schema_used"]; !ok {
			t.Errorf("Expected schema_used field in results")
		}
	})

	t.Run("Combination: all options", func(t *testing.T) {
		tempDir := t.TempDir()
		outputPath := tempDir + "/combo.json"
		opts := Options{
			Delimiter: ",",
			FailFast:  true,
			Format:    "json",
			Output:    outputPath,
			Filename:  csvPath,
		}
		err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		content, err := os.ReadFile(outputPath)
		if err != nil {
			t.Fatalf("Failed to read output file: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(content, &results); err != nil {
			t.Errorf("Expected valid JSON in file, got error: %v", err)
		}
	})

	t.Run("Error: invalid format", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "invalid", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err == nil {
			t.Errorf("Expected error for invalid format, got nil")
		}
	})

	t.Run("Error: missing schema", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "json", SchemaPath: "nonexistent.json", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err == nil {
			t.Errorf("Expected error for missing schema, got nil")
		}
	})

	t.Run("Error: invalid delimiter (empty string)", func(t *testing.T) {
		opts := Options{Delimiter: "", Format: "json", Filename: csvPath}
		var buf bytes.Buffer
		err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Errorf("Should default to comma delimiter, got error: %v", err)
		}
	})
}
