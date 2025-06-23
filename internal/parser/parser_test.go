package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		delimiter     string
		expectHeaders []string
		expectRows    [][]string
		expectError   bool
	}{
		{
			name:          "Valid CSV with comma delimiter",
			input:         "name,email,age\nJohn,john@example.com,30\nJane,jane@example.com,25",
			delimiter:     ",",
			expectHeaders: []string{"name", "email", "age"},
			expectRows: [][]string{
				{"John", "john@example.com", "30"},
				{"Jane", "jane@example.com", "25"},
			},
			expectError: false,
		},
		{
			name:          "Valid CSV with semicolon delimiter",
			input:         "name;email;age\nJohn;john@example.com;30\nJane;jane@example.com;25",
			delimiter:     ";",
			expectHeaders: []string{"name", "email", "age"},
			expectRows: [][]string{
				{"John", "john@example.com", "30"},
				{"Jane", "jane@example.com", "25"},
			},
			expectError: false,
		},
		{
			name:          "Empty CSV",
			input:         "",
			delimiter:     ",",
			expectHeaders: nil,
			expectRows:    nil,
			expectError:   true,
		},
		{
			name:          "Headers only",
			input:         "name,email,age",
			delimiter:     ",",
			expectHeaders: []string{"name", "email", "age"},
			expectRows:    [][]string{},
			expectError:   false,
		},
		{
			name:          "Invalid UTF-8",
			input:         "name,email\nJohn,john@example.com\n" + string([]byte{0xFF, 0xFE, 0xFD}),
			delimiter:     ",",
			expectHeaders: nil,
			expectRows:    nil,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create parser
			p, err := NewParser(strings.NewReader(tc.input), tc.delimiter)
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Unexpected error creating parser: %v", err)
				}
				return
			}
			defer p.Close()

			// Validate UTF-8
			err = p.ValidateUTF8()
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Unexpected UTF-8 validation error: %v", err)
				}
				return
			}

			// Read headers
			headers, err := p.ReadHeaders()
			if err != nil {
				if !tc.expectError {
					t.Fatalf("Unexpected error reading headers: %v", err)
				}
				return
			}

			// Verify headers
			if !equalSlices(headers, tc.expectHeaders) {
				t.Errorf("Expected headers %v, got %v", tc.expectHeaders, headers)
			}

			// Read and verify rows
			var rows [][]string
			for {
				row, err := p.ReadRow()
				if err == io.EOF {
					break
				}
				if err != nil {
					if !tc.expectError {
						t.Fatalf("Unexpected error reading row: %v", err)
					}
					return
				}
				rows = append(rows, row.Data)
			}

			// Verify row count
			if len(rows) != len(tc.expectRows) {
				t.Errorf("Expected %d rows, got %d", len(tc.expectRows), len(rows))
			}

			// Verify row contents
			for i := range rows {
				if i >= len(tc.expectRows) {
					break
				}
				if !equalSlices(rows[i], tc.expectRows[i]) {
					t.Errorf("Row %d: expected %v, got %v", i, tc.expectRows[i], rows[i])
				}
			}
		})
	}
}

func TestParserWithEmptyRows(t *testing.T) {
	input := "name,email\nJohn,john@example.com\n\n\nJane,jane@example.com\n\n"
	p, err := NewParser(strings.NewReader(input), ",")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Read headers
	headers, err := p.ReadHeaders()
	if err != nil {
		t.Fatalf("Failed to read headers: %v", err)
	}
	if !equalSlices(headers, []string{"name", "email"}) {
		t.Errorf("Expected headers [name email], got %v", headers)
	}

	// Read rows and verify empty rows are skipped
	expectedRows := [][]string{
		{"John", "john@example.com"},
		{"Jane", "jane@example.com"},
	}

	var rows [][]string
	for {
		row, err := p.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read row: %v", err)
		}
		if !row.IsEmpty() {
			rows = append(rows, row.Data)
		}
	}

	if len(rows) != len(expectedRows) {
		t.Errorf("Expected %d non-empty rows, got %d", len(expectedRows), len(rows))
	}

	for i := range rows {
		if !equalSlices(rows[i], expectedRows[i]) {
			t.Errorf("Row %d: expected %v, got %v", i, expectedRows[i], rows[i])
		}
	}
}

func TestParserWithLargeInput(t *testing.T) {
	// Create a large input (>4KB to test buffering)
	var buf bytes.Buffer
	buf.WriteString("id,name,email,data\n")
	for i := 0; i < 1000; i++ {
		buf.WriteString("1,John Doe,john@example.com,")
		buf.WriteString(strings.Repeat("x", 100))
		buf.WriteString("\n")
	}

	p, err := NewParser(&buf, ",")
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}
	defer p.Close()

	// Read headers
	headers, err := p.ReadHeaders()
	if err != nil {
		t.Fatalf("Failed to read headers: %v", err)
	}
	if !equalSlices(headers, []string{"id", "name", "email", "data"}) {
		t.Errorf("Expected headers [id name email data], got %v", headers)
	}

	// Read all rows
	rowCount := 0
	for {
		row, err := p.ReadRow()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Failed to read row: %v", err)
		}
		if !row.IsEmpty() {
			rowCount++
		}
	}

	if rowCount != 1000 {
		t.Errorf("Expected 1000 rows, got %d", rowCount)
	}
}

// Helper function to compare slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
