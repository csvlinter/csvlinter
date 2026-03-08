package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/csvlinter/csvlinter/internal/validator"

	"github.com/urfave/cli/v2"
)

func TestValidateCommand(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Helper to create a temporary file
	createTempFile := func(name, content string) string {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file %s: %v", name, err)
		}
		return path
	}

	// Create test files
	schemaPath := createTempFile("schema.json", `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"id": { "type": "integer" },
			"name": { "type": "string" },
			"email": {
				"type": "string",
				"pattern": "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
			}
		},
		"required": ["id", "name", "email"]
	}`)

	validCSVPath := createTempFile("valid.csv", `id,name,email
1,John Doe,john.doe@example.com
2,Jane Doe,jane.doe@example.com`)

	invalidCSVPath := createTempFile("invalid.csv", `id,name,email
1,John Doe,not-an-email
invalid,Jane Doe,jane.doe@example.com`)

	// Define test cases
	testCases := []struct {
		name           string
		args           []string
		stdinContent   string // written to stdin when args contain "-"
		expectedExit   int
		expectError    bool
		expectedOutput string
		assertOutput   func(t *testing.T, output string)
	}{
		{
			name:         "Successful validation with JSON output",
			args:         []string{"--schema", schemaPath, "--format", "json", validCSVPath},
			expectedExit: 0,
			expectError:  false,
			assertOutput: func(t *testing.T, output string) {
				var results validator.Results
				if err := json.Unmarshal([]byte(output), &results); err != nil {
					t.Fatalf("Failed to unmarshal JSON output: %v", err)
				}
				if len(results.Errors) != 0 {
					t.Errorf("Expected 0 validation errors, got %d", len(results.Errors))
				}
			},
		},
		{
			name:         "Validation failure with JSON output",
			args:         []string{"--schema", schemaPath, "--format", "json", invalidCSVPath},
			expectedExit: 1,
			expectError:  true,
			assertOutput: func(t *testing.T, output string) {
				var results validator.Results
				if err := json.Unmarshal([]byte(output), &results); err != nil {
					t.Fatalf("Failed to unmarshal JSON output: %v", err)
				}
				if len(results.Errors) == 0 {
					t.Errorf("Expected validation errors, got none")
				}
				// Optionally, more specific checks on the errors
				if len(results.Errors) != 2 {
					t.Errorf("Expected 2 validation errors, got %d", len(results.Errors))
				}
			},
		},
		{
			name:         "Validation failure with pretty output exits 1",
			args:         []string{"--schema", schemaPath, "--format", "pretty", invalidCSVPath},
			expectedExit: 1,
			expectError:  true,
			assertOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "INVALID") {
					t.Errorf("Expected pretty output to contain INVALID, got: %s", output)
				}
			},
		},
		{
			name:         "Non-existent CSV file",
			args:         []string{"non-existent.csv"},
			expectedExit: 1,
			expectError:  true,
		},
		{
			name:         "Non-existent schema file",
			args:         []string{"--schema", "non-existent.json", validCSVPath},
			expectedExit: 1,
			expectError:  true,
		},
		{
			name:         "STDIN input with JSON output",
			args:         []string{"--format", "json", "-"},
			stdinContent: "name,email\nJohn Doe,john@example.com\n",
			expectedExit: 0,
			expectError:  false,
			assertOutput: func(t *testing.T, output string) {
				var results validator.Results
				if err := json.Unmarshal([]byte(output), &results); err != nil {
					t.Fatalf("Failed to unmarshal JSON output: %v", err)
				}
				if results.File != "STDIN" {
					t.Errorf("Expected file name to be STDIN, got %s", results.File)
				}
			},
		},
		{
			name:         "Empty STDIN with JSON format emits valid JSON",
			args:         []string{"--format", "json", "--filename", "empty.csv", "-"},
			stdinContent: "", // empty — triggers the "no headers found" error path
			expectedExit: 1,
			assertOutput: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output == "" {
					return // nothing written is also acceptable
				}
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Expected valid JSON, got plain text: %s", output)
				}
			},
		},
		{
			name:         "Non-existent CSV file with JSON format emits valid JSON",
			args:         []string{"--format", "json", "no-such-file.csv"},
			expectedExit: 1,
			assertOutput: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output == "" {
					return
				}
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Expected valid JSON, got plain text: %s", output)
				}
			},
		},
		{
			name:         "Non-existent schema with JSON format emits valid JSON",
			args:         []string{"--format", "json", "--schema", "no-such-schema.json", validCSVPath},
			expectedExit: 1,
			assertOutput: func(t *testing.T, output string) {
				output = strings.TrimSpace(output)
				if output == "" {
					return
				}
				var result map[string]any
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Fatalf("Expected valid JSON, got plain text: %s", output)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var exitCode int
			var outputBuffer, errorBuffer bytes.Buffer

			// Create a new app instance for each test
			app := &cli.App{
				Commands: []*cli.Command{
					validateCommand,
				},
				Writer:    &outputBuffer,
				ErrWriter: &errorBuffer,
				ExitErrHandler: func(c *cli.Context, err error) {
					if err != nil {
						if exitErr, ok := err.(cli.ExitCoder); ok {
							exitCode = exitErr.ExitCode()
						} else {
							exitCode = 1
						}
						fmt.Fprintln(c.App.ErrWriter, err)
					}
				},
			}

			// If testing STDIN, set up a pipe
			if tc.stdinContent != "" {
				// Save original stdin
				oldStdin := os.Stdin
				defer func() { os.Stdin = oldStdin }()

				// Create a pipe
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatalf("Failed to create pipe: %v", err)
				}

				// Set stdin to read end of pipe
				os.Stdin = r

				go func() {
					defer w.Close()
					fmt.Fprint(w, tc.stdinContent)
				}()
			}

			// Run the command
			args := append([]string{"csvlinter", "validate"}, tc.args...)
			err := app.Run(args)

			// Check for error during app run (should be handled by ExitErrHandler)
			if err != nil && tc.expectedExit == 0 {
				t.Logf("App run failed unexpectedly: %v", err)
			}

			// Check exit code
			if exitCode != tc.expectedExit {
				t.Errorf("Expected exit code %d, got %d", tc.expectedExit, exitCode)
				t.Logf("Stdout: %s", outputBuffer.String())
				t.Logf("Stderr: %s", errorBuffer.String())
			}

			// Assert output
			if tc.assertOutput != nil {
				tc.assertOutput(t, outputBuffer.String())
			}
		})
	}
}
