package commands

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExport_DefaultJSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "" // defaults to json in non-TTY (tests), table on TTY → overridden to json for export
	stdout, _, err := executeCommand("export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Errorf("expected valid JSON, got %q", stdout)
	}
}

func TestExport_ExplicitJSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["identity"]; !ok {
		t.Error("expected identity in JSON export")
	}
}

func TestExport_Env(t *testing.T) {
	setupTestDB(t)
	flagFormat = "env"
	stdout, _, err := executeCommand("export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "DEETS_IDENTITY_NAME=") {
		t.Error("expected env var format")
	}
	if !strings.Contains(stdout, "DEETS_WEB_GITHUB=") {
		t.Error("expected web github env var")
	}
}

func TestExport_TOML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "toml"
	stdout, _, err := executeCommand("export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "[identity]") {
		t.Error("expected [identity] section")
	}
	if !strings.Contains(stdout, `name = "Alexander Towell"`) {
		t.Error("expected name field in TOML")
	}
}

func TestExport_YAML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "yaml"
	stdout, _, err := executeCommand("export")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "identity:") {
		t.Error("expected identity: key in YAML")
	}
	if !strings.Contains(stdout, "  name: Alexander Towell") {
		t.Error("expected name field in YAML")
	}
}
