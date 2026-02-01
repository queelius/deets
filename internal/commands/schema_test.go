package commands

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/queelius/deets/internal/model"
)

func TestSchema_Table(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Category") {
		t.Error("expected Category column header")
	}
	if !strings.Contains(stdout, "Type") {
		t.Error("expected Type column header")
	}
	if !strings.Contains(stdout, "string") {
		t.Error("expected 'string' type in output")
	}
	if !strings.Contains(stdout, "identity") {
		t.Error("expected 'identity' category in output")
	}
}

func TestSchema_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []model.SchemaField
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected at least one schema entry")
	}

	// Check type inference
	types := make(map[string]string)
	for _, e := range entries {
		types[e.Category+"."+e.Key] = e.Type
	}

	if types["identity.name"] != "string" {
		t.Errorf("expected identity.name type 'string', got %q", types["identity.name"])
	}
	if types["identity.aka"] != "array" {
		t.Errorf("expected identity.aka type 'array', got %q", types["identity.aka"])
	}
	if types["academic.gpa"] != "float" {
		t.Errorf("expected academic.gpa type 'float', got %q", types["academic.gpa"])
	}
}

func TestSchema_HasDescriptions(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("schema")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []model.SchemaField
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Fields with _desc companions should have descriptions
	for _, e := range entries {
		if e.Key == "name" && e.Category == "identity" {
			if e.Description != "Full legal name" {
				t.Errorf("expected description 'Full legal name', got %q", e.Description)
			}
			return
		}
	}
	t.Error("identity.name not found in schema entries")
}
