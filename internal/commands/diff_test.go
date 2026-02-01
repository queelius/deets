package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/queelius/deets/internal/model"
)

func TestDiff_NoLocal(t *testing.T) {
	setupTestDB(t)
	// CWD is already the temp home (from setupTestEnv), no local .deets/ exists
	flagFormat = "table"
	_, _, err := executeCommand("diff")
	if err == nil {
		t.Fatal("expected error when no local file exists")
	}
	if !strings.Contains(err.Error(), "no local") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDiff_Identical(t *testing.T) {
	home := setupTestDB(t)

	// Create a subdirectory with local .deets/ identical to global
	workDir := filepath.Join(home, "project")
	os.MkdirAll(workDir, 0755)
	os.Chdir(workDir)

	localDir := filepath.Join(workDir, ".deets")
	os.MkdirAll(localDir, 0755)

	globalContent, _ := os.ReadFile(filepath.Join(home, ".deets", "me.toml"))
	os.WriteFile(filepath.Join(localDir, "me.toml"), globalContent, 0644)

	flagFormat = "table"
	flagQuiet = false
	stdout, _, err := executeCommand("diff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No differences") {
		t.Errorf("expected 'No differences', got %q", stdout)
	}
}

func TestDiff_Override(t *testing.T) {
	home := setupTestDB(t)

	workDir := filepath.Join(home, "project")
	os.MkdirAll(workDir, 0755)
	os.Chdir(workDir)

	localDir := filepath.Join(workDir, ".deets")
	os.MkdirAll(localDir, 0755)

	localContent := `[identity]
name = "Local Name"
`
	os.WriteFile(filepath.Join(localDir, "me.toml"), []byte(localContent), 0644)

	flagFormat = "json"
	stdout, _, err := executeCommand("diff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []model.DiffEntry
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 diff entry, got %d", len(entries))
	}
	if entries[0].Status != "override" {
		t.Errorf("expected 'override' status, got %q", entries[0].Status)
	}
	if entries[0].Path != "identity.name" {
		t.Errorf("expected path 'identity.name', got %q", entries[0].Path)
	}
}

func TestDiff_LocalOnly(t *testing.T) {
	home := setupTestDB(t)

	workDir := filepath.Join(home, "project")
	os.MkdirAll(workDir, 0755)
	os.Chdir(workDir)

	localDir := filepath.Join(workDir, ".deets")
	os.MkdirAll(localDir, 0755)

	localContent := `[custom]
special = "local value"
`
	os.WriteFile(filepath.Join(localDir, "me.toml"), []byte(localContent), 0644)

	flagFormat = "json"
	stdout, _, err := executeCommand("diff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var entries []model.DiffEntry
	if err := json.Unmarshal([]byte(stdout), &entries); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 diff entry, got %d", len(entries))
	}
	if entries[0].Status != "local-only" {
		t.Errorf("expected 'local-only' status, got %q", entries[0].Status)
	}
}
