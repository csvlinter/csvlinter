package schema

import (
	"os"
	"path/filepath"
)

// Project root indicators
var projectRootIndicators = []string{".git", "package.json"}

// ResolveSchema attempts to find a schema file for the given CSV path according to fallback rules.
// Returns the schema path if found, or an empty string if not found.
func ResolveSchema(csvPath string) string {
	csvDir := filepath.Dir(csvPath)
	csvBase := filepath.Base(csvPath)
	csvName := csvBase[:len(csvBase)-len(filepath.Ext(csvBase))]

	// 1. Look for <filename>.schema.json in the same folder
	candidate := filepath.Join(csvDir, csvName+".schema.json")
	if fileExists(candidate) {
		return candidate
	}

	// 2. Look for csvlinter.schema.json in the same folder
	candidate = filepath.Join(csvDir, "csvlinter.schema.json")
	if fileExists(candidate) {
		return candidate
	}

	// 3. Walk up parent directories, stopping at project root or system root
	dir := csvDir
	for {
		if isProjectRoot(dir) || isSystemRoot(dir) {
			break
		}
		candidate := filepath.Join(dir, "csvlinter.schema.json")
		if fileExists(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir { // system root
			break
		}
		dir = parent
	}

	return ""
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isProjectRoot(dir string) bool {
	for _, marker := range projectRootIndicators {
		if fileExists(filepath.Join(dir, marker)) {
			return true
		}
	}
	return false
}

func isSystemRoot(dir string) bool {
	parent := filepath.Dir(dir)
	return dir == parent
}
