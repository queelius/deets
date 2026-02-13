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

// --- findKey prefix-safety tests ---

func TestFindKey_DoesNotMatchPrefix(t *testing.T) {
	lines := []string{
		"[identity]",
		`name_desc = "Full legal name"`,
		`name = "Alice"`,
	}

	// Searching for "name" should find line 2 (name = "Alice"), not line 1 (name_desc).
	idx := findKey(lines, 1, 3, "name")
	if idx != 2 {
		t.Errorf("expected index 2 for 'name', got %d", idx)
	}

	// Searching for "name_desc" should find line 1.
	idx = findKey(lines, 1, 3, "name_desc")
	if idx != 1 {
		t.Errorf("expected index 1 for 'name_desc', got %d", idx)
	}
}

func TestFindKey_DoesNotMatchShorterPrefix(t *testing.T) {
	lines := []string{
		"[contact]",
		`email2 = "alt@example.com"`,
		`email = "main@example.com"`,
	}

	// Searching for "email" must not match "email2".
	idx := findKey(lines, 1, 3, "email")
	if idx != 2 {
		t.Errorf("expected index 2 for 'email', got %d", idx)
	}
}

func TestFindKey_VariousSpacing(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{`name = "val"`, true},
		{`name="val"`, true},
		{`name  =  "val"`, true},
		{`name	= "val"`, true}, // tab before =
	}

	for _, tt := range tests {
		lines := []string{tt.line}
		idx := findKey(lines, 0, 1, "name")
		got := idx == 0
		if got != tt.want {
			t.Errorf("findKey(%q, 'name') = %v, want %v", tt.line, got, tt.want)
		}
	}
}

func TestSetValue_DoesNotClobberSimilarKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := `[identity]
name_desc = "Full legal name"
name = "Alice"
`
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Setting "name" should not affect "name_desc".
	if err := SetValue(path, "identity", "name", "Bob"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `name_desc = "Full legal name"`) {
		t.Errorf("name_desc should be preserved, got:\n%s", content)
	}
	if !strings.Contains(content, `name = "Bob"`) {
		t.Errorf("name should be updated, got:\n%s", content)
	}
}

// --- formatValue / isArrayLiteral tests ---

func TestFormatValue_BracketString(t *testing.T) {
	// "[hello]" is NOT an array — it's a string that starts with '['.
	result := formatValue("[hello]")
	if result != `"[hello]"` {
		t.Errorf("expected quoted string, got %q", result)
	}
}

func TestFormatValue_EmptyArray(t *testing.T) {
	result := formatValue("[]")
	if result != "[]" {
		t.Errorf("expected empty array as-is, got %q", result)
	}
}

func TestFormatValue_RealArray(t *testing.T) {
	result := formatValue(`["a", "b"]`)
	if result != `["a", "b"]` {
		t.Errorf("expected real array as-is, got %q", result)
	}
}

func TestIsArrayLiteral(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`["a", "b"]`, true},
		{`[]`, true},
		{`[1, 2, 3]`, true},
		{`[123]`, true},
		{`[-1, 0, 1]`, true},
		{`[true, false]`, true},
		{`[ "a", "b"]`, true},  // leading whitespace
		{`[hello]`, false},     // bare word — string, not array
		{`[Hello World]`, false},
		{`[`, false},
		{`]`, false},
		{`""`, false},
		{``, false},
	}
	for _, tt := range tests {
		got := isArrayLiteral(tt.input)
		if got != tt.want {
			t.Errorf("isArrayLiteral(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSetValue_BracketStringProducesValidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := "[identity]\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	if err := SetValue(path, "identity", "motto", "[hello]"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `motto = "[hello]"`) {
		t.Errorf("expected quoted bracket string, got:\n%s", content)
	}

	// Verify it's valid TOML by round-tripping through Load.
	db, err := Load(path, "")
	if err != nil {
		t.Fatalf("produced invalid TOML: %v", err)
	}
	f, ok := db.GetField("identity.motto")
	if !ok {
		t.Fatal("expected identity.motto field")
	}
	if f.Value != "[hello]" {
		t.Errorf("expected value [hello], got %q", f.Value)
	}
}

// --- ValidateName tests ---

func TestValidateName_ValidNames(t *testing.T) {
	valid := []string{"name", "research_interests", "my-key", "Key1", "a", "A-Z_09"}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateName_InvalidNames(t *testing.T) {
	invalid := []string{
		"",           // empty
		"a.b",        // dot
		"a b",        // space
		"evil]",      // bracket
		"[evil",      // bracket
		"new\nline",  // newline
		"tab\there",  // tab
		"café",       // unicode
	}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
}

func TestSetValue_RejectsInvalidCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	err := SetValue(path, "evil]", "name", "x")
	if err == nil {
		t.Fatal("expected error for invalid category, got nil")
	}
	if !strings.Contains(err.Error(), "invalid category") {
		t.Errorf("error should mention invalid category, got: %v", err)
	}
}

func TestSetValue_RejectsInvalidKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	err := SetValue(path, "identity", "a.b", "x")
	if err == nil {
		t.Fatal("expected error for invalid key, got nil")
	}
	if !strings.Contains(err.Error(), "invalid key") {
		t.Errorf("error should mention invalid key, got: %v", err)
	}
}

func TestRemoveValue_RejectsInvalidNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")
	os.WriteFile(path, []byte("[identity]\nname = \"x\"\n"), 0644)

	if err := RemoveValue(path, "evil]", "name"); err == nil {
		t.Fatal("expected error for invalid category")
	}
	if err := RemoveValue(path, "identity", "a b"); err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestSetValue_HashCharacterInValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := "[identity]\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// '#' in a value could be misinterpreted as an inline TOML comment.
	// formatValue uses %q which properly escapes it inside double quotes.
	if err := SetValue(path, "identity", "motto", "C# is great"); err != nil {
		t.Fatalf("SetValue returned error: %v", err)
	}

	// Verify it round-trips through Load.
	db, err := Load(path, "")
	if err != nil {
		t.Fatalf("produced invalid TOML: %v", err)
	}
	f, ok := db.GetField("identity.motto")
	if !ok {
		t.Fatal("expected identity.motto field")
	}
	if f.Value != "C# is great" {
		t.Errorf("expected 'C# is great', got %q", f.Value)
	}
}

func TestSetValue_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	initial := "[identity]\nname = \"Alice\"\n"
	if err := os.WriteFile(path, []byte(initial), 0644); err != nil {
		t.Fatal(err)
	}

	// Set the same value twice — file should not get corrupted.
	for i := 0; i < 2; i++ {
		if err := SetValue(path, "identity", "name", "Alice"); err != nil {
			t.Fatalf("SetValue #%d returned error: %v", i+1, err)
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Count(content, "name") != 1 {
		t.Errorf("expected exactly one 'name' entry, got:\n%s", content)
	}
}

func TestRemoveCategory_RejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")
	os.WriteFile(path, []byte("[identity]\nname = \"x\"\n"), 0644)

	if err := RemoveCategory(path, "evil]"); err == nil {
		t.Fatal("expected error for invalid category")
	}
}

// --- ValidateCategoryName tests ---

func TestValidateCategoryName_DottedNames(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"identity", false},
		{"profiles.github", false},
		{"profiles.github.sub", false},
		{"a.b.c", false},
		{"", true},
		{".github", true},
		{"profiles.", true},
		{"profiles..github", true},
		{"profiles.git hub", true},
		{"profiles.evil]", true},
	}
	for _, tt := range tests {
		err := ValidateCategoryName(tt.name)
		if (err != nil) != tt.wantErr {
			t.Errorf("ValidateCategoryName(%q) error=%v, wantErr=%v", tt.name, err, tt.wantErr)
		}
	}
}

func TestSetValue_DottedCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	err := SetValue(path, "profiles.github", "username", "queelius")
	if err != nil {
		t.Fatalf("SetValue with dotted category failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "[profiles.github]") {
		t.Errorf("expected [profiles.github] section header, got:\n%s", content)
	}
	if !strings.Contains(content, `username = "queelius"`) {
		t.Errorf("expected username field, got:\n%s", content)
	}
}

func TestRemoveValue_DottedCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")
	content := "[profiles.github]\nusername = \"queelius\"\nemail = \"q@g.com\"\n"
	os.WriteFile(path, []byte(content), 0644)

	err := RemoveValue(path, "profiles.github", "email")
	if err != nil {
		t.Fatalf("RemoveValue with dotted category failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "email") {
		t.Errorf("expected email removed, got:\n%s", string(data))
	}
}

func TestRemoveCategory_DottedCategory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")
	content := "[identity]\nname = \"Alice\"\n\n[profiles.github]\nusername = \"alice\"\n"
	os.WriteFile(path, []byte(content), 0644)

	err := RemoveCategory(path, "profiles.github")
	if err != nil {
		t.Fatalf("RemoveCategory with dotted category failed: %v", err)
	}

	data, _ := os.ReadFile(path)
	if strings.Contains(string(data), "profiles.github") {
		t.Errorf("expected profiles.github removed, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "[identity]") {
		t.Errorf("expected identity preserved, got:\n%s", string(data))
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
