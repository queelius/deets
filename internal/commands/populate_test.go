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

// --- GitHub populate tests (unit tests for parseGitHubUser) ---

func TestPopulateGitHub_Mapping(t *testing.T) {
	ghJSON := `{"login":"alice","name":"Alice Smith","email":"alice@gh.com","bio":"Developer","blog":"https://alice.dev","html_url":"https://github.com/alice"}`
	entries, err := parseGitHubUser([]byte(ghJSON))
	if err != nil {
		t.Fatalf("parseGitHubUser: %v", err)
	}
	found := make(map[string]string)
	for _, e := range entries {
		found[e.category+"."+e.key] = e.value
	}
	if found["profiles.github.username"] != "alice" {
		t.Errorf("expected username=alice, got %q", found["profiles.github.username"])
	}
	if found["profiles.github.name"] != "Alice Smith" {
		t.Errorf("expected name='Alice Smith', got %q", found["profiles.github.name"])
	}
	if found["profiles.github.email"] != "alice@gh.com" {
		t.Errorf("expected email, got %q", found["profiles.github.email"])
	}
	if found["profiles.github.url"] != "https://github.com/alice" {
		t.Errorf("expected url, got %q", found["profiles.github.url"])
	}
	if found["profiles.github.bio"] != "Developer" {
		t.Errorf("expected bio, got %q", found["profiles.github.bio"])
	}
	if found["profiles.blog.url"] != "https://alice.dev" {
		t.Errorf("expected blog url, got %q", found["profiles.blog.url"])
	}
	// Also populates identity and contact
	if found["identity.name"] != "Alice Smith" {
		t.Errorf("expected identity.name, got %q", found["identity.name"])
	}
	if found["contact.email"] != "alice@gh.com" {
		t.Errorf("expected contact.email, got %q", found["contact.email"])
	}
}

func TestPopulateGitHub_PartialData(t *testing.T) {
	// Minimal JSON with only login
	ghJSON := `{"login":"alice","html_url":"https://github.com/alice"}`
	entries, err := parseGitHubUser([]byte(ghJSON))
	if err != nil {
		t.Fatalf("parseGitHubUser: %v", err)
	}
	found := make(map[string]string)
	for _, e := range entries {
		found[e.category+"."+e.key] = e.value
	}
	if found["profiles.github.username"] != "alice" {
		t.Errorf("expected username=alice, got %q", found["profiles.github.username"])
	}
	// No name, email, bio, blog -> should NOT be in entries
	if _, ok := found["profiles.github.name"]; ok {
		t.Error("empty name should not produce an entry")
	}
	if _, ok := found["profiles.github.email"]; ok {
		t.Error("empty email should not produce an entry")
	}
	if _, ok := found["profiles.github.bio"]; ok {
		t.Error("empty bio should not produce an entry")
	}
	if _, ok := found["profiles.blog.url"]; ok {
		t.Error("empty blog should not produce an entry")
	}
	// identity.name and contact.email should also be absent
	if _, ok := found["identity.name"]; ok {
		t.Error("empty name should not produce identity.name entry")
	}
	if _, ok := found["contact.email"]; ok {
		t.Error("empty email should not produce contact.email entry")
	}
}

func TestPopulateGitHub_InvalidJSON(t *testing.T) {
	_, err := parseGitHubUser([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestPopulateGitHub_EmptyJSON(t *testing.T) {
	entries, err := parseGitHubUser([]byte("{}"))
	if err != nil {
		t.Fatalf("parseGitHubUser: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected no entries for empty JSON, got %d", len(entries))
	}
}

func TestPopulateGitHub_EntryCount(t *testing.T) {
	// Full profile should produce exactly 8 entries:
	// profiles.github.username, profiles.github.name, identity.name,
	// profiles.github.email, contact.email, profiles.github.url,
	// profiles.github.bio, profiles.blog.url
	ghJSON := `{"login":"alice","name":"Alice","email":"a@b.com","bio":"Dev","blog":"https://a.dev","html_url":"https://github.com/alice"}`
	entries, err := parseGitHubUser([]byte(ghJSON))
	if err != nil {
		t.Fatalf("parseGitHubUser: %v", err)
	}
	if len(entries) != 8 {
		t.Errorf("expected 8 entries for full profile, got %d", len(entries))
		for _, e := range entries {
			t.Logf("  %s.%s = %s", e.category, e.key, e.value)
		}
	}
}
