# Agent-Native Overhaul Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restructure deets around platform-centric schema (`[profiles.*]`), add `deets populate` for interactive data harvest, and make the skill schema-driven.

**Architecture:** The `.` separator is now overloaded: it separates category from key (`identity.name`) AND appears inside category names (`profiles.github`). All path resolution changes to split on the **last** dot. Query tries full-pattern-as-category first. The parser flattens nested TOML tables. Populate shells out to `git`/`gh`/ORCID API.

**Tech Stack:** Go 1.22, cobra, BurntSushi/toml, `os/exec` for git/gh, `net/http` for ORCID API.

---

### Task 1: Add `ValidateCategoryName` and update writer validation

The writer currently rejects dotted category names (`profiles.github`) via `ValidateName`. We need a `ValidateCategoryName` that allows dots between valid bare-key segments.

**Files:**
- Modify: `internal/store/writer.go:9-21` (add ValidateCategoryName, update SetValue/RemoveValue/RemoveCategory)
- Test: `internal/store/writer_test.go`

**Step 1: Write the failing tests**

Add to `internal/store/writer_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/ -run "TestValidateCategoryName|TestSetValue_Dotted|TestRemoveValue_Dotted|TestRemoveCategory_Dotted" -v`
Expected: FAIL — `ValidateCategoryName` undefined, `SetValue` rejects dots

**Step 3: Implement `ValidateCategoryName` and update writer**

In `internal/store/writer.go`, add after `ValidateName`:

```go
// ValidateCategoryName checks that a category name is valid.
// Category names may contain dots (for nested TOML tables like "profiles.github"),
// but each dot-separated segment must be a valid bare key.
func ValidateCategoryName(name string) error {
	if name == "" {
		return fmt.Errorf("name must not be empty")
	}
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if err := ValidateName(part); err != nil {
			return err
		}
	}
	return nil
}
```

Update `SetValue`, `RemoveValue`, `RemoveCategory` to use `ValidateCategoryName(category)` instead of `ValidateName(category)`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/ -run "TestValidateCategoryName|TestSetValue_Dotted|TestRemoveValue_Dotted|TestRemoveCategory_Dotted" -v`
Expected: PASS

Run: `go test ./internal/store/ -v` (ensure no regressions)
Expected: PASS

**Step 5: Commit**

```bash
git add internal/store/writer.go internal/store/writer_test.go
git commit -m "feat: add ValidateCategoryName for dotted category support"
```

---

### Task 2: Flatten nested TOML tables in `LoadFile`

When BurntSushi/toml parses `[profiles.github]`, it produces `{"profiles": {"github": {...}}}`. LoadFile currently sees `profiles` as a category with sub-map values and silently skips them. Fix: recursively flatten nested maps into dotted category names.

**Files:**
- Modify: `internal/store/store.go:19-89` (LoadFile)
- Test: `internal/store/store_test.go`

**Step 1: Write the failing test**

Add to `internal/store/store_test.go`:

```go
func TestLoadFile_NestedTables(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[identity]
name = "Alice"

[profiles.github]
username = "alice"
email = "alice@gh.com"

[profiles.pypi]
username = "alice"
`
	os.WriteFile(path, []byte(content), 0644)

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	// Expect 3 categories: identity, profiles.github, profiles.pypi
	if len(db.Categories) != 3 {
		t.Fatalf("expected 3 categories, got %d: %v", len(db.Categories), catNames(db))
	}

	// Check sorted order
	expected := []string{"identity", "profiles.github", "profiles.pypi"}
	for i, cat := range db.Categories {
		if cat.Name != expected[i] {
			t.Errorf("category[%d]: expected %q, got %q", i, expected[i], cat.Name)
		}
	}

	// Check profiles.github fields
	ghCat := db.Categories[1]
	if len(ghCat.Fields) != 2 {
		t.Fatalf("expected 2 fields in profiles.github, got %d", len(ghCat.Fields))
	}
	if ghCat.Fields[1].Key != "username" || ghCat.Fields[1].Value != "alice" {
		t.Errorf("unexpected field: %+v", ghCat.Fields[1])
	}
	if ghCat.Fields[1].Category != "profiles.github" {
		t.Errorf("expected Category='profiles.github', got %q", ghCat.Fields[1].Category)
	}
}

func TestLoadFile_NestedTablesWithDescs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[profiles.github]
username = "alice"
username_desc = "GitHub handle"
`
	os.WriteFile(path, []byte(content), 0644)

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	if len(db.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(db.Categories))
	}
	f := db.Categories[0].Fields[0]
	if f.Key != "username" {
		t.Errorf("expected key 'username', got %q", f.Key)
	}
	if f.Desc != "GitHub handle" {
		t.Errorf("expected desc 'GitHub handle', got %q", f.Desc)
	}
}

func TestLoadFile_MixedFlatAndNested(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	// identity is flat, profiles has nested sub-tables
	content := `[identity]
name = "Alice"

[contact]
email = "alice@example.com"

[profiles.github]
username = "alice"

[profiles.pypi]
username = "alice"
`
	os.WriteFile(path, []byte(content), 0644)

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	expected := []string{"contact", "identity", "profiles.github", "profiles.pypi"}
	if len(db.Categories) != len(expected) {
		t.Fatalf("expected %d categories, got %d: %v", len(expected), len(db.Categories), catNames(db))
	}
	for i, cat := range db.Categories {
		if cat.Name != expected[i] {
			t.Errorf("category[%d]: expected %q, got %q", i, expected[i], cat.Name)
		}
	}
}

// helper to extract category names for error messages
func catNames(db *model.DB) []string {
	names := make([]string, len(db.Categories))
	for i, c := range db.Categories {
		names[i] = c.Name
	}
	return names
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/ -run "TestLoadFile_Nested|TestLoadFile_Mixed" -v`
Expected: FAIL — wrong category count (nested tables produce no categories)

**Step 3: Implement recursive flattening in `LoadFile`**

Replace the inner loop of `LoadFile` in `internal/store/store.go`. Extract a helper function `parseCategory` that processes a single category map, and call it recursively when it encounters nested maps:

```go
func LoadFile(path string) (*model.DB, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	var raw map[string]interface{}
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}

	db := &model.DB{}
	collectCategories(db, "", raw)

	// Sort categories alphabetically.
	sort.Slice(db.Categories, func(i, j int) bool {
		return db.Categories[i].Name < db.Categories[j].Name
	})

	return db, nil
}

// collectCategories recursively walks the parsed TOML map. Flat tables
// (keys → scalar/array values) become categories; nested tables
// (keys → map values) are flattened with dot-joined names.
func collectCategories(db *model.DB, prefix string, raw map[string]interface{}) {
	// Separate sub-tables from leaf fields.
	subTables := make(map[string]map[string]interface{})
	leafFields := make(map[string]interface{})

	for k, v := range raw {
		if m, ok := v.(map[string]interface{}); ok {
			// Could be a sub-table OR a flat category — check if it has nested maps.
			hasNested := false
			for _, sv := range m {
				if _, isMap := sv.(map[string]interface{}); isMap {
					hasNested = true
					break
				}
			}
			if hasNested {
				// Contains sub-tables: recurse.
				subTables[k] = m
			} else {
				// All leaf values: treat as a category.
				leafFields[k] = v
			}
		}
		// Non-map top-level values are ignored (same as before).
	}

	// Process leaf-field categories (flat tables).
	leafNames := make([]string, 0, len(leafFields))
	for name := range leafFields {
		leafNames = append(leafNames, name)
	}
	sort.Strings(leafNames)

	for _, name := range leafNames {
		catMap := leafFields[name].(map[string]interface{})
		catName := name
		if prefix != "" {
			catName = prefix + "." + name
		}
		cat := buildCategory(catName, catMap)
		if len(cat.Fields) > 0 {
			db.Categories = append(db.Categories, cat)
		}
	}

	// Recurse into sub-tables.
	subNames := make([]string, 0, len(subTables))
	for name := range subTables {
		subNames = append(subNames, name)
	}
	sort.Strings(subNames)

	for _, name := range subNames {
		newPrefix := name
		if prefix != "" {
			newPrefix = prefix + "." + name
		}
		collectCategories(db, newPrefix, subTables[name])
	}
}

// buildCategory creates a model.Category from a flat TOML table map.
func buildCategory(catName string, catMap map[string]interface{}) model.Category {
	var keys []string
	for k := range catMap {
		if !strings.HasSuffix(k, "_desc") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	cat := model.Category{Name: catName}
	for _, key := range keys {
		f := model.Field{
			Key:      key,
			Value:    catMap[key],
			Category: catName,
		}

		if desc, ok := catMap[key+"_desc"]; ok {
			if s, ok := desc.(string); ok {
				f.Desc = s
			}
		}

		if f.Desc == "" {
			if catDescs, ok := DefaultDescriptions[catName]; ok {
				if d, ok := catDescs[key]; ok {
					f.Desc = d
				}
			}
		}

		cat.Fields = append(cat.Fields, f)
	}
	return cat
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/ -v`
Expected: ALL PASS (including existing tests — flat tables still work)

**Step 5: Commit**

```bash
git add internal/store/store.go internal/store/store_test.go
git commit -m "feat: flatten nested TOML tables into dotted category names"
```

---

### Task 3: Update path resolution — split on last dot

Currently `parsePath`, `GetField`, and `Query` split on the **first** dot. With dotted categories (`profiles.github.username`), they must split on the **last** dot.

**Files:**
- Modify: `internal/commands/helpers.go:31-43` (parsePath)
- Modify: `internal/model/model.go:41-59` (GetField), `71-128` (Query)
- Test: `internal/commands/helpers_test.go`
- Test: `internal/model/model_test.go`

**Step 1: Write the failing tests**

Add to `internal/commands/helpers_test.go`:

```go
func TestParsePath_DottedCategory(t *testing.T) {
	tests := []struct {
		path    string
		wantCat string
		wantKey string
		wantErr bool
	}{
		{"identity.name", "identity", "name", false},
		{"profiles.github.username", "profiles.github", "username", false},
		{"a.b.c.d", "a.b.c", "d", false},
		{"nokey", "", "", true},
		{"", "", "", true},
	}
	for _, tt := range tests {
		cat, key, err := parsePath(tt.path)
		if (err != nil) != tt.wantErr {
			t.Errorf("parsePath(%q) error=%v, wantErr=%v", tt.path, err, tt.wantErr)
			continue
		}
		if cat != tt.wantCat || key != tt.wantKey {
			t.Errorf("parsePath(%q) = (%q, %q), want (%q, %q)", tt.path, cat, key, tt.wantCat, tt.wantKey)
		}
	}
}
```

Add to `internal/model/model_test.go` (or create if needed):

```go
func TestGetField_DottedCategory(t *testing.T) {
	db := &model.DB{
		Categories: []model.Category{
			{Name: "identity", Fields: []model.Field{
				{Key: "name", Value: "Alice", Category: "identity"},
			}},
			{Name: "profiles.github", Fields: []model.Field{
				{Key: "username", Value: "alice", Category: "profiles.github"},
				{Key: "email", Value: "a@gh.com", Category: "profiles.github"},
			}},
		},
	}

	// Flat category
	f, ok := db.GetField("identity.name")
	if !ok || f.Value != "Alice" {
		t.Errorf("GetField(identity.name) = %v, %v", f.Value, ok)
	}

	// Dotted category
	f, ok = db.GetField("profiles.github.username")
	if !ok || f.Value != "alice" {
		t.Errorf("GetField(profiles.github.username) = %v, %v", f.Value, ok)
	}

	// Not found
	_, ok = db.GetField("profiles.github.nonexistent")
	if ok {
		t.Error("expected not found for nonexistent key")
	}
}

func TestQuery_DottedCategoryAsCategory(t *testing.T) {
	db := &model.DB{
		Categories: []model.Category{
			{Name: "profiles.github", Fields: []model.Field{
				{Key: "username", Value: "alice", Category: "profiles.github"},
				{Key: "email", Value: "a@gh.com", Category: "profiles.github"},
			}},
			{Name: "profiles.pypi", Fields: []model.Field{
				{Key: "username", Value: "alice", Category: "profiles.pypi"},
			}},
		},
	}

	// "profiles.github" should match as a category name, returning all fields
	fields := db.Query("profiles.github")
	if len(fields) != 2 {
		t.Fatalf("Query(profiles.github) returned %d fields, want 2", len(fields))
	}

	// "profiles.github.username" should return the specific field
	fields = db.Query("profiles.github.username")
	if len(fields) != 1 || fields[0].Value != "alice" {
		t.Errorf("Query(profiles.github.username) = %v, want [alice]", fields)
	}

	// "profiles.*.email" should glob across dotted categories
	fields = db.Query("profiles.*.email")
	if len(fields) != 1 || fields[0].Value != "a@gh.com" {
		t.Errorf("Query(profiles.*.email) = %v, want [a@gh.com]", fields)
	}

	// "profiles.*" should match all profiles categories
	fields = db.Query("profiles.*")
	if len(fields) != 3 {
		t.Errorf("Query(profiles.*) returned %d fields, want 3", len(fields))
	}

	// "profiles.*.username" should find both
	fields = db.Query("profiles.*.username")
	if len(fields) != 2 {
		t.Errorf("Query(profiles.*.username) returned %d fields, want 2", len(fields))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/commands/ -run TestParsePath_Dotted -v`
Run: `go test ./internal/model/ -run "TestGetField_Dotted|TestQuery_Dotted" -v`
Expected: FAIL

**Step 3: Implement the changes**

**`internal/commands/helpers.go` — parsePath (split on last dot):**

```go
func parsePath(path string) (category, key string, err error) {
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 || lastDot == 0 || lastDot == len(path)-1 {
		return "", "", fmt.Errorf("invalid path %q: expected category.key", path)
	}
	category = path[:lastDot]
	key = path[lastDot+1:]
	if err := store.ValidateCategoryName(category); err != nil {
		return "", "", fmt.Errorf("invalid path %q: %w", path, err)
	}
	if err := store.ValidateName(key); err != nil {
		return "", "", fmt.Errorf("invalid path %q: %w", path, err)
	}
	return category, key, nil
}
```

**`internal/model/model.go` — GetField (split on last dot):**

```go
func (db *DB) GetField(path string) (Field, bool) {
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 {
		return Field{}, false
	}
	catName, key := path[:lastDot], path[lastDot+1:]

	for _, cat := range db.Categories {
		if cat.Name == catName {
			for _, f := range cat.Fields {
				if f.Key == key {
					return f, true
				}
			}
			return Field{}, false
		}
	}
	return Field{}, false
}
```

**`internal/model/model.go` — Query (category-first, then last-dot split):**

```go
func (db *DB) Query(pattern string) []Field {
	var results []Field

	// First: try the full pattern as a category name or glob.
	// This handles "profiles.github" → all fields, and "profiles.*" → all profiles fields.
	catMatches := db.matchCategories(pattern)
	if len(catMatches) > 0 {
		for _, cat := range catMatches {
			for _, f := range cat.Fields {
				if !IsDescKey(f.Key) {
					results = append(results, f)
				}
			}
		}
		return results
	}

	// No category match — split on last dot for field lookup.
	lastDot := strings.LastIndex(pattern, ".")
	if lastDot == -1 {
		return results
	}
	catPattern, keyPattern := pattern[:lastDot], pattern[lastDot+1:]

	for _, cat := range db.Categories {
		catMatched, err := filepath.Match(catPattern, cat.Name)
		if err != nil {
			catMatched = catPattern == cat.Name
		}
		if !catMatched {
			continue
		}

		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			keyMatched, err := filepath.Match(keyPattern, f.Key)
			if err != nil {
				keyMatched = keyPattern == f.Key
			}
			if keyMatched {
				results = append(results, f)
			}
		}
	}

	return results
}

// matchCategories returns all categories whose name matches the pattern
// (exact or glob). Returns nil if no category matches.
func (db *DB) matchCategories(pattern string) []Category {
	var matches []Category
	for _, cat := range db.Categories {
		if cat.Name == pattern {
			matches = append(matches, cat)
			continue
		}
		matched, err := filepath.Match(pattern, cat.Name)
		if err == nil && matched {
			matches = append(matches, cat)
		}
	}
	return matches
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands/ -v`
Run: `go test ./internal/model/ -v`
Run: `go test ./... -v` (full regression)
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/commands/helpers.go internal/commands/helpers_test.go internal/model/model.go internal/model/model_test.go
git commit -m "feat: split paths on last dot to support dotted category names"
```

---

### Task 4: Update `rm` to handle dotted categories

`deets rm profiles.github` should remove the entire `[profiles.github]` category, not try to remove field `github` from category `profiles`.

**Files:**
- Modify: `internal/commands/rm.go`
- Modify: `internal/store/writer.go` (add `HasCategory`)
- Test: `internal/commands/rm_test.go`

**Step 1: Write the failing test**

Add to `internal/commands/rm_test.go`:

```go
func TestRm_DottedCategory(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	toml := "[identity]\nname = \"Alice\"\n\n[profiles.github]\nusername = \"alice\"\n"
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644)

	_, _, err := executeCommand("rm", "profiles.github")
	if err != nil {
		t.Fatalf("rm profiles.github failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if strings.Contains(string(data), "profiles.github") {
		t.Errorf("expected profiles.github removed, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "[identity]") {
		t.Errorf("expected identity preserved, got:\n%s", string(data))
	}
}

func TestRm_DottedCategoryField(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	toml := "[profiles.github]\nusername = \"alice\"\nemail = \"a@gh.com\"\n"
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644)

	_, _, err := executeCommand("rm", "profiles.github.email")
	if err != nil {
		t.Fatalf("rm profiles.github.email failed: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(deetsDir, "me.toml"))
	if strings.Contains(string(data), "email") {
		t.Errorf("expected email removed, got:\n%s", string(data))
	}
	if !strings.Contains(string(data), "username") {
		t.Errorf("expected username preserved, got:\n%s", string(data))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/commands/ -run TestRm_Dotted -v`
Expected: FAIL

**Step 3: Implement**

Add `HasCategory` to `internal/store/writer.go`:

```go
// HasCategory reports whether the TOML file at filePath contains a [category] section.
func HasCategory(filePath, category string) bool {
	lines, err := readLines(filePath)
	if err != nil {
		return false
	}
	return findSection(lines, category) != -1
}
```

Update `internal/commands/rm.go`:

```go
var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Remove a field or category",
	Long: `Remove a field or entire category.

Examples:
  deets rm contact.phone         # remove a field
  deets rm cooking               # remove entire category
  deets rm profiles.github       # remove a dotted category
  deets rm profiles.github.email # remove a field in dotted category`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		filePath, err := targetFile()
		if err != nil {
			return err
		}

		// Check if the full path is a category name (handles both
		// "cooking" and "profiles.github").
		if store.HasCategory(filePath, path) {
			return store.RemoveCategory(filePath, path)
		}

		// Not a category — must be a field path.
		if !strings.Contains(path, ".") {
			return fmt.Errorf("category %q not found", path)
		}

		cat, key, err := parsePath(path)
		if err != nil {
			return err
		}
		return store.RemoveValue(filePath, cat, key)
	},
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands/ -run TestRm -v`
Run: `go test ./... -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/store/writer.go internal/commands/rm.go internal/commands/rm_test.go
git commit -m "feat: rm handles dotted category names"
```

---

### Task 5: Update template and DefaultDescriptions

New init template with `[profiles.*]` sections. Expanded DefaultDescriptions covering all profile sub-keys.

**Files:**
- Modify: `internal/store/template.go`
- Test: `internal/store/store_test.go` (DefaultDescriptions for profiles)

**Step 1: Write the failing test**

Add to `internal/store/store_test.go`:

```go
func TestLoadFile_ProfilesDefaultDescriptions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "me.toml")

	content := `[profiles.github]
username = "alice"
email = "a@gh.com"
url = "https://github.com/alice"
`
	os.WriteFile(path, []byte(content), 0644)

	db, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile: %v", err)
	}

	cat := db.Categories[0]
	for _, f := range cat.Fields {
		if f.Desc == "" {
			t.Errorf("field %q has empty description (should have default)", f.Key)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/store/ -run TestLoadFile_ProfilesDefault -v`
Expected: FAIL — no DefaultDescriptions for "profiles.github"

**Step 3: Implement**

Rewrite `internal/store/template.go`:

```go
package store

const DefaultTemplate = `# deets — Personal metadata
# Any [category] with any key = "value" is valid.
# Run 'deets populate --all' to auto-fill from your accounts.

[identity]
# name = "Your Name"
# aka = ["Nickname"]
# pronouns = "they/them"

[contact]
# email = "you@example.com"

[academic]
# orcid = "0000-0000-0000-0000"
# institution = "University of..."
# research_interests = ["topic1", "topic2"]
# scholar = "Google Scholar ID"

[education]
# degrees = ["BS Computer Science (University, 2020)"]
# phd = "Field of study"

[profiles.github]
# username = "your-username"

[profiles.pypi]
# username = "your-username"

[profiles.orcid]
# id = "0000-0000-0000-0000"
`

const LocalTemplate = `# deets — Local project overrides
# Keys here override matching keys from ~/.deets/me.toml
# Only include fields you want to override for this project.
`

var DefaultDescriptions = map[string]map[string]string{
	"identity": {
		"name":     "Full legal name",
		"aka":      "Known aliases and nicknames",
		"pronouns": "Personal pronouns",
	},
	"contact": {
		"email": "Primary email address",
		"phone": "Phone number",
	},
	"academic": {
		"orcid":              "ORCID persistent digital identifier",
		"institution":        "Academic institution",
		"title":              "Academic title or position",
		"research_interests": "Research interest areas",
		"scholar":            "Google Scholar profile ID",
	},
	"education": {
		"degrees":     "Completed degrees with institution and year",
		"field":       "Primary field of study",
		"institution": "Degree-granting institution",
		"phd":              "PhD field of study",
		"phd_institution":  "PhD institution",
	},
	"profiles.github": {
		"username": "GitHub username",
		"name":     "Display name on GitHub",
		"email":    "Email associated with GitHub account",
		"url":      "GitHub profile URL",
		"bio":      "GitHub profile bio",
	},
	"profiles.pypi": {
		"username": "PyPI username",
		"name":     "Author name for PyPI packages",
		"email":    "Email for PyPI packages",
		"url":      "PyPI profile URL",
	},
	"profiles.cran": {
		"name":  "Author name for CRAN packages",
		"email": "Email for CRAN packages",
	},
	"profiles.orcid": {
		"id":    "ORCID identifier",
		"name":  "Name registered with ORCID",
		"email": "Email registered with ORCID",
		"url":   "ORCID profile URL",
	},
	"profiles.bluesky": {
		"handle": "Bluesky handle",
		"url":    "Bluesky profile URL",
	},
	"profiles.mastodon": {
		"handle": "Mastodon handle",
		"url":    "Mastodon profile URL",
	},
	"profiles.blog": {
		"url": "Blog URL",
		"alt": "Alternative blog URL",
	},
	"profiles.zenodo": {
		"username": "Zenodo username",
		"url":      "Zenodo profile URL",
	},
	"profiles.linkedin": {
		"url": "LinkedIn profile URL",
	},
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/ -v`
Run: `go test ./... -v`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/store/template.go internal/store/store_test.go
git commit -m "feat: platform-centric template and expanded DefaultDescriptions"
```

---

### Task 6: Rewrite the Claude Code skill

Replace the 107-line exhaustive skill with a ~40-line schema-driven skill.

**Files:**
- Modify: `internal/commands/skill.md`

**Step 1: Rewrite the skill file**

See the design doc section 4 for full content. The key difference: agents run `deets schema --format json` to discover fields instead of the skill enumerating them.

**Step 2: Run tests to verify skill still embeds correctly**

Run: `go build -o deets ./cmd/deets && ./deets claude install && cat ~/.claude/skills/deets/SKILL.md`
Expected: new skill content appears

**Step 3: Commit**

```bash
git add internal/commands/skill.md
git commit -m "feat: schema-driven skill replaces exhaustive command listing"
```

---

### Task 7: Add `deets populate --git`

The simplest harvest source: reads `git config user.name` and `git config user.email`.

**Files:**
- Create: `internal/commands/populate.go`
- Test: `internal/commands/populate_test.go`

**Step 1: Write the failing test**

```go
func TestPopulate_GitDryRun(t *testing.T) {
	home := setupTestEnv(t)
	deetsDir := filepath.Join(home, ".deets")
	os.MkdirAll(deetsDir, 0755)
	os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte("[identity]\n"), 0644)

	// Configure git in the test env
	exec.Command("git", "config", "--global", "user.name", "Test User").Run()
	exec.Command("git", "config", "--global", "user.email", "test@example.com").Run()

	flagPopulateDryRun = true
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
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/commands/ -run TestPopulate -v`
Expected: FAIL — `populate` command not defined

**Step 3: Implement `populate.go`**

Create `internal/commands/populate.go` with:
- Flags: `--git`, `--github`, `--orcid`, `--all`, `--dry-run`, `--yes`
- Each source returns `[]populateEntry{category, key, value}`
- Diff against existing DB, display additions/changes
- If not dry-run and not `--yes`, prompt for confirmation via stdin
- Write via `store.SetValue`

The `--git` source:
```go
func populateGit() ([]populateEntry, error) {
	var entries []populateEntry
	name, err := exec.Command("git", "config", "user.name").Output()
	if err == nil && len(strings.TrimSpace(string(name))) > 0 {
		entries = append(entries, populateEntry{"identity", "name", strings.TrimSpace(string(name))})
	}
	email, err := exec.Command("git", "config", "user.email").Output()
	if err == nil && len(strings.TrimSpace(string(email))) > 0 {
		entries = append(entries, populateEntry{"contact", "email", strings.TrimSpace(string(email))})
	}
	return entries, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/commands/ -run TestPopulate -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/commands/populate.go internal/commands/populate_test.go
git commit -m "feat: add 'deets populate --git' for git config harvest"
```

---

### Task 8: Add `deets populate --github`

Shells out to `gh api user` to fetch GitHub profile data.

**Files:**
- Modify: `internal/commands/populate.go` (add populateGitHub function)
- Test: `internal/commands/populate_test.go`

**Step 1: Write the test**

Testing with a mock is hard since we shell out to `gh`. Write a unit test for the mapping logic (JSON → entries) and a skip-if-no-gh integration test:

```go
func TestPopulateGitHub_Mapping(t *testing.T) {
	ghJSON := `{"login":"alice","name":"Alice Smith","email":"alice@gh.com","bio":"Developer","blog":"https://alice.dev","html_url":"https://github.com/alice"}`
	entries, err := parseGitHubUser([]byte(ghJSON))
	if err != nil {
		t.Fatalf("parseGitHubUser: %v", err)
	}
	// Should produce profiles.github.username, .name, .email, .url at minimum
	found := make(map[string]string)
	for _, e := range entries {
		found[e.category+"."+e.key] = e.value
	}
	if found["profiles.github.username"] != "alice" {
		t.Errorf("expected username=alice, got %q", found["profiles.github.username"])
	}
	if found["profiles.github.url"] != "https://github.com/alice" {
		t.Errorf("expected url, got %q", found["profiles.github.url"])
	}
}
```

**Step 2-5: Implement, test, commit**

Run: `go test ./internal/commands/ -run TestPopulateGitHub -v`

```bash
git add internal/commands/populate.go internal/commands/populate_test.go
git commit -m "feat: add 'deets populate --github' for GitHub profile harvest"
```

---

### Task 9: Add `deets populate --orcid`

Uses the public ORCID API to fetch profile data. Requires `academic.orcid` to be set first.

**Files:**
- Modify: `internal/commands/populate.go` (add populateOrcid function)
- Test: `internal/commands/populate_test.go`

**Step 1: Write the test**

```go
func TestPopulateOrcid_Mapping(t *testing.T) {
	// Minimal ORCID API response structure
	orcidJSON := `{"name":{"given-names":{"value":"Alice"},"family-name":{"value":"Smith"}}}`
	entries, err := parseOrcidPerson([]byte(orcidJSON), "0000-0001-2345-6789")
	if err != nil {
		t.Fatalf("parseOrcidPerson: %v", err)
	}
	found := make(map[string]string)
	for _, e := range entries {
		found[e.category+"."+e.key] = e.value
	}
	if found["profiles.orcid.name"] != "Alice Smith" {
		t.Errorf("expected name='Alice Smith', got %q", found["profiles.orcid.name"])
	}
}
```

**Step 2-5: Implement, test, commit**

```bash
git add internal/commands/populate.go internal/commands/populate_test.go
git commit -m "feat: add 'deets populate --orcid' for ORCID profile harvest"
```

---

### Task 10: Update user's `me.toml` to new schema

Rewrite `~/.deets/me.toml` to use the new `[profiles.*]` structure, replacing `[web]`, `[packages]`, and `[publications]`.

**This is a manual task.** After all code changes are working, run:

```bash
./deets edit    # manually restructure to new schema
# OR
./deets populate --git --github --yes   # auto-populate profiles
```

**Step 1: Build and verify**

```bash
go build -o deets ./cmd/deets
./deets show --format json    # verify all categories parse correctly
./deets schema --format json  # verify all fields have descriptions
./deets get profiles.github   # verify dotted category queries work
```

**Step 2: Reinstall skill**

```bash
./deets claude uninstall && ./deets claude install
```

**Step 3: Commit the updated me.toml** (user's choice, not tracked in repo)

---

### Task 11: Final verification and version bump

**Files:**
- Modify: `CITATION.cff` (version)
- Modify: `.zenodo.json` (version)
- Modify: `README.md` (update examples for profiles schema)

**Step 1: Full test suite**

```bash
go vet ./...
go test ./... -v
go test -cover ./internal/store/
go test -cover ./internal/model/
go test -cover ./internal/commands/
```

**Step 2: Manual smoke tests**

```bash
./deets get profiles.github
./deets get profiles.*.email
./deets get profiles.*.url
./deets set profiles.test.foo "bar"
./deets rm profiles.test
./deets schema --format json | head -30
./deets populate --git --dry-run
```

**Step 3: Update version and docs, commit**

```bash
# Update version in CITATION.cff, .zenodo.json
# Update README examples to show profiles.* queries
git add -A
git commit -m "v0.8: agent-native overhaul with platform profiles and populate"
```
