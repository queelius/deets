package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRm_SingleField(t *testing.T) {
	home := setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "contact.email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the field is gone.
	data, err := os.ReadFile(filepath.Join(home, ".deets", "me.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, `email = "alex@example.com"`) {
		t.Error("removed field should not appear in file")
	}
	// Other fields should remain.
	if !strings.Contains(content, `name = "Alexander Towell"`) {
		t.Error("other fields should be preserved")
	}
}

func TestRm_EntireCategory(t *testing.T) {
	home := setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".deets", "me.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "[web]") {
		t.Error("removed category should not appear")
	}
	if !strings.Contains(content, "[identity]") {
		t.Error("other categories should remain")
	}
}

func TestRm_NonexistentField(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "identity.nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent field")
	}
}

func TestRm_NonexistentCategory(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent category")
	}
}

func TestRm_LastFieldRemovesSection(t *testing.T) {
	home := setupTestEnv(t)

	deetsDir := filepath.Join(home, ".deets")
	if err := os.MkdirAll(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}

	toml := `[solo]
only = "value"

[other]
key = "val"
`
	if err := os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err := executeCommand("rm", "solo.only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "[solo]") {
		t.Error("empty section should be removed")
	}
	if !strings.Contains(content, "[other]") {
		t.Error("other sections should remain")
	}
}

func TestRm_FieldLeavesDescOrphaned(t *testing.T) {
	home := setupTestDB(t)

	// Remove email — email_desc should remain (rm only removes what you ask).
	_, _, err := executeCommand("rm", "contact.email")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(home, ".deets", "me.toml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, `email = "alex@example.com"`) {
		t.Error("email field should be removed")
	}
	// The _desc companion is a separate key — rm doesn't cascade.
	if !strings.Contains(content, "email_desc") {
		t.Error("email_desc should remain (rm does not cascade to companion keys)")
	}
}

func TestRm_MultiDotPath(t *testing.T) {
	setupTestDB(t)

	// "a.b.c" → parsePath splits on last dot to category="a.b", key="c".
	// This is a valid path, but the category doesn't exist in the test DB.
	_, _, err := executeCommand("rm", "a.b.c")
	if err == nil {
		t.Fatal("expected error for nonexistent dotted category")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestRm_InvalidCategoryName(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "evil]")
	if err == nil {
		t.Fatal("expected error for invalid category name")
	}
	if !strings.Contains(err.Error(), "invalid") {
		t.Errorf("error should mention invalid, got: %v", err)
	}
}

func TestRm_InvalidFieldPath(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"

	_, _, err := executeCommand("rm", "a b.key")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestRm_DottedCategory(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	toml := "[identity]\nname = \"Alice\"\n\n[platforms.github]\nurl = \"https://github.com/alice\"\n"
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644)

	_, _, err := executeCommand("rm", "platforms.github")
	if err != nil {
		t.Fatalf("rm platforms.github failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if strings.Contains(string(data), "platforms.github") {
		t.Errorf("expected platforms.github removed, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "[identity]") {
		t.Errorf("expected identity preserved, got:\n%s", string(data))
	}
}

func TestRm_DottedCategoryField(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	toml := "[platforms.github]\nurl = \"https://github.com/alice\"\nbio = \"Developer\"\n"
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644)

	_, _, err := executeCommand("rm", "platforms.github.bio")
	if err != nil {
		t.Fatalf("rm platforms.github.bio failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if strings.Contains(string(data), "bio") {
		t.Errorf("expected bio removed, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "url") {
		t.Errorf("expected url preserved, got:\n%s", string(data))
	}
}
