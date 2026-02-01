package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/queelius/deets/internal/model"
)

func TestImport_IntoExisting(t *testing.T) {
	home := setupTestDB(t)
	flagFormat = ""

	// Create import file
	importContent := `[identity]
nickname = "Lex"

[newcat]
foo = "bar"
`
	importFile := filepath.Join(home, "import.toml")
	if err := os.WriteFile(importFile, []byte(importContent), 0644); err != nil {
		t.Fatalf("writing import file: %v", err)
	}

	flagQuiet = true
	_, _, err := executeCommand("import", importFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify imported fields
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.nickname")
	if err != nil {
		t.Fatalf("unexpected error getting nickname: %v", err)
	}
	if strings.TrimSpace(stdout) != "Lex" {
		t.Errorf("expected 'Lex', got %q", stdout)
	}

	stdout, _, err = executeCommand("get", "newcat.foo")
	if err != nil {
		t.Fatalf("unexpected error getting newcat.foo: %v", err)
	}
	if strings.TrimSpace(stdout) != "bar" {
		t.Errorf("expected 'bar', got %q", stdout)
	}
}

func TestImport_DryRun(t *testing.T) {
	home := setupTestDB(t)

	importContent := `[identity]
name = "Different Name"
nickname = "Lex"
`
	importFile := filepath.Join(home, "import.toml")
	if err := os.WriteFile(importFile, []byte(importContent), 0644); err != nil {
		t.Fatalf("writing import file: %v", err)
	}

	flagFormat = "json"
	stdout, _, err := executeCommand("import", importFile, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []model.DiffEntry
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(entries) < 1 {
		t.Fatal("expected at least one diff entry")
	}

	// The original name should differ, so there should be a "change" entry
	// and nickname is new, so there should be an "add" entry
	foundChange, foundAdd := false, false
	for _, e := range entries {
		if e.Path == "identity.name" && e.Status == "change" {
			foundChange = true
		}
		if e.Path == "identity.nickname" && e.Status == "add" {
			foundAdd = true
		}
	}
	if !foundChange {
		t.Error("expected 'change' entry for identity.name")
	}
	if !foundAdd {
		t.Error("expected 'add' entry for identity.nickname")
	}

	// Verify nothing was actually written (name should still be original)
	flagFormat = "table"
	stdout, _, err = executeCommand("get", "identity.name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "Alexander Towell" {
		t.Error("dry-run should not modify the database")
	}
}

func TestImport_MissingFile(t *testing.T) {
	setupTestDB(t)
	_, _, err := executeCommand("import", "/nonexistent/file.toml")
	if err == nil {
		t.Error("expected error for missing import file")
	}
}
