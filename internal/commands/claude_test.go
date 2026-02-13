package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaude_Install(t *testing.T) {
	home := setupTestEnv(t)

	stdout, _, err := executeCommand("claude", "install")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillPath := filepath.Join(home, ".claude", "skills", "deets", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected skill file at %s", skillPath)
	}

	if !strings.Contains(stdout, "Installed") {
		t.Errorf("expected install confirmation, got %q", stdout)
	}
}

func TestClaude_Uninstall(t *testing.T) {
	home := setupTestEnv(t)

	// First install.
	_, _, err := executeCommand("claude", "install")
	if err != nil {
		t.Fatalf("install error: %v", err)
	}

	// Then uninstall.
	stdout, _, err := executeCommand("claude", "uninstall")
	if err != nil {
		t.Fatalf("uninstall error: %v", err)
	}

	skillDir := filepath.Join(home, ".claude", "skills", "deets")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("skill directory should be removed after uninstall")
	}

	if !strings.Contains(stdout, "Removed") {
		t.Errorf("expected removal confirmation, got %q", stdout)
	}
}

func TestClaude_Install_CleansOldFormat(t *testing.T) {
	home := setupTestEnv(t)

	// Create old flat-file skill.
	oldDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "deets.md")
	if err := os.WriteFile(oldPath, []byte("old skill content"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeCommand("claude", "install")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Old flat file should be removed.
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old flat-file skill should be cleaned up")
	}

	// New skill should exist.
	newPath := filepath.Join(home, ".claude", "skills", "deets", "SKILL.md")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("new skill file should be created")
	}
}

func TestClaude_Install_Idempotent(t *testing.T) {
	setupTestEnv(t)

	// Install twice — should not error.
	_, _, err := executeCommand("claude", "install")
	if err != nil {
		t.Fatalf("first install error: %v", err)
	}

	_, _, err = executeCommand("claude", "install")
	if err != nil {
		t.Fatalf("second install error: %v", err)
	}
}

func TestClaude_Uninstall_NoDir(t *testing.T) {
	setupTestEnv(t)

	// Uninstall when no skill exists — should not error.
	stdout, _, err := executeCommand("claude", "uninstall")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No skill directory") {
		t.Errorf("expected 'no skill directory' message, got %q", stdout)
	}
}

func TestClaude_Install_Local(t *testing.T) {
	home := setupTestEnv(t)

	stdout, _, err := executeCommand("claude", "install", "--local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	skillPath := filepath.Join(home, ".claude", "skills", "deets", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected local skill file at %s", skillPath)
	}

	if !strings.Contains(stdout, "Installed") {
		t.Errorf("expected install confirmation, got %q", stdout)
	}
}

func TestClaude_Uninstall_Local(t *testing.T) {
	home := setupTestEnv(t)

	// Install locally first.
	_, _, err := executeCommand("claude", "install", "--local")
	if err != nil {
		t.Fatalf("install error: %v", err)
	}

	// Uninstall.
	stdout, _, err := executeCommand("claude", "uninstall", "--local")
	if err != nil {
		t.Fatalf("uninstall error: %v", err)
	}

	skillDir := filepath.Join(home, ".claude", "skills", "deets")
	if _, err := os.Stat(skillDir); !os.IsNotExist(err) {
		t.Error("local skill directory should be removed after uninstall")
	}

	if !strings.Contains(stdout, "Removed") {
		t.Errorf("expected removal confirmation, got %q", stdout)
	}
}

func TestClaude_Install_Quiet(t *testing.T) {
	setupTestEnv(t)

	stdout, _, err := executeCommand("claude", "install", "--quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stdout != "" {
		t.Errorf("quiet mode should suppress output, got %q", stdout)
	}
}

func TestClaude_Uninstall_CleansOldFormat(t *testing.T) {
	home := setupTestEnv(t)

	// Create old flat-file skill.
	oldDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "deets.md")
	if err := os.WriteFile(oldPath, []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, err := executeCommand("claude", "uninstall")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("old flat-file should be cleaned up")
	}

	if !strings.Contains(stdout, "Removed old skill file") {
		t.Errorf("expected old file removal message, got %q", stdout)
	}
}
