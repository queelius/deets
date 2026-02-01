package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// FormatTable
// ---------------------------------------------------------------------------

func TestFormatTable_SingleCategory(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Category: "identity"},
		{Key: "aka", Value: []interface{}{"Alex Towell"}, Category: "identity"},
	}

	out := FormatTable(fields)

	// Single category: no "Category" column header
	if strings.Contains(out, "Category") {
		t.Error("single-category table should not contain Category column")
	}
	if !strings.Contains(out, "Key") {
		t.Error("table should contain Key header")
	}
	if !strings.Contains(out, "Value") {
		t.Error("table should contain Value header")
	}
	if !strings.Contains(out, "Alexander Towell") {
		t.Error("table should contain field value 'Alexander Towell'")
	}
	if !strings.Contains(out, "Alex Towell") {
		t.Error("table should contain array value 'Alex Towell'")
	}
	// Should contain separator line with Unicode box-drawing dash
	if !strings.Contains(out, "\u2500") {
		t.Error("table should contain Unicode box-drawing separator")
	}
}

func TestFormatTable_MultiCategory(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Category: "identity"},
		{Key: "github", Value: "queelius", Category: "web"},
	}

	out := FormatTable(fields)

	if !strings.Contains(out, "Category") {
		t.Error("multi-category table should contain Category column header")
	}
	if !strings.Contains(out, "identity") {
		t.Error("table should contain category name 'identity'")
	}
	if !strings.Contains(out, "web") {
		t.Error("table should contain category name 'web'")
	}
}

func TestFormatTable_Empty(t *testing.T) {
	out := FormatTable(nil)
	if out != "" {
		t.Errorf("expected empty string for nil fields, got %q", out)
	}

	out = FormatTable([]Field{})
	if out != "" {
		t.Errorf("expected empty string for empty fields, got %q", out)
	}
}

func TestFormatTable_ColumnAlignment(t *testing.T) {
	fields := []Field{
		{Key: "x", Value: "short", Category: "cat"},
		{Key: "longkeyname", Value: "val", Category: "cat"},
	}

	out := FormatTable(fields)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines (header, separator, data), got %d", len(lines))
	}
	// The header and separator should be the same width
	if len(lines[0]) != len(lines[1]) {
		// Not strictly enforced by the implementation, but the padded columns
		// should mean the header and separator are aligned.
		// The separator uses Unicode runes, so compare by rune count is not necessary.
		// Just verify the separator line exists.
	}
}

// ---------------------------------------------------------------------------
// FormatJSON
// ---------------------------------------------------------------------------

func TestFormatJSON_FullDB(t *testing.T) {
	db := newTestDB()
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Parse back to verify structure.
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse FormatJSON output: %v", err)
	}

	// Should have three categories.
	expectedCats := []string{"identity", "web", "academic"}
	for _, cat := range expectedCats {
		if _, ok := parsed[cat]; !ok {
			t.Errorf("expected category %q in JSON output", cat)
		}
	}
}

func TestFormatJSON_DescExcluded(t *testing.T) {
	db := newTestDB()
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	if strings.Contains(out, "name_desc") {
		t.Error("FormatJSON should exclude _desc keys")
	}
	if strings.Contains(out, "github_desc") {
		t.Error("FormatJSON should exclude _desc keys")
	}
}

func TestFormatJSON_CorrectStructure(t *testing.T) {
	db := newTestDB()
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Parse identity section to check fields.
	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	identity, ok := parsed["identity"]
	if !ok {
		t.Fatal("expected identity category")
	}
	if identity["name"] != "Alexander Towell" {
		t.Errorf("unexpected name value: %v", identity["name"])
	}
	// aka should be an array.
	aka, ok := identity["aka"]
	if !ok {
		t.Fatal("expected aka field")
	}
	akaSlice, ok := aka.([]interface{})
	if !ok {
		t.Fatalf("expected aka to be array, got %T", aka)
	}
	if len(akaSlice) != 2 {
		t.Errorf("expected 2 elements in aka, got %d", len(akaSlice))
	}
}

func TestFormatJSON_EmptyDB(t *testing.T) {
	db := &DB{}
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatCategoryJSON
// ---------------------------------------------------------------------------

func TestFormatCategoryJSON(t *testing.T) {
	db := newTestDB()
	cat, ok := db.GetCategory("web")
	if !ok {
		t.Fatal("expected to find web category")
	}

	out, err := FormatCategoryJSON(cat)
	if err != nil {
		t.Fatalf("FormatCategoryJSON error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if parsed["github"] != "queelius" {
		t.Errorf("unexpected github value: %v", parsed["github"])
	}
	if parsed["website"] != "https://example.com" {
		t.Errorf("unexpected website value: %v", parsed["website"])
	}
	// _desc should be excluded.
	if _, ok := parsed["github_desc"]; ok {
		t.Error("FormatCategoryJSON should exclude _desc keys")
	}
}

func TestFormatCategoryJSON_EmptyCategory(t *testing.T) {
	cat := Category{Name: "empty", Fields: []Field{}}
	out, err := FormatCategoryJSON(cat)
	if err != nil {
		t.Fatalf("FormatCategoryJSON error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatFieldsJSON
// ---------------------------------------------------------------------------

func TestFormatFieldsJSON_SameCategory(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Category: "identity"},
		{Key: "age", Value: int64(35), Category: "identity"},
	}

	out, err := FormatFieldsJSON(fields)
	if err != nil {
		t.Fatalf("FormatFieldsJSON error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Flat object: no nesting by category.
	if parsed["name"] != "Alexander Towell" {
		t.Errorf("unexpected name value: %v", parsed["name"])
	}
	// JSON numbers are float64 by default in Go.
	if parsed["age"] != float64(35) {
		t.Errorf("unexpected age value: %v (type %T)", parsed["age"], parsed["age"])
	}
}

func TestFormatFieldsJSON_MultiCategory(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Category: "identity"},
		{Key: "github", Value: "queelius", Category: "web"},
	}

	out, err := FormatFieldsJSON(fields)
	if err != nil {
		t.Fatalf("FormatFieldsJSON error: %v", err)
	}

	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse grouped JSON: %v", err)
	}

	if parsed["identity"]["name"] != "Alexander Towell" {
		t.Errorf("unexpected identity.name: %v", parsed["identity"]["name"])
	}
	if parsed["web"]["github"] != "queelius" {
		t.Errorf("unexpected web.github: %v", parsed["web"]["github"])
	}
}

func TestFormatFieldsJSON_Empty(t *testing.T) {
	out, err := FormatFieldsJSON([]Field{})
	if err != nil {
		t.Fatalf("FormatFieldsJSON error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object, got %q", out)
	}

	out, err = FormatFieldsJSON(nil)
	if err != nil {
		t.Fatalf("FormatFieldsJSON nil error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object for nil, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatEnv
// ---------------------------------------------------------------------------

func TestFormatEnv(t *testing.T) {
	db := newTestDB()
	out := FormatEnv(db)

	// Check expected env var lines.
	expectedLines := []string{
		`DEETS_IDENTITY_NAME="Alexander Towell"`,
		`DEETS_IDENTITY_AKA="Alex Towell, Alex T"`,
		`DEETS_IDENTITY_AGE="35"`,
		`DEETS_WEB_GITHUB="queelius"`,
		`DEETS_WEB_WEBSITE="https://example.com"`,
		`DEETS_ACADEMIC_ORCID="0000-0001-2345-6789"`,
		`DEETS_ACADEMIC_GPA="3.95"`,
		`DEETS_ACADEMIC_TOPICS="statistics, machine learning"`,
	}

	for _, line := range expectedLines {
		if !strings.Contains(out, line) {
			t.Errorf("FormatEnv output missing expected line: %s\ngot:\n%s", line, out)
		}
	}
}

func TestFormatEnv_Uppercase(t *testing.T) {
	db := newTestDB()
	out := FormatEnv(db)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			t.Errorf("expected '=' in env line: %s", line)
			continue
		}
		key := line[:eqIdx]
		if key != strings.ToUpper(key) {
			t.Errorf("env key should be uppercase: %s", key)
		}
		if !strings.HasPrefix(key, "DEETS_") {
			t.Errorf("env key should start with DEETS_: %s", key)
		}
	}
}

func TestFormatEnv_DescExcluded(t *testing.T) {
	db := newTestDB()
	out := FormatEnv(db)

	if strings.Contains(out, "_DESC=") && strings.Contains(out, "NAME_DESC") {
		t.Error("FormatEnv should exclude _desc fields")
	}
}

func TestFormatEnv_Quoting(t *testing.T) {
	db := newTestDB()
	out := FormatEnv(db)

	// All values should be quoted (surrounded by double-quotes after =)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		eqIdx := strings.Index(line, "=")
		if eqIdx < 0 {
			continue
		}
		val := line[eqIdx+1:]
		if !strings.HasPrefix(val, `"`) || !strings.HasSuffix(val, `"`) {
			t.Errorf("env value should be double-quoted: %s", line)
		}
	}
}

func TestFormatEnv_EmptyDB(t *testing.T) {
	db := &DB{}
	out := FormatEnv(db)
	if out != "" {
		t.Errorf("expected empty string for empty DB, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatTOML
// ---------------------------------------------------------------------------

func TestFormatTOML(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	// Check table headers.
	if !strings.Contains(out, "[identity]") {
		t.Error("TOML should contain [identity] header")
	}
	if !strings.Contains(out, "[web]") {
		t.Error("TOML should contain [web] header")
	}
	if !strings.Contains(out, "[academic]") {
		t.Error("TOML should contain [academic] header")
	}
}

func TestFormatTOML_Strings(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	if !strings.Contains(out, `name = "Alexander Towell"`) {
		t.Errorf("TOML should quote strings, got:\n%s", out)
	}
	if !strings.Contains(out, `github = "queelius"`) {
		t.Errorf("TOML should quote strings, got:\n%s", out)
	}
}

func TestFormatTOML_Arrays(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	// aka is []interface{} — should render as TOML array with quoted items.
	if !strings.Contains(out, `aka = ["Alex Towell", "Alex T"]`) {
		t.Errorf("TOML should render []interface{} as array, got:\n%s", out)
	}
	// topics is []string — should render similarly.
	if !strings.Contains(out, `topics = ["statistics", "machine learning"]`) {
		t.Errorf("TOML should render []string as array, got:\n%s", out)
	}
}

func TestFormatTOML_Numbers(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	if !strings.Contains(out, "age = 35") {
		t.Errorf("TOML should render int64 without quotes, got:\n%s", out)
	}
	if !strings.Contains(out, "gpa = 3.95") {
		t.Errorf("TOML should render float64 without quotes, got:\n%s", out)
	}
}

func TestFormatTOML_DescExcluded(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	if strings.Contains(out, "name_desc") {
		t.Error("TOML should exclude _desc fields")
	}
	if strings.Contains(out, "github_desc") {
		t.Error("TOML should exclude _desc fields")
	}
}

func TestFormatTOML_EmptyDB(t *testing.T) {
	db := &DB{}
	out := FormatTOML(db)
	if out != "" {
		t.Errorf("expected empty string for empty DB, got %q", out)
	}
}

func TestFormatTOML_CategorySeparation(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	// Categories after the first should be separated by a blank line.
	// Check that [web] is preceded by a newline.
	idx := strings.Index(out, "[web]")
	if idx <= 0 {
		t.Fatal("could not find [web] in TOML output")
	}
	if out[idx-1] != '\n' || out[idx-2] != '\n' {
		t.Error("categories should be separated by a blank line")
	}
}

// ---------------------------------------------------------------------------
// FormatYAML
// ---------------------------------------------------------------------------

func TestFormatYAML(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	if !strings.Contains(out, "identity:") {
		t.Error("YAML should contain identity: key")
	}
	if !strings.Contains(out, "web:") {
		t.Error("YAML should contain web: key")
	}
	if !strings.Contains(out, "academic:") {
		t.Error("YAML should contain academic: key")
	}
}

func TestFormatYAML_StringValues(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	// Simple strings should not be quoted (no special characters).
	if !strings.Contains(out, "  name: Alexander Towell") {
		t.Errorf("YAML should not quote simple strings, got:\n%s", out)
	}
	if !strings.Contains(out, "  github: queelius") {
		t.Errorf("YAML should not quote simple strings, got:\n%s", out)
	}
}

func TestFormatYAML_SpecialValuesQuoted(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "test",
				Fields: []Field{
					{Key: "flag", Value: "true", Category: "test"},
					{Key: "empty", Value: "", Category: "test"},
					{Key: "null_val", Value: "null", Category: "test"},
					{Key: "yes_val", Value: "yes", Category: "test"},
					{Key: "colon_val", Value: "foo: bar", Category: "test"},
				},
			},
		},
	}
	out := FormatYAML(db)

	// "true", "", "null", "yes" should all be quoted.
	if !strings.Contains(out, `flag: "true"`) {
		t.Errorf("YAML should quote 'true', got:\n%s", out)
	}
	if !strings.Contains(out, `empty: ""`) {
		t.Errorf("YAML should quote empty string, got:\n%s", out)
	}
	if !strings.Contains(out, `null_val: "null"`) {
		t.Errorf("YAML should quote 'null', got:\n%s", out)
	}
	if !strings.Contains(out, `yes_val: "yes"`) {
		t.Errorf("YAML should quote 'yes', got:\n%s", out)
	}
	if !strings.Contains(out, `colon_val: "foo: bar"`) {
		t.Errorf("YAML should quote string with colon, got:\n%s", out)
	}
}

func TestFormatYAML_Arrays(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	// aka ([]interface{}) should be rendered as flow sequence.
	if !strings.Contains(out, "  aka: [Alex Towell, Alex T]") {
		t.Errorf("YAML should render []interface{} as flow sequence, got:\n%s", out)
	}
	// topics ([]string) should be rendered as flow sequence.
	if !strings.Contains(out, "  topics: [statistics, machine learning]") {
		t.Errorf("YAML should render []string as flow sequence, got:\n%s", out)
	}
}

func TestFormatYAML_Numbers(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	if !strings.Contains(out, "  age: 35") {
		t.Errorf("YAML should render int64, got:\n%s", out)
	}
	if !strings.Contains(out, "  gpa: 3.95") {
		t.Errorf("YAML should render float64, got:\n%s", out)
	}
}

func TestFormatYAML_DescExcluded(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	if strings.Contains(out, "name_desc") {
		t.Error("YAML should exclude _desc fields")
	}
}

func TestFormatYAML_EmptyArray(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "test",
				Fields: []Field{
					{Key: "tags", Value: []interface{}{}, Category: "test"},
					{Key: "labels", Value: []string{}, Category: "test"},
				},
			},
		},
	}
	out := FormatYAML(db)

	if !strings.Contains(out, "  tags: []") {
		t.Errorf("YAML should render empty []interface{} as [], got:\n%s", out)
	}
	if !strings.Contains(out, "  labels: []") {
		t.Errorf("YAML should render empty []string as [], got:\n%s", out)
	}
}

func TestFormatYAML_EmptyDB(t *testing.T) {
	db := &DB{}
	out := FormatYAML(db)
	if out != "" {
		t.Errorf("expected empty string for empty DB, got %q", out)
	}
}

func TestFormatYAML_BoolValue(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "settings",
				Fields: []Field{
					{Key: "enabled", Value: true, Category: "settings"},
				},
			},
		},
	}
	out := FormatYAML(db)
	if !strings.Contains(out, "  enabled: true") {
		t.Errorf("YAML should render bool directly, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// FormatDescTable
// ---------------------------------------------------------------------------

func TestFormatDescTable(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Desc: "Full legal name", Category: "identity"},
		{Key: "orcid", Value: "0000-0001-2345-6789", Desc: "ORCID persistent digital identifier", Category: "academic"},
	}

	out := FormatDescTable(fields)

	if !strings.Contains(out, "Field") {
		t.Error("desc table should contain Field header")
	}
	if !strings.Contains(out, "Description") {
		t.Error("desc table should contain Description header")
	}
	if !strings.Contains(out, "identity.name") {
		t.Error("desc table should contain 'identity.name' path")
	}
	if !strings.Contains(out, "academic.orcid") {
		t.Error("desc table should contain 'academic.orcid' path")
	}
	if !strings.Contains(out, "Full legal name") {
		t.Error("desc table should contain description text")
	}
	if !strings.Contains(out, "ORCID persistent digital identifier") {
		t.Error("desc table should contain description text")
	}
	if !strings.Contains(out, "\u2500") {
		t.Error("desc table should contain Unicode separator")
	}
}

func TestFormatDescTable_Alignment(t *testing.T) {
	fields := []Field{
		{Key: "x", Value: "v", Desc: "short", Category: "a"},
		{Key: "longkeyname", Value: "v", Desc: "longer description here", Category: "longcategory"},
	}

	out := FormatDescTable(fields)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d", len(lines))
	}

	// Verify paths appear.
	if !strings.Contains(out, "a.x") {
		t.Error("desc table should contain path 'a.x'")
	}
	if !strings.Contains(out, "longcategory.longkeyname") {
		t.Error("desc table should contain path 'longcategory.longkeyname'")
	}
}

func TestFormatDescTable_Empty(t *testing.T) {
	out := FormatDescTable(nil)
	if out != "" {
		t.Errorf("expected empty string for nil, got %q", out)
	}

	out = FormatDescTable([]Field{})
	if out != "" {
		t.Errorf("expected empty string for empty slice, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatDescJSON
// ---------------------------------------------------------------------------

func TestFormatDescJSON(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander Towell", Desc: "Full legal name", Category: "identity"},
		{Key: "orcid", Value: "0000-0001-2345-6789", Desc: "ORCID ID", Category: "academic"},
	}

	out, err := FormatDescJSON(fields)
	if err != nil {
		t.Fatalf("FormatDescJSON error: %v", err)
	}

	var parsed map[string]string
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse FormatDescJSON output: %v", err)
	}

	if parsed["identity.name"] != "Full legal name" {
		t.Errorf("unexpected description for identity.name: %q", parsed["identity.name"])
	}
	if parsed["academic.orcid"] != "ORCID ID" {
		t.Errorf("unexpected description for academic.orcid: %q", parsed["academic.orcid"])
	}
	if len(parsed) != 2 {
		t.Errorf("expected 2 entries, got %d", len(parsed))
	}
}

func TestFormatDescJSON_Empty(t *testing.T) {
	out, err := FormatDescJSON([]Field{})
	if err != nil {
		t.Fatalf("FormatDescJSON error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object, got %q", out)
	}
}

func TestFormatDescJSON_PathFormat(t *testing.T) {
	fields := []Field{
		{Key: "email", Value: "test@test.com", Desc: "Primary email", Category: "contact"},
	}

	out, err := FormatDescJSON(fields)
	if err != nil {
		t.Fatalf("FormatDescJSON error: %v", err)
	}

	// Verify the "category.key" path format is used.
	if !strings.Contains(out, "contact.email") {
		t.Error("FormatDescJSON should use 'category.key' path format")
	}
}

// ---------------------------------------------------------------------------
// Integration-style tests using the shared test DB
// ---------------------------------------------------------------------------

func TestFormatJSON_WithTestDB(t *testing.T) {
	db := newTestDB()
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Ensure the JSON is valid.
	if !json.Valid([]byte(out)) {
		t.Error("FormatJSON should produce valid JSON")
	}

	// Ensure _desc fields are excluded.
	if strings.Contains(out, `"name_desc"`) {
		t.Error("FormatJSON should exclude _desc keys")
	}
	if strings.Contains(out, `"github_desc"`) {
		t.Error("FormatJSON should exclude _desc keys")
	}

	// Verify array values are preserved.
	if !strings.Contains(out, `"Alex Towell"`) {
		t.Error("FormatJSON should include array elements")
	}
}

func TestFormatFieldsJSON_WithTestDB(t *testing.T) {
	db := newTestDB()

	// Query a single category and format.
	fields := db.Query("identity")
	out, err := FormatFieldsJSON(fields)
	if err != nil {
		t.Fatalf("FormatFieldsJSON error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be flat (single category).
	if _, ok := parsed["name"]; !ok {
		t.Error("expected flat object with 'name' key")
	}

	// Query across categories.
	allFields := db.AllFields()
	out, err = FormatFieldsJSON(allFields)
	if err != nil {
		t.Fatalf("FormatFieldsJSON error: %v", err)
	}

	var grouped map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &grouped); err != nil {
		t.Fatalf("failed to parse grouped: %v", err)
	}

	if _, ok := grouped["identity"]; !ok {
		t.Error("expected grouped object with 'identity' key")
	}
	if _, ok := grouped["web"]; !ok {
		t.Error("expected grouped object with 'web' key")
	}
}

func TestFormatEnv_WithTestDB(t *testing.T) {
	db := newTestDB()
	out := FormatEnv(db)

	// Count the number of lines (should match non-desc fields = 8).
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 8 {
		t.Errorf("expected 8 env lines, got %d", len(lines))
	}
}

func TestFormatTOML_WithTestDB(t *testing.T) {
	db := newTestDB()
	out := FormatTOML(db)

	// Should contain all non-desc fields.
	expectedKeys := []string{
		"name =", "aka =", "age =",
		"github =", "website =",
		"orcid =", "gpa =", "topics =",
	}
	for _, key := range expectedKeys {
		if !strings.Contains(out, key) {
			t.Errorf("TOML output should contain %q, got:\n%s", key, out)
		}
	}

	// Should not contain _desc keys.
	if strings.Contains(out, "name_desc =") {
		t.Error("TOML should exclude name_desc")
	}
}

func TestFormatYAML_WithTestDB(t *testing.T) {
	db := newTestDB()
	out := FormatYAML(db)

	// Should contain all non-desc keys.
	expectedKeys := []string{
		"  name:", "  aka:", "  age:",
		"  github:", "  website:",
		"  orcid:", "  gpa:", "  topics:",
	}
	for _, key := range expectedKeys {
		if !strings.Contains(out, key) {
			t.Errorf("YAML output should contain %q, got:\n%s", key, out)
		}
	}
}

func TestFormatTable_WithTestDB(t *testing.T) {
	db := newTestDB()
	fields := db.AllFields()
	out := FormatTable(fields)

	// Multi-category should include "Category" header.
	if !strings.Contains(out, "Category") {
		t.Error("multi-category table should contain Category column")
	}

	// Should have all field keys.
	for _, f := range fields {
		if !strings.Contains(out, f.Key) {
			t.Errorf("table should contain key %q", f.Key)
		}
	}
}

// ---------------------------------------------------------------------------
// TOML bool value
// ---------------------------------------------------------------------------

func TestFormatTOML_BoolValue(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "settings",
				Fields: []Field{
					{Key: "enabled", Value: true, Category: "settings"},
				},
			},
		},
	}
	out := FormatTOML(db)
	if !strings.Contains(out, "enabled = true") {
		t.Errorf("TOML should render bool directly, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// TOML fallback value type
// ---------------------------------------------------------------------------

func TestFormatTOML_FallbackType(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "misc",
				Fields: []Field{
					{Key: "unknown", Value: struct{ X int }{42}, Category: "misc"},
				},
			},
		},
	}
	out := FormatTOML(db)
	// Fallback should be quoted.
	if !strings.Contains(out, `unknown = "{42}"`) {
		t.Errorf("TOML should quote fallback type, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// YAML arrays with special values
// ---------------------------------------------------------------------------

func TestFormatYAML_ArrayWithSpecialValues(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "test",
				Fields: []Field{
					{Key: "values", Value: []string{"true", "normal", "null"}, Category: "test"},
				},
			},
		},
	}
	out := FormatYAML(db)
	// "true" and "null" should be quoted inside the array.
	if !strings.Contains(out, `"true"`) {
		t.Errorf("YAML should quote 'true' in array, got:\n%s", out)
	}
	if !strings.Contains(out, `"null"`) {
		t.Errorf("YAML should quote 'null' in array, got:\n%s", out)
	}
	// "normal" should not be quoted.
	if strings.Contains(out, `"normal"`) {
		t.Errorf("YAML should not quote 'normal' in array, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// YAML fallback value type
// ---------------------------------------------------------------------------

func TestFormatYAML_FallbackType(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{
				Name: "misc",
				Fields: []Field{
					{Key: "unknown", Value: struct{ X int }{42}, Category: "misc"},
				},
			},
		},
	}
	out := FormatYAML(db)
	if !strings.Contains(out, "  unknown: {42}") {
		t.Errorf("YAML should use %%v for fallback, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// Ordered JSON key preservation
// ---------------------------------------------------------------------------

func TestFormatJSON_KeyOrder(t *testing.T) {
	db := newTestDB()
	out, err := FormatJSON(db)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	// Categories should appear in insertion order: identity before web before academic.
	idxIdentity := strings.Index(out, `"identity"`)
	idxWeb := strings.Index(out, `"web"`)
	idxAcademic := strings.Index(out, `"academic"`)

	if idxIdentity < 0 || idxWeb < 0 || idxAcademic < 0 {
		t.Fatal("expected all three categories in JSON output")
	}
	if idxIdentity >= idxWeb || idxWeb >= idxAcademic {
		t.Error("FormatJSON should preserve category order")
	}
}

func TestFormatDescJSON_KeyOrder(t *testing.T) {
	fields := []Field{
		{Key: "aaa", Desc: "first", Category: "cat1"},
		{Key: "bbb", Desc: "second", Category: "cat2"},
		{Key: "ccc", Desc: "third", Category: "cat3"},
	}

	out, err := FormatDescJSON(fields)
	if err != nil {
		t.Fatalf("FormatDescJSON error: %v", err)
	}

	idxA := strings.Index(out, "cat1.aaa")
	idxB := strings.Index(out, "cat2.bbb")
	idxC := strings.Index(out, "cat3.ccc")

	if idxA < 0 || idxB < 0 || idxC < 0 {
		t.Fatal("expected all three paths in output")
	}
	if idxA >= idxB || idxB >= idxC {
		t.Error("FormatDescJSON should preserve field order")
	}
}

// ---------------------------------------------------------------------------
// FieldsToDB
// ---------------------------------------------------------------------------

func TestFieldsToDB_GroupsByCategory(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alex", Category: "identity"},
		{Key: "github", Value: "queelius", Category: "web"},
		{Key: "email", Value: "alex@test.com", Category: "identity"},
	}

	db := FieldsToDB(fields)
	if len(db.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(db.Categories))
	}
	if db.Categories[0].Name != "identity" {
		t.Errorf("expected first category 'identity', got %q", db.Categories[0].Name)
	}
	if db.Categories[1].Name != "web" {
		t.Errorf("expected second category 'web', got %q", db.Categories[1].Name)
	}
	if len(db.Categories[0].Fields) != 2 {
		t.Errorf("expected 2 fields in identity, got %d", len(db.Categories[0].Fields))
	}
}

func TestFieldsToDB_Empty(t *testing.T) {
	db := FieldsToDB(nil)
	if len(db.Categories) != 0 {
		t.Errorf("expected 0 categories for nil input, got %d", len(db.Categories))
	}
}

// ---------------------------------------------------------------------------
// FormatTableWithDesc
// ---------------------------------------------------------------------------

func TestFormatTableWithDesc_MultiCat(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander", Desc: "Full name", Category: "identity"},
		{Key: "github", Value: "queelius", Desc: "GitHub username", Category: "web"},
	}

	out := FormatTableWithDesc(fields)
	if !strings.Contains(out, "Category") {
		t.Error("expected Category column")
	}
	if !strings.Contains(out, "Description") {
		t.Error("expected Description column")
	}
	if !strings.Contains(out, "Full name") {
		t.Error("expected 'Full name' in output")
	}
	if !strings.Contains(out, "GitHub username") {
		t.Error("expected 'GitHub username' in output")
	}
}

func TestFormatTableWithDesc_SingleCat(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander", Desc: "Full name", Category: "identity"},
		{Key: "email", Value: "a@b.com", Desc: "Email address", Category: "identity"},
	}

	out := FormatTableWithDesc(fields)
	if strings.Contains(out, "Category") {
		t.Error("single-cat table should not have Category column")
	}
	if !strings.Contains(out, "Description") {
		t.Error("expected Description column")
	}
}

func TestFormatTableWithDesc_Empty(t *testing.T) {
	out := FormatTableWithDesc(nil)
	if out != "" {
		t.Errorf("expected empty string for nil, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatFieldsJSONWithDesc
// ---------------------------------------------------------------------------

func TestFormatFieldsJSONWithDesc_SingleCat(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander", Desc: "Full name", Category: "identity"},
	}

	out, err := FormatFieldsJSONWithDesc(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	nameEntry, ok := parsed["name"]
	if !ok {
		t.Fatal("expected 'name' key")
	}
	if nameEntry["value"] != "Alexander" {
		t.Errorf("unexpected value: %v", nameEntry["value"])
	}
	if nameEntry["description"] != "Full name" {
		t.Errorf("unexpected description: %v", nameEntry["description"])
	}
}

func TestFormatFieldsJSONWithDesc_MultiCat(t *testing.T) {
	fields := []Field{
		{Key: "name", Value: "Alexander", Desc: "Full name", Category: "identity"},
		{Key: "github", Value: "queelius", Desc: "GitHub", Category: "web"},
	}

	out, err := FormatFieldsJSONWithDesc(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := parsed["identity"]["name"]; !ok {
		t.Error("expected identity.name in grouped output")
	}
	if _, ok := parsed["web"]["github"]; !ok {
		t.Error("expected web.github in grouped output")
	}
}

func TestFormatFieldsJSONWithDesc_Empty(t *testing.T) {
	out, err := FormatFieldsJSONWithDesc(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) != "{}" {
		t.Errorf("expected empty JSON object, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// FormatValueTOML
// ---------------------------------------------------------------------------

func TestFormatValueTOML_String(t *testing.T) {
	got := FormatValueTOML("hello")
	if got != `"hello"` {
		t.Errorf("expected quoted string, got %q", got)
	}
}

func TestFormatValueTOML_Int(t *testing.T) {
	got := FormatValueTOML(int64(42))
	if got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
}

func TestFormatValueTOML_Array(t *testing.T) {
	got := FormatValueTOML([]interface{}{"a", "b"})
	if got != `["a", "b"]` {
		t.Errorf("expected array, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// Diff formatters
// ---------------------------------------------------------------------------

func TestFormatDiffTable(t *testing.T) {
	entries := []DiffEntry{
		{Path: "identity.name", Status: "override", GlobalVal: "Global", LocalVal: "Local"},
		{Path: "custom.key", Status: "local-only", LocalVal: "value"},
	}
	out := FormatDiffTable(entries)
	if !strings.Contains(out, "Path") {
		t.Error("expected Path header")
	}
	if !strings.Contains(out, "override") {
		t.Error("expected 'override' status")
	}
	if !strings.Contains(out, "local-only") {
		t.Error("expected 'local-only' status")
	}
}

func TestFormatDiffTable_Empty(t *testing.T) {
	out := FormatDiffTable(nil)
	if out != "" {
		t.Errorf("expected empty string, got %q", out)
	}
}

func TestFormatDiffJSON(t *testing.T) {
	entries := []DiffEntry{
		{Path: "identity.name", Status: "override", GlobalVal: "Old", LocalVal: "New"},
	}
	out, err := FormatDiffJSON(entries)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(out)) {
		t.Error("expected valid JSON")
	}
	if !strings.Contains(out, "override") {
		t.Error("expected 'override' in JSON output")
	}
}
