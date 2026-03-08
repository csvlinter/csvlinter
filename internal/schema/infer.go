package schema

import (
	"encoding/json"
	"fmt"
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
	Type string `json:"type"`
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
		props[h] = propSchema{Type: t}
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
