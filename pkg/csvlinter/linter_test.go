package csvlinter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
			logicalName := filepath.Base(tc.filePath)
			results, err := Lint(reader, logicalName, tc.delimiter)

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
	schemaPath := "../../testdata/schema.json"
	results, err := LintWithSchema(strings.NewReader(validCSVForSchema), "test.csv", ",", schemaPath)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
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
		_, err := LintAdvanced(bytes.NewReader(invalidData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(strings.NewReader(validCSVForSchema), opts, &buf)
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
		_, err := LintAdvanced(strings.NewReader(invalidCSVForSchema), opts, &buf)
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
		_, err := LintAdvanced(strings.NewReader(validCSVForSchema), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(semicolonData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(tabData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(invalidSampleData), optsFF, &bufFF)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var resultsFF map[string]interface{}
		_ = json.Unmarshal(bufFF.Bytes(), &resultsFF)
		errsFF := resultsFF["errors"].([]interface{})

		// Without fail-fast
		optsNF := Options{Delimiter: ",", FailFast: false, Format: "json", Filename: invalidSamplePath, SchemaPath: schemaPath}
		var bufNF bytes.Buffer
		_, err = LintAdvanced(bytes.NewReader(invalidSampleData), optsNF, &bufNF)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, nil)
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
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err == nil {
			t.Errorf("Expected error for invalid format, got nil")
		}
	})

	t.Run("Error: missing schema", func(t *testing.T) {
		opts := Options{Delimiter: ",", Format: "json", SchemaPath: "nonexistent.json", Filename: csvPath}
		var buf bytes.Buffer
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err == nil {
			t.Errorf("Expected error for missing schema, got nil")
		}
	})

	t.Run("Error: invalid delimiter (empty string)", func(t *testing.T) {
		opts := Options{Delimiter: "", Format: "json", Filename: csvPath}
		var buf bytes.Buffer
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Errorf("Should default to comma delimiter, got error: %v", err)
		}
	})
}

func TestLintAdvanced_InferSchema(t *testing.T) {
	csvContent := "id,name\n1,Alice\n2,Bob\n"
	dir := t.TempDir()
	csvPath := dir + "/data.csv"
	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	csvData, _ := os.ReadFile(csvPath)

	t.Run("InferSchema with no schema gives schema_used and schema_inferred", func(t *testing.T) {
		opts := Options{
			Delimiter:   ",",
			Format:      "json",
			Filename:    csvPath,
			InferSchema: true,
		}
		var buf bytes.Buffer
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if used, _ := results["schema_used"].(bool); !used {
			t.Errorf("expected schema_used true when InferSchema")
		}
		if inferred, _ := results["schema_inferred"].(bool); !inferred {
			t.Errorf("expected schema_inferred true when InferSchema")
		}
		if valid, _ := results["valid"].(bool); !valid {
			t.Errorf("expected valid true for valid CSV with inferred schema, errors: %v", results["errors"])
		}
	})

	t.Run("InferSchema with explicit SchemaPath uses file schema not inferred", func(t *testing.T) {
		schemaPath := "../../testdata/csvlinter.schema.json"
		opts := Options{
			Delimiter:   ",",
			Format:      "json",
			Filename:    csvPath,
			SchemaPath:  schemaPath,
			InferSchema: true,
		}
		var buf bytes.Buffer
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if used, _ := results["schema_used"].(bool); !used {
			t.Errorf("expected schema_used true")
		}
		if inferred, ok := results["schema_inferred"].(bool); ok && inferred {
			t.Errorf("expected schema_inferred false or absent when explicit schema provided")
		}
	})

	t.Run("invalid_sample_inferred_schema fails when age inferred as integer from sample", func(t *testing.T) {
		csvPath := "../../testdata/invalid_sample_inferred_schema.csv"
		csvData, err := os.ReadFile(csvPath)
		if err != nil {
			t.Fatalf("read testdata: %v", err)
		}
		opts := Options{
			Delimiter:   ",",
			Format:      "json",
			Filename:    csvPath,
			InferSchema: true,
		}
		var buf bytes.Buffer
		_, lintErr := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if lintErr != nil {
			t.Fatalf("LintAdvanced returned unexpected error: %v", lintErr)
		}
		var results map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
			t.Fatalf("invalid JSON output: %v\nout: %s", err, buf.String())
		}
		if inferred, _ := results["schema_inferred"].(bool); !inferred {
			t.Errorf("expected schema_inferred true")
		}
		if valid, _ := results["valid"].(bool); valid {
			t.Errorf("expected valid false for CSV with a non-integer age value")
		}
		rawErrors, _ := results["errors"].([]interface{})
		if len(rawErrors) == 0 {
			t.Fatal("expected at least one validation error, got none")
		}
		found := false
		for _, e := range rawErrors {
			em, _ := e.(map[string]interface{})
			field, _ := em["field"].(string)
			if field == "age" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected an error on field 'age', got errors: %v", rawErrors)
		}
	})

	t.Run("InferSchemaOutput writes schema file", func(t *testing.T) {
		outPath := dir + "/inferred.schema.json"
		opts := Options{
			Delimiter:         ",",
			Format:            "json",
			Filename:          csvPath,
			InferSchema:       true,
			InferSchemaOutput: outPath,
		}
		var buf bytes.Buffer
		_, err := LintAdvanced(bytes.NewReader(csvData), opts, &buf)
		if err != nil {
			t.Fatalf("LintAdvanced failed: %v", err)
		}
		written, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("read inferred schema file: %v", err)
		}
		var schemaObj map[string]interface{}
		if err := json.Unmarshal(written, &schemaObj); err != nil {
			t.Fatalf("inferred schema not valid JSON: %v", err)
		}
		if schemaObj["$schema"] != "http://json-schema.org/draft-07/schema#" {
			t.Errorf("expected draft-07 schema, got %v", schemaObj["$schema"])
		}
		props, _ := schemaObj["properties"].(map[string]interface{})
		if props == nil || props["id"] == nil || props["name"] == nil {
			t.Errorf("expected id and name in properties, got %v", schemaObj["properties"])
		}
	})
}
