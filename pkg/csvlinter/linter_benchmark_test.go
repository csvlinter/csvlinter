package csvlinter

import (
	"bytes"
	"fmt"
	"testing"
)

// Helper to generate a CSV of n rows
func generateCSV(rowCount int) []byte {
	var buf bytes.Buffer
	buf.WriteString("name,age,email,city\n")
	for i := 0; i < rowCount; i++ {
		fmt.Fprintf(&buf, "User%d,%d,user%d@example.com,City%d\n", i, 20+(i%50), i, i%100)
	}
	return buf.Bytes()
}

func BenchmarkLint_SmallCSV_NoSchema(b *testing.B) {
	data := generateCSV(1000) // ~small CSV
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := Lint(reader, "small.csv", ",")
		if err != nil {
			b.Fatalf("Lint failed: %v", err)
		}
	}
}

func BenchmarkLint_SmallCSV_WithSchema(b *testing.B) {
	data := generateCSV(1000)
	schemaPath := "../../testdata/schemas/schema.json"
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := LintWithSchema(reader, "small.csv", ",", schemaPath)
		if err != nil {
			b.Fatalf("LintWithSchema failed: %v", err)
		}
	}
}

func BenchmarkLint_LargeCSV_NoSchema(b *testing.B) {
	data := generateCSV(1_000_000) // ~large CSV
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := Lint(reader, "large.csv", ",")
		if err != nil {
			b.Fatalf("Lint failed: %v", err)
		}
	}
}

func BenchmarkLint_LargeCSV_WithSchema(b *testing.B) {
	data := generateCSV(1_000_000)
	schemaPath := "../../testdata/schemas/schema.json"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, err := LintWithSchema(reader, "large.csv", ",", schemaPath)
		if err != nil {
			b.Fatalf("LintWithSchema failed: %v", err)
		}
	}
}
