package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSet_BasicValue(t *testing.T) {
	setupTestDB(t)
	flagFormat = ""
	_, _, err := executeCommand("set", "identity.nickname", "Alex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the value was written
	flagFormat = "table"
	stdout, _, err := executeCommand("get", "identity.nickname")
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if strings.TrimSpace(stdout) != "Alex" {
		t.Errorf("expected 'Alex', got %q", stdout)
	}
}

func TestSet_CreateCategory(t *testing.T) {
	setupTestDB(t)
	_, _, err := executeCommand("set", "cooking.favorite", "lasagna")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	flagFormat = "table"
	stdout, _, err := executeCommand("get", "cooking.favorite")
	if err != nil {
		t.Fatalf("unexpected error reading back: %v", err)
	}
	if strings.TrimSpace(stdout) != "lasagna" {
		t.Errorf("expected 'lasagna', got %q", stdout)
	}
}

func TestSet_InvalidPath(t *testing.T) {
	setupTestDB(t)
	_, _, err := executeCommand("set", "noperiod", "val")
	if err == nil {
		t.Error("expected error for path without period")
	}
}

func TestSet_Local(t *testing.T) {
	setupTestDB(t)

	// Work in a temp dir to avoid polluting cwd
	workDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(workDir)
	defer os.Chdir(origDir)

	_, _, err := executeCommand("set", "--local", "identity.name", "Local Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the file was written to local
	localFile := filepath.Join(workDir, ".deets", "me.toml")
	data, err := os.ReadFile(localFile)
	if err != nil {
		t.Fatalf("reading local file: %v", err)
	}
	if !strings.Contains(string(data), "Local Name") {
		t.Errorf("expected 'Local Name' in local file, got %q", string(data))
	}
}
