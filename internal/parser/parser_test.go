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
			expectHeaders: []string{"name", "email"},
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

func TestComputeSampleK(t *testing.T) {
	cases := []struct {
		name      string
		totalRows int
		maxRows   int
		wantK     int
	}{
		// Fewer rows than the floor — should still get all rows.
		{"zero rows", 0, 100, 0},
		{"1 row", 1, 100, 1},
		{"2 rows", 2, 100, 2},
		{"4 rows", 4, 100, 4},
		// Floor kicks in (ceil(5*0.25)=2 < floor=5, so k=5).
		{"5 rows", 5, 100, 5},
		{"10 rows", 10, 100, 5},
		// 25% ramps up past floor.
		{"20 rows", 20, 100, 5},
		{"28 rows (testdata file)", 28, 100, 7},
		// Hard cap enforced.
		{"400 rows cap=100", 400, 100, 100},
		{"1000 rows cap=100", 1000, 100, 100},
		// maxRows lower than would-be k — maxRows wins.
		{"5 rows cap=2", 5, 2, 2},
		{"100 rows cap=10", 100, 10, 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeSampleK(tc.totalRows, tc.maxRows)
			if got != tc.wantK {
				t.Errorf("computeSampleK(%d, %d) = %d, want %d",
					tc.totalRows, tc.maxRows, got, tc.wantK)
			}
		})
	}
}

func TestReadSampleFromReader(t *testing.T) {
	t.Run("returns headers and sample", func(t *testing.T) {
		r := strings.NewReader("id,name\n1,Alice\n2,Bob")
		headers, sample, replay, err := ReadSampleFromReader(r, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromReader: %v", err)
		}
		if !equalSlices(headers, []string{"id", "name"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 rows, got %d", len(sample))
		}
		if replay == nil {
			t.Fatal("replay must not be nil")
		}
	})

	t.Run("respects maxRows: sample is capped, replay delivers all rows", func(t *testing.T) {
		// 5 data rows; sample only 2; replay must still yield all 5.
		input := "a,b\n1,x\n2,y\n3,z\n4,w\n5,v"
		headers, sample, replay, err := ReadSampleFromReader(strings.NewReader(input), ",", 2)
		if err != nil {
			t.Fatalf("ReadSampleFromReader: %v", err)
		}
		if !equalSlices(headers, []string{"a", "b"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 rows (maxRows=2), got %d", len(sample))
		}

		// Parse the replay stream and count total data rows — must be 5.
		p, err := NewParser(replay, ",")
		if err != nil {
			t.Fatalf("NewParser on replay: %v", err)
		}
		if _, err := p.ReadHeaders(); err != nil {
			t.Fatalf("ReadHeaders on replay: %v", err)
		}
		totalRows := 0
		for {
			row, err := p.ReadRow()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("ReadRow on replay: %v", err)
			}
			if !row.IsEmpty() {
				totalRows++
			}
		}
		if totalRows != 5 {
			t.Errorf("replay: want 5 total data rows, got %d", totalRows)
		}
	})

	t.Run("skips empty rows in sample", func(t *testing.T) {
		r := strings.NewReader("x,y\n1,a\n\n\n2,b")
		_, sample, _, err := ReadSampleFromReader(r, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromReader: %v", err)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 non-empty rows, got %d", len(sample))
		}
	})

	t.Run("file shorter than maxRows", func(t *testing.T) {
		r := strings.NewReader("a,b\n1,x")
		_, sample, replay, err := ReadSampleFromReader(r, ",", 100)
		if err != nil {
			t.Fatalf("ReadSampleFromReader: %v", err)
		}
		if len(sample) != 1 {
			t.Errorf("sample: want 1 row, got %d", len(sample))
		}
		if replay == nil {
			t.Fatal("replay must not be nil even when file shorter than maxRows")
		}
	})

	t.Run("empty input returns error", func(t *testing.T) {
		_, _, _, err := ReadSampleFromReader(strings.NewReader(""), ",", 10)
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("headers only returns empty sample and full replay", func(t *testing.T) {
		r := strings.NewReader("a,b,c")
		headers, sample, replay, err := ReadSampleFromReader(r, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromReader: %v", err)
		}
		if !equalSlices(headers, []string{"a", "b", "c"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 0 {
			t.Errorf("sample: want 0 rows, got %d", len(sample))
		}
		if replay == nil {
			t.Fatal("replay must not be nil")
		}
	})
}

func TestReadSampleFromBytes(t *testing.T) {
	t.Run("headers and sample from small csv", func(t *testing.T) {
		csv := []byte("id,name\n1,Alice\n2,Bob")
		headers, sample, err := ReadSampleFromBytes(csv, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromBytes: %v", err)
		}
		if !equalSlices(headers, []string{"id", "name"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 rows, got %d", len(sample))
		}
		if len(sample) > 0 && !equalSlices(sample[0], []string{"1", "Alice"}) {
			t.Errorf("sample[0]: got %v", sample[0])
		}
		if len(sample) > 1 && !equalSlices(sample[1], []string{"2", "Bob"}) {
			t.Errorf("sample[1]: got %v", sample[1])
		}
	})

	t.Run("respects maxRows", func(t *testing.T) {
		csv := []byte("a,b\n1,x\n2,y\n3,z")
		headers, sample, err := ReadSampleFromBytes(csv, ",", 2)
		if err != nil {
			t.Fatalf("ReadSampleFromBytes: %v", err)
		}
		if !equalSlices(headers, []string{"a", "b"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 rows (maxRows=2), got %d", len(sample))
		}
	})

	t.Run("skips empty rows", func(t *testing.T) {
		csv := []byte("x,y\n1,a\n\n\n2,b")
		headers, sample, err := ReadSampleFromBytes(csv, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromBytes: %v", err)
		}
		if !equalSlices(headers, []string{"x", "y"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 2 {
			t.Errorf("sample: want 2 non-empty rows, got %d", len(sample))
		}
		if len(sample) > 0 && !equalSlices(sample[0], []string{"1", "a"}) {
			t.Errorf("sample[0]: got %v", sample[0])
		}
		if len(sample) > 1 && !equalSlices(sample[1], []string{"2", "b"}) {
			t.Errorf("sample[1]: got %v", sample[1])
		}
	})

	t.Run("empty input error", func(t *testing.T) {
		_, _, err := ReadSampleFromBytes([]byte{}, ",", 10)
		if err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("headers only returns empty sample", func(t *testing.T) {
		csv := []byte("a,b,c")
		headers, sample, err := ReadSampleFromBytes(csv, ",", 10)
		if err != nil {
			t.Fatalf("ReadSampleFromBytes: %v", err)
		}
		if !equalSlices(headers, []string{"a", "b", "c"}) {
			t.Errorf("headers: got %v", headers)
		}
		if len(sample) != 0 {
			t.Errorf("sample: want 0 rows, got %d", len(sample))
		}
	})

	t.Run("same bytes second read gives same result", func(t *testing.T) {
		csv := []byte("id,name\n1,Alice\n2,Bob")
		h1, s1, err1 := ReadSampleFromBytes(csv, ",", 10)
		if err1 != nil {
			t.Fatalf("first read: %v", err1)
		}
		h2, s2, err2 := ReadSampleFromBytes(csv, ",", 10)
		if err2 != nil {
			t.Fatalf("second read: %v", err2)
		}
		if !equalSlices(h1, h2) {
			t.Errorf("headers differ: %v vs %v", h1, h2)
		}
		if len(s1) != len(s2) {
			t.Errorf("sample length differ: %d vs %d", len(s1), len(s2))
		}
		for i := range s1 {
			if !equalSlices(s1[i], s2[i]) {
				t.Errorf("sample row %d differ: %v vs %v", i, s1[i], s2[i])
			}
		}
	})
}
