package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPopulate_GitDryRun(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	// Set up git config in temp HOME
	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--git", "--dry-run")
	if err != nil {
		t.Fatalf("populate --git --dry-run failed: %v", err)
	}
	if !strings.Contains(stdout, "identity.name") {
		t.Errorf("expected identity.name in dry-run output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Test User") {
		t.Errorf("expected 'Test User' in dry-run output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "contact.email") {
		t.Errorf("expected contact.email in dry-run output, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "test@example.com") {
		t.Errorf("expected 'test@example.com' in dry-run output, got:\n%s", stdout)
	}

	// Verify nothing was written (dry-run should not modify the file)
	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if strings.Contains(string(data), "Test User") {
		t.Error("dry-run should not write to file")
	}
}

func TestPopulate_GitYes(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	// Set up git config
	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	_, _, err := executeCommand("populate", "--git", "--yes")
	if err != nil {
		t.Fatalf("populate --git --yes failed: %v", err)
	}

	// Verify values were written
	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	content := string(data)
	if !strings.Contains(content, `name = "Test User"`) {
		t.Errorf("expected name written, got:\n%s", content)
	}
	if !strings.Contains(content, `email = "test@example.com"`) {
		t.Errorf("expected email written, got:\n%s", content)
	}
}

func TestPopulate_GitSkipsUnchanged(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\nname = \"Test User\"\n"), 0644)

	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--git", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// name should not appear as a change (already set to same value)
	if strings.Contains(stdout, "~ identity.name") {
		t.Error("unchanged name should not show as change")
	}
	// email should appear as a new field
	if !strings.Contains(stdout, "contact.email") {
		t.Error("new email should appear in output")
	}
}

func TestPopulate_GitShowsUpdate(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\nname = \"Old Name\"\n"), 0644)

	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = New Name\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--git", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// name should appear as a change since the value differs
	if !strings.Contains(stdout, "~ identity.name") {
		t.Errorf("changed name should show as update, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "Old Name") {
		t.Errorf("should show old value, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "New Name") {
		t.Errorf("should show new value, got:\n%s", stdout)
	}
}

func TestPopulate_NoSources(t *testing.T) {
	setupTestEnv(t)
	_, _, err := executeCommand("populate")
	if err == nil {
		t.Fatal("expected error when no source specified")
	}
	if !strings.Contains(err.Error(), "no source") {
		t.Errorf("error should mention no source, got: %v", err)
	}
}

func TestPopulate_AllFlag(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	// Set up git config
	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--all", "--dry-run")
	if err != nil {
		t.Fatalf("populate --all --dry-run failed: %v", err)
	}
	// Should at least show git-harvested data
	if !strings.Contains(stdout, "identity.name") {
		t.Errorf("expected identity.name with --all, got:\n%s", stdout)
	}
}

func TestPopulate_GitYesApplied(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--git", "--yes")
	if err != nil {
		t.Fatalf("populate --git --yes failed: %v", err)
	}
	if !strings.Contains(stdout, "Applied") {
		t.Errorf("expected 'Applied' in output, got:\n%s", stdout)
	}
}

func TestPopulate_NoChanges(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\nname = \"Test User\"\n\n[contact]\nemail = \"test@example.com\"\n"), 0644)

	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	stdout, _, err := executeCommand("populate", "--git", "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "No changes") {
		t.Errorf("expected 'No changes' when all values match, got:\n%s", stdout)
	}
}

func TestPopulate_GitLocalFlag(t *testing.T) {
	home := setupTestEnv(t)

	// Create global deets so loadDB works
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	// Set up git config
	os.WriteFile(filepath.Join(home, ".gitconfig"), []byte("[user]\n\tname = Test User\n\temail = test@example.com\n"), 0644)

	_, _, err := executeCommand("populate", "--git", "--yes", "--local")
	if err != nil {
		t.Fatalf("populate --git --yes --local failed: %v", err)
	}

	// Local file should be written in cwd/.deets/me.toml
	localPath := filepath.Join(home, ".deets", "me.toml")
	data, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("expected local file to be written: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Test User") {
		t.Errorf("expected local file to contain Test User, got:\n%s", content)
	}
}
