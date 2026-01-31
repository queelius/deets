package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile_ValidTOMLMultipleCategories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[identity]
name = "Alice"
pronouns = "she/her"

[contact]
email = "alice@example.com"

[web]
github = "alice"
blog = "https://alice.dev"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(db.Categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(db.Categories))
	}

	// Categories should be sorted alphabetically.
	expectedCats := []string{"contact", "identity", "web"}
	for i, cat := range db.Categories {
		if cat.Name != expectedCats[i] {
			t.Errorf("category[%d]: expected %q, got %q", i, expectedCats[i], cat.Name)
		}
	}

	// Check identity category has 2 fields.
	identityCat := db.Categories[1] // "identity" sorts second
	if len(identityCat.Fields) != 2 {
		t.Fatalf("expected 2 fields in identity, got %d", len(identityCat.Fields))
	}
	if identityCat.Fields[0].Key != "name" {
		t.Errorf("expected first field key 'name', got %q", identityCat.Fields[0].Key)
	}
	if identityCat.Fields[0].Value != "Alice" {
		t.Errorf("expected value 'Alice', got %v", identityCat.Fields[0].Value)
	}

	// Check contact category.
	contactCat := db.Categories[0]
	if len(contactCat.Fields) != 1 {
		t.Fatalf("expected 1 field in contact, got %d", len(contactCat.Fields))
	}
	if contactCat.Fields[0].Key != "email" {
		t.Errorf("expected key 'email', got %q", contactCat.Fields[0].Key)
	}
}

func TestLoadFile_DescCompanions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[academic]
orcid = "0000-0001-2345-6789"
orcid_desc = "My ORCID identifier"
institution = "MIT"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(db.Categories))
	}

	cat := db.Categories[0]
	if len(cat.Fields) != 2 {
		t.Fatalf("expected 2 fields (desc keys excluded), got %d", len(cat.Fields))
	}

	// Fields sorted alphabetically: institution, orcid
	if cat.Fields[0].Key != "institution" {
		t.Errorf("expected first key 'institution', got %q", cat.Fields[0].Key)
	}
	if cat.Fields[1].Key != "orcid" {
		t.Errorf("expected second key 'orcid', got %q", cat.Fields[1].Key)
	}

	// The explicit _desc should be used for orcid.
	if cat.Fields[1].Desc != "My ORCID identifier" {
		t.Errorf("expected orcid desc 'My ORCID identifier', got %q", cat.Fields[1].Desc)
	}

	// institution has no explicit _desc, so it should fall back to DefaultDescriptions.
	if cat.Fields[0].Desc != "Academic institution" {
		t.Errorf("expected institution desc from defaults 'Academic institution', got %q", cat.Fields[0].Desc)
	}
}

func TestLoadFile_UnknownKeysPreserved(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[custom]
foo = "bar"
baz = 42
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(db.Categories))
	}

	cat := db.Categories[0]
	if cat.Name != "custom" {
		t.Errorf("expected category name 'custom', got %q", cat.Name)
	}
	if len(cat.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(cat.Fields))
	}

	// Sorted: baz, foo
	if cat.Fields[0].Key != "baz" {
		t.Errorf("expected first key 'baz', got %q", cat.Fields[0].Key)
	}
	if cat.Fields[1].Key != "foo" {
		t.Errorf("expected second key 'foo', got %q", cat.Fields[1].Key)
	}

	// Unknown keys have no default description.
	if cat.Fields[0].Desc != "" {
		t.Errorf("expected empty desc for unknown key 'baz', got %q", cat.Fields[0].Desc)
	}
}

func TestLoadFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.toml")

	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(db.Categories) != 0 {
		t.Errorf("expected 0 categories for empty file, got %d", len(db.Categories))
	}
}

func TestLoadFile_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")

	content := `[identity
name = "broken`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

func TestLoadFile_MissingFile(t *testing.T) {
	_, err := LoadFile("/nonexistent/path/me.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadFile_SkipsEmptyCategories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	// A top-level key whose value is not a map should be skipped.
	content := `[identity]
name = "Alice"

[empty]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	// "empty" has no fields so it should be skipped.
	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category (empty skipped), got %d", len(db.Categories))
	}
	if db.Categories[0].Name != "identity" {
		t.Errorf("expected category 'identity', got %q", db.Categories[0].Name)
	}
}

func TestLoadFile_DefaultDescriptionFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[web]
github = "alice"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(db.Categories))
	}

	field := db.Categories[0].Fields[0]
	if field.Desc != "GitHub username" {
		t.Errorf("expected default desc 'GitHub username', got %q", field.Desc)
	}
}

func TestLoadFile_ExplicitDescOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[web]
github = "alice"
github_desc = "My GitHub handle"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	field := db.Categories[0].Fields[0]
	if field.Desc != "My GitHub handle" {
		t.Errorf("expected explicit desc 'My GitHub handle', got %q", field.Desc)
	}
}

func TestLoadFile_FieldCategorySet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[contact]
email = "test@test.com"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	field := db.Categories[0].Fields[0]
	if field.Category != "contact" {
		t.Errorf("expected field.Category 'contact', got %q", field.Category)
	}
}

func TestLoadFile_ArrayValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[identity]
aka = ["Nick", "Nickname"]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile returned error: %v", err)
	}

	field := db.Categories[0].Fields[0]
	arr, ok := field.Value.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{} value, got %T", field.Value)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 elements in array, got %d", len(arr))
	}
	if arr[0] != "Nick" || arr[1] != "Nickname" {
		t.Errorf("unexpected array values: %v", arr)
	}
}

// --- Load tests ---

func TestLoad_GlobalOnly(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")

	content := `[identity]
name = "Alice"

[contact]
email = "alice@example.com"
`
	if err := os.WriteFile(globalPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := Load(globalPath, "")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(db.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(db.Categories))
	}
}

func TestLoad_GlobalPlusLocalMerge(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")
	localPath := filepath.Join(dir, "local.toml")

	globalContent := `[identity]
name = "Alice"
pronouns = "she/her"

[contact]
email = "alice@example.com"
`
	localContent := `[identity]
name = "Bob"

[web]
github = "bob"
`
	if err := os.WriteFile(globalPath, []byte(globalContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localPath, []byte(localContent), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := Load(globalPath, localPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Should have 3 categories: contact (global-only), identity (merged), web (local-only)
	if len(db.Categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(db.Categories))
	}

	// identity.name should be overridden to "Bob"
	for _, cat := range db.Categories {
		if cat.Name == "identity" {
			for _, f := range cat.Fields {
				if f.Key == "name" {
					if f.Value != "Bob" {
						t.Errorf("expected identity.name = 'Bob', got %v", f.Value)
					}
				}
				if f.Key == "pronouns" {
					if f.Value != "she/her" {
						t.Errorf("expected identity.pronouns = 'she/her' (from global), got %v", f.Value)
					}
				}
			}
		}
	}
}

func TestLoad_MissingGlobal(t *testing.T) {
	_, err := Load("/nonexistent/global.toml", "")
	if err == nil {
		t.Fatal("expected error for missing global file, got nil")
	}
}

func TestLoad_EmptyLocalPath(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")

	content := `[identity]
name = "Alice"
`
	if err := os.WriteFile(globalPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	db, err := Load(globalPath, "")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(db.Categories))
	}
	if db.Categories[0].Name != "identity" {
		t.Errorf("expected category 'identity', got %q", db.Categories[0].Name)
	}
}

func TestLoad_MissingLocalFile(t *testing.T) {
	dir := t.TempDir()
	globalPath := filepath.Join(dir, "global.toml")

	content := `[identity]
name = "Alice"
`
	if err := os.WriteFile(globalPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(globalPath, "/nonexistent/local.toml")
	if err == nil {
		t.Fatal("expected error for missing local file, got nil")
	}
}
