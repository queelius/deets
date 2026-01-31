package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- SetValue tests ---

func TestSetValue_NewFileCreation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.toml")

	if err := SetValue(path, "identity", "name", "Alice"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading created file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "[identity]") {
		t.Error("expected [identity] section header in output")
	}
	if !strings.Contains(content, `name = "Alice"`) {
		t.Errorf("expected name = \"Alice\" in output, got:\n%s", content)
	}
}

func TestSetValue_AddToExistingSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "identity", "pronouns", "she/her"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `name = "Alice"`) {
		t.Error("existing key should be preserved")
	}
	if !strings.Contains(content, `pronouns = "she/her"`) {
		t.Errorf("expected new key in output, got:\n%s", content)
	}

	// The section header should appear only once.
	if strings.Count(content, "[identity]") != 1 {
		t.Error("section header should appear exactly once")
	}
}

func TestSetValue_AddNewSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "contact", "email", "alice@example.com"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "[identity]") {
		t.Error("original section should be preserved")
	}
	if !strings.Contains(content, "[contact]") {
		t.Error("new section should be added")
	}
	if !strings.Contains(content, `email = "alice@example.com"`) {
		t.Errorf("expected email key, got:\n%s", content)
	}
}

func TestSetValue_ReplaceExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
pronouns = "she/her"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "identity", "name", "Bob"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, `"Alice"`) {
		t.Error("old value 'Alice' should be replaced")
	}
	if !strings.Contains(content, `name = "Bob"`) {
		t.Errorf("expected name = \"Bob\", got:\n%s", content)
	}
	// pronouns should be untouched.
	if !strings.Contains(content, `pronouns = "she/her"`) {
		t.Error("other keys should be preserved")
	}
}

func TestSetValue_ArrayValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Array values start with "[" and should be written as-is.
	if err := SetValue(path, "identity", "aka", `["Nick", "Nickname"]`); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `aka = ["Nick", "Nickname"]`) {
		t.Errorf("expected array value written as-is, got:\n%s", content)
	}
}

func TestSetValue_QuotedValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Already-quoted values should be written as-is.
	if err := SetValue(path, "identity", "motto", `"To be or not to be"`); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `motto = "To be or not to be"`) {
		t.Errorf("expected pre-quoted value, got:\n%s", content)
	}
}

func TestSetValue_PreservesComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `# This is a comment about identity
[identity]
# Name comment
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "identity", "pronouns", "she/her"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "# This is a comment about identity") {
		t.Error("top-level comment should be preserved")
	}
	if !strings.Contains(content, "# Name comment") {
		t.Error("inline comment should be preserved")
	}
}

func TestSetValue_AddToExistingSectionWithMultipleSections(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"

[contact]
email = "alice@example.com"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Add a new key to the first section.
	if err := SetValue(path, "identity", "pronouns", "she/her"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	// Both sections should still exist.
	if !strings.Contains(content, "[identity]") || !strings.Contains(content, "[contact]") {
		t.Error("both sections should be preserved")
	}

	// The new key should be in the identity section (before [contact]).
	idxPronouns := strings.Index(content, "pronouns")
	idxContact := strings.Index(content, "[contact]")
	if idxPronouns == -1 {
		t.Fatal("pronouns key not found")
	}
	if idxPronouns > idxContact {
		t.Error("pronouns should be inserted before [contact] section")
	}
}

func TestSetValue_EmptyExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "identity", "name", "Alice"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, "[identity]") {
		t.Error("section header should be present")
	}
	if !strings.Contains(content, `name = "Alice"`) {
		t.Errorf("key should be present, got:\n%s", content)
	}
}

// --- RemoveValue tests ---

func TestRemoveValue_RemoveExistingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
pronouns = "she/her"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveValue(path, "identity", "name"); err != nil {
		t.Fatalf("RemoveValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "name") {
		t.Errorf("removed key 'name' should not appear, got:\n%s", content)
	}
	if !strings.Contains(content, `pronouns = "she/her"`) {
		t.Error("other keys should be preserved")
	}
	if !strings.Contains(content, "[identity]") {
		t.Error("section header should remain since section is not empty")
	}
}

func TestRemoveValue_SectionBecomesEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"

[contact]
email = "alice@example.com"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveValue(path, "identity", "name"); err != nil {
		t.Fatalf("RemoveValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	// The identity section should be removed entirely since it's now empty.
	if strings.Contains(content, "[identity]") {
		t.Error("empty section should be removed")
	}
	// Contact section should remain.
	if !strings.Contains(content, "[contact]") {
		t.Error("other section should be preserved")
	}
	if !strings.Contains(content, `email = "alice@example.com"`) {
		t.Error("other section's keys should be preserved")
	}
}

func TestRemoveValue_KeyNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := RemoveValue(path, "identity", "nonexistent")
	if err == nil {
		t.Fatal("expected error for key not found, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the missing key, got: %v", err)
	}
}

func TestRemoveValue_CategoryNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := RemoveValue(path, "nonexistent", "name")
	if err == nil {
		t.Fatal("expected error for category not found, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the missing category, got: %v", err)
	}
}

func TestRemoveValue_FileNotFound(t *testing.T) {
	err := RemoveValue("/nonexistent/path/me.toml", "identity", "name")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// --- RemoveCategory tests ---

func TestRemoveCategory_RemoveExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
pronouns = "she/her"

[contact]
email = "alice@example.com"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveCategory(path, "identity"); err != nil {
		t.Fatalf("RemoveCategory returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if strings.Contains(content, "[identity]") {
		t.Error("removed category section should not appear")
	}
	if strings.Contains(content, "name") {
		t.Error("removed category keys should not appear")
	}
	if !strings.Contains(content, "[contact]") {
		t.Error("other category should be preserved")
	}
	if !strings.Contains(content, `email = "alice@example.com"`) {
		t.Error("other category's keys should be preserved")
	}
}

func TestRemoveCategory_NotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	err := RemoveCategory(path, "nonexistent")
	if err == nil {
		t.Fatal("expected error for category not found, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention missing category, got: %v", err)
	}
}

func TestRemoveCategory_FileNotFound(t *testing.T) {
	err := RemoveCategory("/nonexistent/path/me.toml", "identity")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestRemoveCategory_OnlyCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveCategory(path, "identity"); err != nil {
		t.Fatalf("RemoveCategory returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := strings.TrimSpace(string(data))
	if content != "" {
		t.Errorf("file should be empty after removing only category, got:\n%s", content)
	}
}

// --- formatValue tests ---

func TestFormatValue_PlainString(t *testing.T) {
	result := formatValue("hello")
	if result != `"hello"` {
		t.Errorf("expected quoted string, got %q", result)
	}
}

func TestFormatValue_ArrayLiteral(t *testing.T) {
	result := formatValue(`["a", "b"]`)
	if result != `["a", "b"]` {
		t.Errorf("expected array as-is, got %q", result)
	}
}

func TestFormatValue_AlreadyQuoted(t *testing.T) {
	result := formatValue(`"already quoted"`)
	if result != `"already quoted"` {
		t.Errorf("expected already-quoted as-is, got %q", result)
	}
}

// --- Helper function tests ---

func TestFindSection(t *testing.T) {
	lines := []string{"[identity]", "name = \"Alice\"", "", "[contact]", "email = \"a@b.com\""}

	idx := findSection(lines, "identity")
	if idx != 0 {
		t.Errorf("expected index 0 for [identity], got %d", idx)
	}

	idx = findSection(lines, "contact")
	if idx != 3 {
		t.Errorf("expected index 3 for [contact], got %d", idx)
	}

	idx = findSection(lines, "nonexistent")
	if idx != -1 {
		t.Errorf("expected -1 for nonexistent section, got %d", idx)
	}
}

func TestFindNextSection(t *testing.T) {
	lines := []string{"[identity]", "name = \"Alice\"", "", "[contact]", "email = \"a@b.com\""}

	idx := findNextSection(lines, 0)
	if idx != 3 {
		t.Errorf("expected next section at 3, got %d", idx)
	}

	// After last section, should return len(lines).
	idx = findNextSection(lines, 3)
	if idx != len(lines) {
		t.Errorf("expected len(lines) = %d, got %d", len(lines), idx)
	}
}

func TestFindKey(t *testing.T) {
	lines := []string{"[identity]", "name = \"Alice\"", "pronouns = \"she/her\""}

	idx := findKey(lines, 1, 3, "name")
	if idx != 1 {
		t.Errorf("expected index 1 for 'name', got %d", idx)
	}

	idx = findKey(lines, 1, 3, "pronouns")
	if idx != 2 {
		t.Errorf("expected index 2 for 'pronouns', got %d", idx)
	}

	idx = findKey(lines, 1, 3, "nonexistent")
	if idx != -1 {
		t.Errorf("expected -1 for nonexistent key, got %d", idx)
	}
}

func TestReadLines_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := readLines(path)
	if err != nil {
		t.Fatalf("readLines returned error: %v", err)
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty file, got %d", len(lines))
	}
}

func TestReadLines_WithContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.toml")

	if err := os.WriteFile(path, []byte("[identity]\nname = \"Alice\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	lines, err := readLines(path)
	if err != nil {
		t.Fatalf("readLines returned error: %v", err)
	}
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestWriteLines_AppendsNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.toml")

	if err := writeLines(path, []string{"[identity]", "name = \"Alice\""}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.HasSuffix(content, "\n") {
		t.Error("output should end with a newline")
	}

	expected := "[identity]\nname = \"Alice\"\n"
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}
