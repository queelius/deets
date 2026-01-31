package model

import (
	"testing"
)

// newTestDB builds a representative test database with multiple categories,
// various value types, _desc companion fields, and enough variety to exercise
// every code path in the model package.
func newTestDB() *DB {
	return &DB{
		Categories: []Category{
			{
				Name: "identity",
				Fields: []Field{
					{Key: "name", Value: "Alexander Towell", Desc: "Full legal name", Category: "identity"},
					{Key: "name_desc", Value: "Description for name field", Desc: "", Category: "identity"},
					{Key: "aka", Value: []interface{}{"Alex Towell", "Alex T"}, Desc: "Known aliases", Category: "identity"},
					{Key: "age", Value: int64(35), Desc: "", Category: "identity"},
				},
			},
			{
				Name: "web",
				Fields: []Field{
					{Key: "github", Value: "queelius", Desc: "GitHub username", Category: "web"},
					{Key: "github_desc", Value: "companion desc", Desc: "", Category: "web"},
					{Key: "website", Value: "https://example.com", Desc: "Personal website", Category: "web"},
				},
			},
			{
				Name: "academic",
				Fields: []Field{
					{Key: "orcid", Value: "0000-0001-2345-6789", Desc: "ORCID persistent digital identifier", Category: "academic"},
					{Key: "gpa", Value: float64(3.95), Desc: "", Category: "academic"},
					{Key: "topics", Value: []string{"statistics", "machine learning"}, Desc: "Research topics", Category: "academic"},
				},
			},
		},
	}
}

// ---------------------------------------------------------------------------
// GetField
// ---------------------------------------------------------------------------

func TestGetField_Found(t *testing.T) {
	db := newTestDB()
	f, ok := db.GetField("identity.name")
	if !ok {
		t.Fatal("expected to find identity.name")
	}
	if f.Key != "name" || f.Category != "identity" {
		t.Errorf("unexpected field: %+v", f)
	}
	if f.Value != "Alexander Towell" {
		t.Errorf("unexpected value: %v", f.Value)
	}
}

func TestGetField_NotFound(t *testing.T) {
	db := newTestDB()
	_, ok := db.GetField("identity.nonexistent")
	if ok {
		t.Error("expected not found for identity.nonexistent")
	}
}

func TestGetField_InvalidPath_NoDot(t *testing.T) {
	db := newTestDB()
	_, ok := db.GetField("nodotpath")
	if ok {
		t.Error("expected not found for path without dot")
	}
}

func TestGetField_WrongCategory(t *testing.T) {
	db := newTestDB()
	_, ok := db.GetField("nosuchcategory.name")
	if ok {
		t.Error("expected not found for wrong category")
	}
}

func TestGetField_WrongKey(t *testing.T) {
	db := newTestDB()
	_, ok := db.GetField("identity.orcid")
	if ok {
		t.Error("expected not found for key in wrong category")
	}
}

func TestGetField_DescKey(t *testing.T) {
	db := newTestDB()
	f, ok := db.GetField("identity.name_desc")
	if !ok {
		t.Fatal("GetField should find _desc keys when queried directly")
	}
	if f.Key != "name_desc" {
		t.Errorf("unexpected key: %s", f.Key)
	}
}

// ---------------------------------------------------------------------------
// Query
// ---------------------------------------------------------------------------

func TestQuery_ExactMatch(t *testing.T) {
	db := newTestDB()
	results := db.Query("identity.name")
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Key != "name" || results[0].Category != "identity" {
		t.Errorf("unexpected result: %+v", results[0])
	}
}

func TestQuery_CategoryShorthand(t *testing.T) {
	db := newTestDB()
	results := db.Query("identity")
	// identity has 4 fields, 1 is _desc => 3 non-desc fields
	if len(results) != 3 {
		t.Fatalf("expected 3 results for identity shorthand, got %d", len(results))
	}
	for _, f := range results {
		if f.Category != "identity" {
			t.Errorf("expected category identity, got %s", f.Category)
		}
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded", f.Key)
		}
	}
}

func TestQuery_CategoryDotStar(t *testing.T) {
	db := newTestDB()
	results := db.Query("web.*")
	// web has 3 fields, 1 is _desc => 2 non-desc fields
	if len(results) != 2 {
		t.Fatalf("expected 2 results for web.*, got %d", len(results))
	}
	for _, f := range results {
		if f.Category != "web" {
			t.Errorf("expected category web, got %s", f.Category)
		}
	}
}

func TestQuery_StarDotKey(t *testing.T) {
	db := newTestDB()
	// "github" only exists in web
	results := db.Query("*.github")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for *.github, got %d", len(results))
	}
	if results[0].Key != "github" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestQuery_StarDotKey_MultipleCategories(t *testing.T) {
	// Build a DB where the same key exists in two categories.
	db := &DB{
		Categories: []Category{
			{Name: "a", Fields: []Field{{Key: "email", Value: "a@a.com", Category: "a"}}},
			{Name: "b", Fields: []Field{{Key: "email", Value: "b@b.com", Category: "b"}}},
		},
	}
	results := db.Query("*.email")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for *.email, got %d", len(results))
	}
}

func TestQuery_PrefixGlob(t *testing.T) {
	db := newTestDB()
	// "identity.a*" should match "aka" and "age"
	results := db.Query("identity.a*")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for identity.a*, got %d", len(results))
	}
	keys := map[string]bool{}
	for _, f := range results {
		keys[f.Key] = true
	}
	if !keys["aka"] || !keys["age"] {
		t.Errorf("expected aka and age, got %v", keys)
	}
}

func TestQuery_NoResults(t *testing.T) {
	db := newTestDB()
	results := db.Query("identity.zzz*")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestQuery_DescFieldsExcluded(t *testing.T) {
	db := newTestDB()
	results := db.Query("identity.*")
	for _, f := range results {
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded from query results", f.Key)
		}
	}
}

func TestQuery_NonexistentCategoryShorthand(t *testing.T) {
	db := newTestDB()
	results := db.Query("nonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent category, got %d", len(results))
	}
}

func TestQuery_CategoryGlobNoDot(t *testing.T) {
	db := newTestDB()
	// "w*" should match "web" category
	results := db.Query("w*")
	if len(results) != 2 {
		t.Fatalf("expected 2 results for w* glob, got %d", len(results))
	}
	for _, f := range results {
		if f.Category != "web" {
			t.Errorf("expected category web, got %s", f.Category)
		}
	}
}

// ---------------------------------------------------------------------------
// GetCategory
// ---------------------------------------------------------------------------

func TestGetCategory_Found(t *testing.T) {
	db := newTestDB()
	cat, ok := db.GetCategory("web")
	if !ok {
		t.Fatal("expected to find category web")
	}
	if cat.Name != "web" {
		t.Errorf("unexpected name: %s", cat.Name)
	}
	if len(cat.Fields) != 3 {
		t.Errorf("expected 3 fields in web, got %d", len(cat.Fields))
	}
}

func TestGetCategory_NotFound(t *testing.T) {
	db := newTestDB()
	_, ok := db.GetCategory("nonexistent")
	if ok {
		t.Error("expected not found for nonexistent category")
	}
}

// ---------------------------------------------------------------------------
// CategoryNames
// ---------------------------------------------------------------------------

func TestCategoryNames(t *testing.T) {
	db := newTestDB()
	names := db.CategoryNames()
	expected := []string{"identity", "web", "academic"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("index %d: expected %q, got %q", i, expected[i], name)
		}
	}
}

func TestCategoryNames_Empty(t *testing.T) {
	db := &DB{}
	names := db.CategoryNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestSearch_MatchInKey(t *testing.T) {
	db := newTestDB()
	results := db.Search("github")
	if len(results) != 1 {
		t.Fatalf("expected 1 result matching key 'github', got %d", len(results))
	}
	if results[0].Key != "github" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestSearch_MatchInValue(t *testing.T) {
	db := newTestDB()
	results := db.Search("Alexander")
	if len(results) != 1 {
		t.Fatalf("expected 1 result matching value 'Alexander', got %d", len(results))
	}
	if results[0].Key != "name" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestSearch_MatchInDescription(t *testing.T) {
	db := newTestDB()
	results := db.Search("ORCID persistent")
	if len(results) != 1 {
		t.Fatalf("expected 1 result matching desc 'ORCID persistent', got %d", len(results))
	}
	if results[0].Key != "orcid" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	db := newTestDB()
	results := db.Search("QUEELIUS")
	if len(results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive search, got %d", len(results))
	}
	if results[0].Key != "github" {
		t.Errorf("unexpected key: %s", results[0].Key)
	}
}

func TestSearch_NoMatches(t *testing.T) {
	db := newTestDB()
	results := db.Search("zzznoMatchHere")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_DescFieldsExcluded(t *testing.T) {
	db := newTestDB()
	// "companion desc" is the value of a _desc field; search should not return it
	results := db.Search("companion desc")
	for _, f := range results {
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded from search results", f.Key)
		}
	}
}

func TestSearch_MatchInArrayValue(t *testing.T) {
	db := newTestDB()
	// "Alex Towell" is an element in the aka []interface{} slice
	results := db.Search("Alex Towell")
	found := false
	for _, f := range results {
		if f.Key == "aka" {
			found = true
		}
	}
	if !found {
		t.Error("expected search to find the aka field via its array value")
	}
}

// ---------------------------------------------------------------------------
// AllFields
// ---------------------------------------------------------------------------

func TestAllFields_CorrectCount(t *testing.T) {
	db := newTestDB()
	fields := db.AllFields()
	// identity: 3 non-desc (name, aka, age), web: 2 (github, website), academic: 3 (orcid, gpa, topics) = 8
	if len(fields) != 8 {
		t.Errorf("expected 8 non-desc fields, got %d", len(fields))
	}
}

func TestAllFields_DescExcluded(t *testing.T) {
	db := newTestDB()
	for _, f := range db.AllFields() {
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded from AllFields", f.Key)
		}
	}
}

func TestAllFields_Empty(t *testing.T) {
	db := &DB{}
	fields := db.AllFields()
	if len(fields) != 0 {
		t.Errorf("expected 0 fields for empty DB, got %d", len(fields))
	}
}

// ---------------------------------------------------------------------------
// DescribeField
// ---------------------------------------------------------------------------

func TestDescribeField_WithDesc(t *testing.T) {
	db := newTestDB()
	desc := db.DescribeField("identity.name")
	if desc != "Full legal name" {
		t.Errorf("expected 'Full legal name', got %q", desc)
	}
}

func TestDescribeField_WithoutDesc(t *testing.T) {
	db := newTestDB()
	desc := db.DescribeField("identity.age")
	if desc != "" {
		t.Errorf("expected empty description, got %q", desc)
	}
}

func TestDescribeField_NotFound(t *testing.T) {
	db := newTestDB()
	desc := db.DescribeField("identity.nonexistent")
	if desc != "" {
		t.Errorf("expected empty string for nonexistent field, got %q", desc)
	}
}

func TestDescribeField_InvalidPath(t *testing.T) {
	db := newTestDB()
	desc := db.DescribeField("nopath")
	if desc != "" {
		t.Errorf("expected empty string for invalid path, got %q", desc)
	}
}

// ---------------------------------------------------------------------------
// DescribeCategory
// ---------------------------------------------------------------------------

func TestDescribeCategory_WithDescriptions(t *testing.T) {
	db := newTestDB()
	fields := db.DescribeCategory("identity")
	// identity has "name" (has desc) and "aka" (has desc), age has no desc
	if len(fields) != 2 {
		t.Fatalf("expected 2 described fields in identity, got %d", len(fields))
	}
	for _, f := range fields {
		if f.Desc == "" {
			t.Errorf("field %q should have a description", f.Key)
		}
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded", f.Key)
		}
	}
}

func TestDescribeCategory_NotFound(t *testing.T) {
	db := newTestDB()
	fields := db.DescribeCategory("nonexistent")
	if fields != nil {
		t.Errorf("expected nil for nonexistent category, got %v", fields)
	}
}

func TestDescribeCategory_EmptyCategory(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{Name: "empty", Fields: []Field{}},
		},
	}
	fields := db.DescribeCategory("empty")
	if len(fields) != 0 {
		t.Errorf("expected 0 described fields for empty category, got %d", len(fields))
	}
}

func TestDescribeCategory_NoDescriptions(t *testing.T) {
	db := &DB{
		Categories: []Category{
			{Name: "nodesc", Fields: []Field{
				{Key: "x", Value: "val", Desc: "", Category: "nodesc"},
			}},
		},
	}
	fields := db.DescribeCategory("nodesc")
	if len(fields) != 0 {
		t.Errorf("expected 0 described fields when no descs, got %d", len(fields))
	}
}

// ---------------------------------------------------------------------------
// AllDescriptions
// ---------------------------------------------------------------------------

func TestAllDescriptions(t *testing.T) {
	db := newTestDB()
	fields := db.AllDescriptions()
	// Fields with non-empty Desc (excluding _desc keys):
	// identity.name, identity.aka, web.github, web.website, academic.orcid, academic.topics = 6
	if len(fields) != 6 {
		t.Fatalf("expected 6 described fields, got %d", len(fields))
	}
	for _, f := range fields {
		if f.Desc == "" {
			t.Errorf("field %q should have a description", f.Key)
		}
		if IsDescKey(f.Key) {
			t.Errorf("_desc key %q should be excluded", f.Key)
		}
	}
}

func TestAllDescriptions_Empty(t *testing.T) {
	db := &DB{}
	fields := db.AllDescriptions()
	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

// ---------------------------------------------------------------------------
// FormatValue
// ---------------------------------------------------------------------------

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "[]interface{}",
			input:    []interface{}{"one", "two", "three"},
			expected: "one, two, three",
		},
		{
			name:     "[]interface{} single element",
			input:    []interface{}{"solo"},
			expected: "solo",
		},
		{
			name:     "[]interface{} empty",
			input:    []interface{}{},
			expected: "",
		},
		{
			name:     "[]interface{} with mixed types",
			input:    []interface{}{"text", 42, 3.14},
			expected: "text, 42, 3.14",
		},
		{
			name:     "[]string",
			input:    []string{"alpha", "beta"},
			expected: "alpha, beta",
		},
		{
			name:     "[]string empty",
			input:    []string{},
			expected: "",
		},
		{
			name:     "int64",
			input:    int64(42),
			expected: "42",
		},
		{
			name:     "int64 zero",
			input:    int64(0),
			expected: "0",
		},
		{
			name:     "int64 negative",
			input:    int64(-10),
			expected: "-10",
		},
		{
			name:     "float64",
			input:    float64(3.14),
			expected: "3.14",
		},
		{
			name:     "float64 zero",
			input:    float64(0),
			expected: "0",
		},
		{
			name:     "nil",
			input:    nil,
			expected: "<nil>",
		},
		{
			name:     "bool fallback",
			input:    true,
			expected: "true",
		},
		{
			name:     "int fallback",
			input:    42,
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValue(tt.input)
			if result != tt.expected {
				t.Errorf("FormatValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// IsDescKey
// ---------------------------------------------------------------------------

func TestIsDescKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"name_desc", true},
		{"github_desc", true},
		{"_desc", true},
		{"long_key_name_desc", true},
		{"name", false},
		{"description", false},
		{"desc", false},
		{"desc_name", false},
		{"", false},
		{"_descx", false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := IsDescKey(tt.key); got != tt.expected {
				t.Errorf("IsDescKey(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BaseKey
// ---------------------------------------------------------------------------

func TestBaseKey(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"name_desc", "name"},
		{"github_desc", "github"},
		{"_desc", ""},
		{"long_key_name_desc", "long_key_name"},
		{"name", "name"},
		{"description", "description"},
		{"desc", "desc"},
		{"", ""},
		{"nodesc", "nodesc"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := BaseKey(tt.key); got != tt.expected {
				t.Errorf("BaseKey(%q) = %q, want %q", tt.key, got, tt.expected)
			}
		})
	}
}
