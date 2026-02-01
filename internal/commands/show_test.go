package commands

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestShow_Table(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Category") {
		t.Error("expected Category column header in table output")
	}
	if !strings.Contains(stdout, "identity") {
		t.Error("expected identity category in output")
	}
	if !strings.Contains(stdout, "Alexander Towell") {
		t.Error("expected name value in output")
	}
}

func TestShow_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Errorf("expected valid JSON, got %q", stdout)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("invalid JSON structure: %v", err)
	}
	if _, ok := parsed["identity"]; !ok {
		t.Error("expected identity category in JSON")
	}
}

func TestShow_TOML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "toml"
	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "[identity]") {
		t.Error("expected [identity] section")
	}
	if !strings.Contains(stdout, "[web]") {
		t.Error("expected [web] section")
	}
}

func TestShow_YAML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "yaml"
	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "identity:") {
		t.Error("expected identity: key")
	}
	if !strings.Contains(stdout, "web:") {
		t.Error("expected web: key")
	}
}

func TestShow_Env(t *testing.T) {
	setupTestDB(t)
	flagFormat = "env"
	stdout, _, err := executeCommand("show")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "DEETS_IDENTITY_NAME=") {
		t.Error("expected env format output")
	}
}

func TestShow_SingleCategory_Table(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("show", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "name") {
		t.Error("expected name key in output")
	}
	// Single category should not show Category column
	if strings.Contains(stdout, "Category") {
		t.Error("single-category show should not include Category column")
	}
}

func TestShow_SingleCategory_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("show", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["name"]; !ok {
		t.Error("expected flat object with 'name' key")
	}
}

func TestShow_SingleCategory_NotFound(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("show", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent category")
	}
}
