package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWhich_GlobalOnly(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "(exists)") {
		t.Errorf("expected (exists) for global file, got %q", stdout)
	}
	if !strings.Contains(stdout, "Local:  none") {
		t.Errorf("expected no local override, got %q", stdout)
	}
}

func TestWhich_NoGlobalFile(t *testing.T) {
	setupTestEnv(t) // creates home but no .deets/
	flagFormat = "table"
	stdout, _, err := executeCommand("which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "(not found)") {
		t.Errorf("expected (not found) for missing global, got %q", stdout)
	}
}

func TestWhich_WithLocalOverride(t *testing.T) {
	home := setupTestDB(t)

	// The local dir is found by walking up from CWD; since CWD=home and
	// FindLocalDir stops before HOME, create a subdir with .deets.
	subDir := filepath.Join(home, "project")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	localDeetsDir := filepath.Join(subDir, ".deets")
	if err := os.MkdirAll(localDeetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(localDeetsDir, "me.toml"), []byte("[identity]\nname = \"Local\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change CWD to the subdir.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	flagFormat = "table"
	stdout, _, err := executeCommand("which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "active override") {
		t.Errorf("expected active override indication, got %q", stdout)
	}
}

func TestWhich_LocalDirNoFile(t *testing.T) {
	home := setupTestDB(t)

	// Create a subdir with .deets/ but no me.toml inside it.
	subDir := filepath.Join(home, "project")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	localDeetsDir := filepath.Join(subDir, ".deets")
	if err := os.MkdirAll(localDeetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Note: no me.toml created in the .deets dir.

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	flagFormat = "table"
	stdout, _, err := executeCommand("which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "dir exists, no me.toml") {
		t.Errorf("expected 'dir exists, no me.toml' indication, got %q", stdout)
	}
}

func TestWhich_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("which")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["global_file"]; !ok {
		t.Error("expected global_file key in JSON")
	}
	if _, ok := result["global_exists"]; !ok {
		t.Error("expected global_exists key in JSON")
	}
}
