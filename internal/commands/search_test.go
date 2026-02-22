package commands

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestSearch_ByValue(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("search", "Alexander")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Alexander Towell") {
		t.Errorf("expected value match, got %q", stdout)
	}
}

func TestSearch_ByKey(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("search", "orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "0000-0001-2345-6789") {
		t.Errorf("expected orcid value in results, got %q", stdout)
	}
}

func TestSearch_ByDescription(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("search", "Primary email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "alex@example.com") {
		t.Errorf("expected email in results, got %q", stdout)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("search", "GITHUB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "queelius") {
		t.Errorf("expected github value in results, got %q", stdout)
	}
}

func TestSearch_NoMatch(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("search", "zzz_nonexistent_zzz")
	if err == nil {
		t.Fatal("expected error for no match")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 2 {
		t.Errorf("expected exit code 2, got %d", exitErr.Code)
	}
}

func TestSearch_FormatJSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("search", "orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Errorf("expected valid JSON, got %q", stdout)
	}
}

func TestSearch_FormatTOML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "toml"
	stdout, _, err := executeCommand("search", "orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "[academic]") {
		t.Errorf("expected TOML section header, got %q", stdout)
	}
}

func TestSearch_FormatYAML(t *testing.T) {
	setupTestDB(t)
	flagFormat = "yaml"
	stdout, _, err := executeCommand("search", "orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "academic:") {
		t.Errorf("expected YAML category, got %q", stdout)
	}
}

func TestSearch_FormatEnv(t *testing.T) {
	setupTestDB(t)
	flagFormat = "env"
	stdout, _, err := executeCommand("search", "orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "DEETS_ACADEMIC_ORCID=") {
		t.Errorf("expected env format, got %q", stdout)
	}
}

func TestSearch_MultipleMatches(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	// "example" appears in email and domains
	stdout, _, err := executeCommand("search", "example")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "alex@example.com") {
		t.Errorf("expected email in results, got %q", stdout)
	}
	if !strings.Contains(stdout, "example.com") {
		t.Errorf("expected domains in results, got %q", stdout)
	}
}

func TestSearch_ArrayElementContent(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	// "statistics" appears inside the topics array ["statistics", "machine learning"]
	stdout, _, err := executeCommand("search", "statistics")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "topics") {
		t.Errorf("expected topics field in results, got %q", stdout)
	}
}
