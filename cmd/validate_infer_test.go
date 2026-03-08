package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/csvlinter/csvlinter/internal/validator"

	"github.com/urfave/cli/v2"
)

func runValidateWithInfer(t *testing.T, csvPath string, extraArgs []string, wantExit int, assertFn func(validator.Results)) {
	t.Helper()
	args := []string{"csvlinter", "validate", "--format", "json", "--infer-schema"}
	args = append(args, extraArgs...)
	args = append(args, csvPath)
	var stdout, stderr bytes.Buffer
	var exitCode int
	app := &cli.App{
		Commands: []*cli.Command{validateCommand},
		Writer:    &stdout,
		ErrWriter: &stderr,
		ExitErrHandler: func(c *cli.Context, err error) {
			if err != nil {
				if ec, ok := err.(cli.ExitCoder); ok {
					exitCode = ec.ExitCode()
				} else {
					exitCode = 1
				}
			}
		},
	}
	_ = app.Run(args)
	if exitCode != wantExit {
		t.Fatalf("want exit %d, got %d; stderr=%s", wantExit, exitCode, stderr.String())
	}
	var res validator.Results
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		t.Fatalf("invalid json output: %v\nstdout=%s", err, stdout.String())
	}
	assertFn(res)
}

func TestValidateCommand_InferSchema(t *testing.T) {
	t.Run("with --infer-schema and --format json", func(t *testing.T) {
		dir := t.TempDir()
		csvPath := filepath.Join(dir, "data.csv")
		if err := os.WriteFile(csvPath, []byte("id,name\n1,Alice\n2,Bob"), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		runValidateWithInfer(t, csvPath, nil, 0, func(res validator.Results) {
			if !res.SchemaUsed {
				t.Errorf("expected schema_used true when using --infer-schema")
			}
			if !res.SchemaInferred {
				t.Errorf("expected schema_inferred true when using --infer-schema")
			}
			if len(res.Errors) != 0 {
				t.Errorf("expected 0 errors for valid CSV, got %d", len(res.Errors))
			}
			if !res.Valid {
				t.Errorf("expected valid true")
			}
		})
	})

	t.Run("explicit schema wins over --infer-schema", func(t *testing.T) {
		dir := t.TempDir()
		csvPath := filepath.Join(dir, "data.csv")
		schemaPath := filepath.Join(dir, "data.schema.json")
		if err := os.WriteFile(csvPath, []byte("id,name\n1,Alice"), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		if err := os.WriteFile(schemaPath, []byte(`{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","properties":{"id":{"type":"integer"},"name":{"type":"string"}},"required":["id","name"],"additionalProperties":false}`), 0o644); err != nil {
			t.Fatalf("write schema: %v", err)
		}
		runValidateWithInfer(t, csvPath, []string{"--schema", schemaPath}, 0, func(res validator.Results) {
			if !res.SchemaUsed {
				t.Errorf("expected schema_used true")
			}
			if res.SchemaInferred {
				t.Errorf("expected schema_inferred false when explicit schema provided")
			}
		})
	})

	t.Run("invalid CSV with inferred schema exits non-zero", func(t *testing.T) {
		dir := t.TempDir()
		csvPath := filepath.Join(dir, "bad.csv")
		if err := os.WriteFile(csvPath, []byte("id,name\n1,Alice\n2,Bob\n3,Charlie,extra"), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		var stdout, stderr bytes.Buffer
		var exitCode int
		app := &cli.App{
			Commands: []*cli.Command{validateCommand},
			Writer:    &stdout,
			ErrWriter: &stderr,
			ExitErrHandler: func(c *cli.Context, err error) {
				if err != nil {
					if ec, ok := err.(cli.ExitCoder); ok {
						exitCode = ec.ExitCode()
					} else {
						exitCode = 1
					}
				}
			},
		}
		_ = app.Run([]string{"csvlinter", "validate", "--format", "json", "--infer-schema", csvPath})
		if exitCode != 1 {
			t.Fatalf("expected exit 1 for invalid CSV, got %d", exitCode)
		}
		var res validator.Results
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			t.Fatalf("invalid json: %v", err)
		}
		if res.Valid {
			t.Errorf("expected valid false")
		}
		if len(res.Errors) == 0 {
			t.Errorf("expected at least one error")
		}
		if !res.SchemaInferred {
			t.Errorf("expected schema_inferred true")
		}
	})

	t.Run("--infer-schema-output writes schema file", func(t *testing.T) {
		dir := t.TempDir()
		csvPath := filepath.Join(dir, "data.csv")
		outSchema := filepath.Join(dir, "inferred.schema.json")
		if err := os.WriteFile(csvPath, []byte("a,b\n1,foo"), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		runValidateWithInfer(t, csvPath, []string{"--infer-schema-output", outSchema}, 0, func(res validator.Results) {
			if !res.SchemaInferred {
				t.Errorf("expected schema_inferred true")
			}
		})
		body, err := os.ReadFile(outSchema)
		if err != nil {
			t.Fatalf("read inferred schema: %v", err)
		}
		var schema map[string]interface{}
		if err := json.Unmarshal(body, &schema); err != nil {
			t.Fatalf("inferred schema not JSON: %v", err)
		}
		if schema["$schema"] != "http://json-schema.org/draft-07/schema#" {
			t.Errorf("expected draft-07 schema")
		}
		props, _ := schema["properties"].(map[string]interface{})
		if props["a"] == nil || props["b"] == nil {
			t.Errorf("expected a and b in properties")
		}
	})
}
