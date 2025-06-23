package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"csvlinter/internal/validator"

	"github.com/urfave/cli/v2"
)

// validateSchemaCases verifies that the validate command can automatically
// discover schemas when --schema is not provided, using the same fallback
// rules implemented in internal/schema.ResolveSchema.
func TestValidateCommand_SchemaResolution(t *testing.T) {
	mkTempCSV := func(dir, name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write csv: %v", err)
		}
		return p
	}
	mkTempSchema := func(path string) {
		if err := os.WriteFile(path, []byte(`{"type":"object"}`), 0o644); err != nil {
			t.Fatalf("write schema: %v", err)
		}
	}

	t.Run("colocated filename.schema.json", func(t *testing.T) {
		dir := t.TempDir()
		csv := mkTempCSV(dir, "data.csv", "id,name\n1,Alice")
		schema := filepath.Join(dir, "data.schema.json")
		mkTempSchema(schema)
		runValidateAndAssertJSON(t, csv, 0, func(res validator.Results) {
			if len(res.Errors) != 0 {
				t.Errorf("expected 0 errors, got %d", len(res.Errors))
			}
		})
	})

	t.Run("csvlinter.schema.json in same dir", func(t *testing.T) {
		dir := t.TempDir()
		csv := mkTempCSV(dir, "file.csv", "id,name\n1,Bob")
		schema := filepath.Join(dir, "csvlinter.schema.json")
		mkTempSchema(schema)
		runValidateAndAssertJSON(t, csv, 0, func(res validator.Results) {
			if len(res.Errors) != 0 {
				t.Errorf("expected 0 errors, got %d", len(res.Errors))
			}
		})
	})

	t.Run("schema one level up until .git root", func(t *testing.T) {
		base := t.TempDir()
		// create fake project root with .git
		proj := filepath.Join(base, "project")
		if err := os.MkdirAll(filepath.Join(proj, ".git"), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		nested := filepath.Join(proj, "sub")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		csv := mkTempCSV(nested, "deep.csv", "id,name\n1,Carl")
		schema := filepath.Join(proj, "csvlinter.schema.json")
		mkTempSchema(schema)
		runValidateAndAssertJSON(t, csv, 0, func(res validator.Results) {
			if len(res.Errors) != 0 {
				t.Errorf("expected 0 errors, got %d", len(res.Errors))
			}
		})
	})

	t.Run("no schema found", func(t *testing.T) {
		dir := t.TempDir()
		csv := mkTempCSV(dir, "plain.csv", "id,name\n1,Zed")
		runValidateAndAssertJSON(t, csv, 0, func(res validator.Results) {
			// structure valid, schema skipped -> no errors
			if len(res.Errors) != 0 {
				t.Errorf("expected 0 errors w/o schema, got %d", len(res.Errors))
			}
		})
	})
}

// runValidateAndAssertJSON invokes the validate command with --format json and no --schema.
func runValidateAndAssertJSON(t *testing.T, csvPath string, wantExit int, assertFn func(validator.Results)) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	var exitCode int

	app := &cli.App{
		Commands:  []*cli.Command{validateCommand},
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

	args := []string{"csvlinter", "validate", "--format", "json", csvPath}
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
