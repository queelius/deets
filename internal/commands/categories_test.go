package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategories_Table(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("categories")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Test data has: academic, contact, identity, web (alphabetical).
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 categories, got %d: %v", len(lines), lines)
	}
	// Should be alphabetically sorted.
	expected := []string{"academic", "contact", "identity", "web"}
	for i, want := range expected {
		got := strings.TrimSpace(lines[i])
		if got != want {
			t.Errorf("line %d: expected %q, got %q", i, want, got)
		}
	}
}

func TestCategories_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("categories")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var names []string
	if err := json.Unmarshal([]byte(stdout), &names); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(names) != 4 {
		t.Errorf("expected 4 categories, got %d", len(names))
	}
}

func TestCategories_EmptyDB(t *testing.T) {
	home := setupTestEnv(t)

	deetsDir := filepath.Join(home, ".deets")
	if err := os.MkdirAll(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("# empty\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagFormat = "table"
	stdout, _, err := executeCommand("categories")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("expected empty output for empty DB, got %q", stdout)
	}
}

func TestCategories_EmptyDB_JSON(t *testing.T) {
	home := setupTestEnv(t)

	deetsDir := filepath.Join(home, ".deets")
	if err := os.MkdirAll(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("# empty\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagFormat = "json"
	stdout, _, err := executeCommand("categories")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	trimmed := strings.TrimSpace(stdout)
	if trimmed != "[]" {
		t.Errorf("expected empty JSON array, got %q", trimmed)
	}
}
