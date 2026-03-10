package schema

import (
	"strings"
	"testing"
)

// TestValidateRow_RootLevelErrors tests that constraints evaluated at the root of
// the JSON object (InstanceLocation == "") are reported, not silently dropped.
func TestValidateRow_RootLevelErrors(t *testing.T) {
	// Schema: requires "id" (integer), no additional properties allowed.
	const schemaJSON = `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"required": ["id"],
		"properties": {
			"id": {"type": "integer"}
		},
		"additionalProperties": false
	}`

	v, err := NewValidatorFromReader(strings.NewReader(schemaJSON))
	if err != nil {
		t.Fatalf("NewValidatorFromReader: %v", err)
	}

	t.Run("required field missing produces error", func(t *testing.T) {
		// Row has "name" column only — "id" is required but absent.
		errs, err := v.ValidateRow([]string{"name"}, []string{"Alice"})
		if err != nil {
			t.Fatalf("ValidateRow: %v", err)
		}
		if len(errs) == 0 {
			t.Fatal("expected errors for missing required field 'id' " +
				"(and extra field 'name'), but got none")
		}
	})

	t.Run("additionalProperties violation produces error", func(t *testing.T) {
		// "name" is not declared in properties, so it violates additionalProperties:false.
		errs, err := v.ValidateRow([]string{"name"}, []string{"Alice"})
		if err != nil {
			t.Fatalf("ValidateRow: %v", err)
		}
		found := false
		for _, e := range errs {
			msg := strings.ToLower(e.Message)
			if strings.Contains(msg, "additional") || strings.Contains(msg, "required") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error mentioning 'additional' or 'required', got: %v", errs)
		}
	})

	t.Run("field-level type error still reported", func(t *testing.T) {
		// "id" present but wrong type — this produces InstanceLocation=="/id" and
		// must still work after the fix.
		const schemaWithID = `{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type": "object",
			"properties": {"id": {"type": "integer"}}
		}`
		vt, err := NewValidatorFromReader(strings.NewReader(schemaWithID))
		if err != nil {
			t.Fatalf("NewValidatorFromReader: %v", err)
		}
		errs, err := vt.ValidateRow([]string{"id"}, []string{"not-a-number"})
		if err != nil {
			t.Fatalf("ValidateRow: %v", err)
		}
		if len(errs) == 0 {
			t.Error("expected type error for non-integer 'id' value, got none")
		}
		for _, e := range errs {
			if e.Field != "id" {
				t.Errorf("expected Field=id, got %q", e.Field)
			}
		}
	})
}
