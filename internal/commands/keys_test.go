package commands

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestKeys_Table(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("keys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one key path")
	}
	// Check some expected paths
	found := make(map[string]bool)
	for _, line := range lines {
		found[strings.TrimSpace(line)] = true
	}
	for _, expected := range []string{"identity.name", "identity.aka", "web.github", "academic.orcid"} {
		if !found[expected] {
			t.Errorf("expected key path %q in output", expected)
		}
	}
	// Ensure no _desc keys
	for path := range found {
		if strings.HasSuffix(path, "_desc") {
			t.Errorf("_desc key should be excluded: %s", path)
		}
	}
}

func TestKeys_JSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("keys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var paths []string
	if err := json.Unmarshal([]byte(stdout), &paths); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(paths) == 0 {
		t.Error("expected at least one path in JSON array")
	}
	found := make(map[string]bool)
	for _, p := range paths {
		found[p] = true
	}
	if !found["identity.name"] {
		t.Error("expected identity.name in JSON array")
	}
}

func TestKeys_EmptyDB(t *testing.T) {
	home := setupTestEnv(t)
	// Create an empty deets file
	deetsDir := home + "/.deets"
	if err := writeTestFile(deetsDir+"/me.toml", ""); err != nil {
		t.Fatalf("writing empty TOML: %v", err)
	}
	flagFormat = "json"
	stdout, _, err := executeCommand("keys")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var paths []string
	if err := json.Unmarshal([]byte(stdout), &paths); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(paths) != 0 {
		t.Errorf("expected empty array for empty DB, got %d items", len(paths))
	}
}

// writeTestFile is a helper to create a file with given content, creating
// parent directories as needed.
func writeTestFile(path, content string) error {
	dir := path[:strings.LastIndex(path, "/")]
	if err := mkdirAll(dir); err != nil {
		return err
	}
	return writeFile(path, content)
}

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
