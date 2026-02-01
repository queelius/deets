package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestInferType(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected string
	}{
		{"hello", "string"},
		{[]interface{}{"a"}, "array"},
		{[]string{"a"}, "array"},
		{int64(42), "integer"},
		{float64(3.14), "float"},
		{true, "boolean"},
		{struct{}{}, "unknown"},
	}

	for _, tt := range tests {
		got := InferType(tt.value)
		if got != tt.expected {
			t.Errorf("InferType(%v) = %q, want %q", tt.value, got, tt.expected)
		}
	}
}

func TestBuildSchema(t *testing.T) {
	db := newTestDB()
	schema := BuildSchema(db)

	if len(schema) == 0 {
		t.Fatal("expected at least one schema entry")
	}

	// Check that _desc keys are excluded
	for _, s := range schema {
		if strings.HasSuffix(s.Key, "_desc") {
			t.Errorf("_desc key should be excluded: %s", s.Key)
		}
	}

	// Check specific entries
	types := make(map[string]string)
	for _, s := range schema {
		types[s.Category+"."+s.Key] = s.Type
	}

	expected := map[string]string{
		"identity.name": "string",
		"identity.aka":  "array",
		"identity.age":  "integer",
		"web.github":    "string",
		"academic.gpa":  "float",
	}

	for path, expectedType := range expected {
		if types[path] != expectedType {
			t.Errorf("expected %s type %q, got %q", path, expectedType, types[path])
		}
	}
}

func TestBuildSchema_HasDescriptions(t *testing.T) {
	db := newTestDB()
	schema := BuildSchema(db)

	for _, s := range schema {
		if s.Key == "name" && s.Category == "identity" {
			if s.Description != "Full legal name" {
				t.Errorf("expected 'Full legal name', got %q", s.Description)
			}
			return
		}
	}
	t.Error("identity.name not found in schema")
}

func TestBuildSchema_HasExamples(t *testing.T) {
	db := newTestDB()
	schema := BuildSchema(db)

	for _, s := range schema {
		if s.Key == "name" && s.Category == "identity" {
			if s.Example != "Alexander Towell" {
				t.Errorf("expected example 'Alexander Towell', got %q", s.Example)
			}
			return
		}
	}
	t.Error("identity.name not found in schema")
}

func TestFormatSchemaTable(t *testing.T) {
	entries := []SchemaField{
		{Category: "identity", Key: "name", Type: "string", Description: "Full name", Example: "Alex"},
		{Category: "academic", Key: "gpa", Type: "float", Description: "", Example: "3.95"},
	}

	out := FormatSchemaTable(entries)
	if !strings.Contains(out, "Category") {
		t.Error("expected Category header")
	}
	if !strings.Contains(out, "Type") {
		t.Error("expected Type header")
	}
	if !strings.Contains(out, "string") {
		t.Error("expected 'string' type")
	}
	if !strings.Contains(out, "float") {
		t.Error("expected 'float' type")
	}
}

func TestFormatSchemaTable_Empty(t *testing.T) {
	out := FormatSchemaTable(nil)
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}

func TestFormatSchemaJSON(t *testing.T) {
	entries := []SchemaField{
		{Category: "identity", Key: "name", Type: "string", Description: "Full name", Example: "Alex"},
	}

	out, err := FormatSchemaJSON(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed []SchemaField
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0].Type != "string" {
		t.Errorf("expected type 'string', got %q", parsed[0].Type)
	}
}

func TestBuildSchema_EmptyDB(t *testing.T) {
	db := &DB{}
	schema := BuildSchema(db)
	if len(schema) != 0 {
		t.Errorf("expected 0 entries for empty DB, got %d", len(schema))
	}
}
