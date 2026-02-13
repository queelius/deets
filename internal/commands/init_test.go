package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/queelius/deets/internal/store"
)

func TestInit_Global(t *testing.T) {
	home := setupTestEnv(t)

	stdout, _, err := executeCommand("init")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(home, ".deets", "me.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("init should create file: %v", err)
	}

	content := string(data)
	if content != store.DefaultTemplate {
		t.Errorf("file content should match DefaultTemplate")
	}

	if !strings.Contains(stdout, "Created") {
		t.Errorf("expected confirmation message, got %q", stdout)
	}
}

func TestInit_Local(t *testing.T) {
	home := setupTestEnv(t)

	stdout, _, err := executeCommand("init", "--local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	path := filepath.Join(home, ".deets", "me.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("init --local should create file: %v", err)
	}

	content := string(data)
	if content != store.LocalTemplate {
		t.Errorf("file content should match LocalTemplate")
	}

	if !strings.Contains(stdout, "Created") {
		t.Errorf("expected confirmation message, got %q", stdout)
	}
}

func TestInit_AlreadyExists(t *testing.T) {
	setupTestDB(t) // creates ~/.deets/me.toml

	_, _, err := executeCommand("init")
	if err == nil {
		t.Fatal("expected error when file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention already exists, got: %v", err)
	}
}

func TestInit_Quiet(t *testing.T) {
	setupTestEnv(t)

	stdout, _, err := executeCommand("init", "--quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("quiet mode should suppress output, got %q", stdout)
	}
}

func TestInit_LocalAlreadyExists(t *testing.T) {
	home := setupTestEnv(t)

	// Create the local file first.
	deetsDir := filepath.Join(home, ".deets")
	if err := os.MkdirAll(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("exists"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeCommand("init", "--local")
	if err == nil {
		t.Fatal("expected error when local file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention already exists, got: %v", err)
	}
}

func TestInit_LocalQuiet(t *testing.T) {
	setupTestEnv(t)

	stdout, _, err := executeCommand("init", "--local", "--quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("quiet mode should suppress output, got %q", stdout)
	}
}
