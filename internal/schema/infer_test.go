package schema

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestInfer_ValidDraft07(t *testing.T) {
	headers := []string{"id", "name"}
	sample := [][]string{{"1", "Alice"}, {"2", "Bob"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["$schema"] != "http://json-schema.org/draft-07/schema#" {
		t.Errorf("expected draft-07 $schema, got %v", parsed["$schema"])
	}
	if parsed["type"] != "object" {
		t.Errorf("expected type object, got %v", parsed["type"])
	}
	props, ok := parsed["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected properties object, got %T", parsed["properties"])
	}
	if len(props) != 2 {
		t.Errorf("expected 2 properties, got %d", len(props))
	}
	if parsed["additionalProperties"] != false {
		t.Errorf("expected additionalProperties false, got %v", parsed["additionalProperties"])
	}
}

func TestInfer_HeadersInProperties(t *testing.T) {
	headers := []string{"a", "b", "c"}
	sample := [][]string{{"x", "y", "z"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	for _, h := range headers {
		if _, ok := props[h]; !ok {
			t.Errorf("missing property %q", h)
		}
	}
}

func TestInfer_IntegerColumn(t *testing.T) {
	headers := []string{"id"}
	sample := [][]string{{"1"}, {"2"}, {"3"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	idProp := props["id"].(map[string]interface{})
	if idProp["type"] != "integer" {
		t.Errorf("expected id type integer, got %v", idProp["type"])
	}
}

func TestInfer_StringColumn(t *testing.T) {
	headers := []string{"name"}
	sample := [][]string{{"Alice"}, {"Bob"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	nameProp := props["name"].(map[string]interface{})
	if nameProp["type"] != "string" {
		t.Errorf("expected name type string, got %v", nameProp["type"])
	}
}

func TestInfer_NumberColumn(t *testing.T) {
	headers := []string{"score"}
	sample := [][]string{{"1.5"}, {"2.0"}, {"3.14"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	scoreProp := props["score"].(map[string]interface{})
	if scoreProp["type"] != "number" {
		t.Errorf("expected score type number, got %v", scoreProp["type"])
	}
}

func TestInfer_BooleanColumn(t *testing.T) {
	headers := []string{"active"}
	sample := [][]string{{"true"}, {"false"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	activeProp := props["active"].(map[string]interface{})
	if activeProp["type"] != "boolean" {
		t.Errorf("expected active type boolean, got %v", activeProp["type"])
	}
}

func TestInfer_EmptySample_AllString(t *testing.T) {
	headers := []string{"id", "name"}
	sample := [][]string{}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	for _, h := range headers {
		p := props[h].(map[string]interface{})
		if p["type"] != "string" {
			t.Errorf("expected %q type string with empty sample, got %v", h, p["type"])
		}
	}
	if required, ok := parsed["required"]; ok && required != nil {
		reqSlice, _ := required.([]interface{})
		if len(reqSlice) != 0 {
			t.Errorf("expected no required with empty sample, got %v", required)
		}
	}
}

func TestInfer_SingleRow(t *testing.T) {
	headers := []string{"id", "label"}
	sample := [][]string{{"42", "foo"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	if props["id"].(map[string]interface{})["type"] != "integer" {
		t.Errorf("expected id integer from single row")
	}
	if props["label"].(map[string]interface{})["type"] != "string" {
		t.Errorf("expected label string from single row")
	}
}

func TestInfer_MixedTypes_ConservativeString(t *testing.T) {
	headers := []string{"mixed"}
	sample := [][]string{{"1"}, {"two"}, {"3"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	props := parsed["properties"].(map[string]interface{})
	if props["mixed"].(map[string]interface{})["type"] != "string" {
		t.Errorf("expected mixed column to fall back to string, got %v", props["mixed"].(map[string]interface{})["type"])
	}
}

func TestInfer_RequiredFromSample(t *testing.T) {
	headers := []string{"a", "b"}
	sample := [][]string{{"1", ""}, {"2", "y"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	required, ok := parsed["required"].([]interface{})
	if !ok {
		t.Fatalf("expected required array")
	}
	if len(required) != 2 {
		t.Errorf("expected both columns required (each has at least one non-empty), got %d", len(required))
	}
}

func TestInfer_EmptyHeaders_Error(t *testing.T) {
	_, err := Infer([]string{}, [][]string{})
	if err == nil {
		t.Fatal("expected error for empty headers")
	}
}

func TestInfer_CompilesWithValidator(t *testing.T) {
	headers := []string{"id", "name"}
	sample := [][]string{{"1", "Alice"}, {"2", "Bob"}}
	out, err := Infer(headers, sample)
	if err != nil {
		t.Fatalf("Infer: %v", err)
	}
	_, err = NewValidatorFromReader(bytes.NewReader(out))
	if err != nil {
		t.Errorf("inferred schema should compile: %v", err)
	}
}
