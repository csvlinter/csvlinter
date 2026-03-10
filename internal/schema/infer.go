package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const draft07Meta = "http://json-schema.org/draft-07/schema#"

type inferredSchema struct {
	Schema               string                `json:"$schema"`
	Title                string                `json:"title,omitempty"`
	Type                 string                `json:"type"`
	Required             []string              `json:"required,omitempty"`
	Properties           map[string]propSchema `json:"properties"`
	AdditionalProperties bool                  `json:"additionalProperties"`
}

type propSchema struct {
	Type   string `json:"type"`
	Format string `json:"format,omitempty"`
}

// Format-detection regexes. All patterns require full-string matches.
var (
	// RFC 3339 date-time: 2024-01-15T10:30:00Z or ...+05:30
	reDateTime = regexp.MustCompile(
		`^\d{4}-\d{2}-\d{2}[T]\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`)
	// ISO 8601 / RFC 3339 date: 2024-01-15
	// This is a structural heuristic: it matches the YYYY-MM-DD shape but does
	// not validate calendar ranges, so values like 9999-99-99 are accepted.
	reDate = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	// RFC 3339 partial-time with optional timezone: 10:30:00, 10:30:00.123Z, 10:30:00+05:30
	reTime = regexp.MustCompile(
		`^\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})?$`)
	// Simple email heuristic: local@domain.tld
	reEmail = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	// URI: scheme://...
	reURI = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+\-.]*://\S+$`)
)

// inferColumnFormat returns a JSON Schema draft-07 format keyword for a
// string column whose non-empty sample values all match a recognised pattern,
// or "" if no single format applies to every value.
func inferColumnFormat(values []string) string {
	type formatRule struct {
		name string
		re   *regexp.Regexp
	}
	// Ordered from most-specific to least-specific so date-time is checked
	// before plain date.
	rules := []formatRule{
		{"date-time", reDateTime},
		{"date", reDate},
		{"time", reTime},
		{"email", reEmail},
		{"uri", reURI},
	}
	for _, rule := range rules {
		hasValue := false
		allMatch := true
		for _, s := range values {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			hasValue = true
			if !rule.re.MatchString(s) {
				allMatch = false
				break
			}
		}
		if hasValue && allMatch {
			return rule.name
		}
	}
	return ""
}

func inferColumnType(values []string) string {
	var seenInteger, seenNumber, seenBoolean, seenOther bool
	for _, s := range values {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, err := strconv.ParseInt(s, 10, 64); err == nil {
			seenInteger = true
			continue
		}
		if _, err := strconv.ParseFloat(s, 64); err == nil {
			seenNumber = true
			continue
		}
		lower := strings.ToLower(s)
		if lower == "true" || lower == "false" {
			seenBoolean = true
			continue
		}
		seenOther = true
	}
	if seenOther || (seenBoolean && (seenInteger || seenNumber)) || (seenInteger && seenNumber) {
		return "string"
	}
	if seenBoolean && !seenInteger && !seenNumber {
		return "boolean"
	}
	if seenInteger && !seenNumber {
		return "integer"
	}
	if seenNumber {
		return "number"
	}
	return "string"
}

func columnValues(sample [][]string, colIndex int) []string {
	out := make([]string, 0, len(sample))
	for _, row := range sample {
		if colIndex < len(row) {
			out = append(out, row[colIndex])
		}
	}
	return out
}

func hasNonEmpty(sample [][]string, colIndex int) bool {
	for _, row := range sample {
		if colIndex < len(row) && strings.TrimSpace(row[colIndex]) != "" {
			return true
		}
	}
	return false
}

// Infer produces a JSON Schema (draft-07) from headers and sample rows.
// Types are inferred per column (string, integer, number, boolean); when in doubt, string is used.
// Columns with at least one non-empty value in the sample are required.
func Infer(headers []string, sample [][]string) ([]byte, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("headers cannot be empty")
	}
	props := make(map[string]propSchema, len(headers))
	var required []string
	for i, h := range headers {
		values := columnValues(sample, i)
		t := inferColumnType(values)
		var format string
		if t == "string" {
			format = inferColumnFormat(values)
		}
		props[h] = propSchema{Type: t, Format: format}
		if hasNonEmpty(sample, i) {
			required = append(required, h)
		}
	}
	s := inferredSchema{
		Schema:               draft07Meta,
		Title:                "CSV row",
		Type:                 "object",
		Required:             required,
		Properties:           props,
		AdditionalProperties: false,
	}
	return json.MarshalIndent(s, "", "  ")
}
