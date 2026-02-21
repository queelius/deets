package commands

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestGet_BareValue(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "Alexander Towell" {
		t.Errorf("expected bare value, got %q", stdout)
	}
}

func TestGet_JSON_BareValue(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("get", "identity.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Single exact field with --format json now returns bare value (not JSON)
	if strings.TrimSpace(stdout) != "Alexander Towell" {
		t.Errorf("expected bare value, got %q", stdout)
	}
}

func TestGet_CategoryQuery(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("get", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := parsed["name"]; !ok {
		t.Error("expected 'name' key in identity category output")
	}
}

func TestGet_NotFound_ExitCode2(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("get", "nonexistent.key")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 2 {
		t.Errorf("expected exit code 2, got %d", exitErr.Code)
	}
}

func TestGet_Default(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	// Use a flag-like invocation — we need to pass --default via args
	stdout, _, err := executeCommand("get", "nonexistent.key", "--default", "fallback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "fallback" {
		t.Errorf("expected 'fallback', got %q", stdout)
	}
}

func TestGet_DefaultEmpty(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "nonexistent.key", "--default", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected empty default, got %q", stdout)
	}
}

func TestGet_Exists_Found(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.name", "--exists")
	if err != nil {
		t.Fatalf("unexpected error for existing field: %v", err)
	}
	if stdout != "" {
		t.Errorf("--exists should produce no output, got %q", stdout)
	}
}

func TestGet_Exists_NotFound(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("get", "nonexistent.key", "--exists")
	if err == nil {
		t.Fatal("expected error for nonexistent field with --exists")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Errorf("expected ExitError with code 2, got %v", err)
	}
}

func TestGet_Desc_BareValue(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.name", "--desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should include both value and description separated by tab
	if !strings.Contains(stdout, "Alexander Towell") {
		t.Errorf("expected value in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Full legal name") {
		t.Errorf("expected description in output, got %q", stdout)
	}
}

func TestGet_FormatTOML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "toml"
	stdout, _, err := executeCommand("get", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "[identity]") {
		t.Errorf("expected TOML section header, got %q", stdout)
	}
	if !strings.Contains(stdout, `name = "Alexander Towell"`) {
		t.Errorf("expected TOML key-value, got %q", stdout)
	}
}

func TestGet_FormatYAML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "yaml"
	stdout, _, err := executeCommand("get", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "identity:") {
		t.Errorf("expected YAML category, got %q", stdout)
	}
	if !strings.Contains(stdout, "name: Alexander Towell") {
		t.Errorf("expected YAML key-value, got %q", stdout)
	}
}

func TestGet_FormatEnv(t *testing.T) {
	setupTestDB(t)
	flagFormat = "env"
	stdout, _, err := executeCommand("get", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, `DEETS_IDENTITY_NAME="Alexander Towell"`) {
		t.Errorf("expected env format, got %q", stdout)
	}
}

func TestGet_GlobPattern(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("get", "*.orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "0000-0001-2345-6789") {
		t.Errorf("expected orcid value in output, got %q", stdout)
	}
}

// --- Multi-path tests ---

func TestGet_MultiPath(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("get", "identity.name", "contact.email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Multi-path always returns structured output
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Fatalf("expected valid JSON, got %q", stdout)
	}
	if !strings.Contains(stdout, "Alexander Towell") {
		t.Errorf("expected name in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "alex@example.com") {
		t.Errorf("expected email in output, got %q", stdout)
	}
}

func TestGet_MultiPath_OneNotFound(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	_, _, err := executeCommand("get", "identity.name", "nonexistent.field")
	if err == nil {
		t.Fatal("expected error when one path not found")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Errorf("expected ExitError with code 2, got %v", err)
	}
}

func TestGet_MultiPath_Default(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("get", "identity.name", "nonexistent.field", "--default", "N/A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Alexander Towell") {
		t.Errorf("expected name in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "N/A") {
		t.Errorf("expected default value in output, got %q", stdout)
	}
}

func TestGet_MultiPath_Exists_AllFound(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.name", "contact.email", "--exists")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("--exists should produce no output, got %q", stdout)
	}
}

func TestGet_MultiPath_Exists_OneMissing(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("get", "identity.name", "nonexistent.field", "--exists")
	if err == nil {
		t.Fatal("expected error when one path missing with --exists")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Errorf("expected ExitError with code 2, got %v", err)
	}
}
