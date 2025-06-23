package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(p string) {
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(`{}`), 0o644)
}

func TestResolveSchema(t *testing.T) {
	cases := []struct {
		name  string
		setup func(dir string) (csvPath, want string)
	}{
		{
			name: "colocated fileName.schema.json",
			setup: func(dir string) (string, string) {
				csv := filepath.Join(dir, "data.csv")
				writeFile(csv)
				schema := filepath.Join(dir, "data.schema.json")
				writeFile(schema)
				return csv, schema
			},
		},
		{
			name: "csvlinter.schema.json in same dir",
			setup: func(dir string) (string, string) {
				csv := filepath.Join(dir, "sample.csv")
				writeFile(csv)
				schema := filepath.Join(dir, "csvlinter.schema.json")
				writeFile(schema)
				return csv, schema
			},
		},
		{
			name: "schema up the tree until .git project root",
			setup: func(dir string) (string, string) {
				// tmp/project/.git
				proj := filepath.Join(dir, "project")
				_ = os.MkdirAll(filepath.Join(proj, ".git"), 0o755)

				// tmp/project/sub/nested/file.csv
				nested := filepath.Join(proj, "sub", "nested")
				_ = os.MkdirAll(nested, 0o755)
				csv := filepath.Join(nested, "file.csv")
				writeFile(csv)

				// tmp/project/csvlinter.schema.json
				schema := filepath.Join(proj, "csvlinter.schema.json")
				writeFile(schema)

				return csv, schema
			},
		},
		{
			name: "no schema anywhere",
			setup: func(dir string) (string, string) {
				csv := filepath.Join(dir, "lonely.csv")
				writeFile(csv)
				return csv, "" // want empty
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			csvPath, want := tc.setup(base)

			got := ResolveSchema(csvPath)
			if got != want {
				t.Fatalf("want %q, got %q", want, got)
			}
		})
	}
}
