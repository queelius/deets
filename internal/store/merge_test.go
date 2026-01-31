package store

import (
	"testing"

	"github.com/queelius/deets/internal/model"
)

func TestMerge_OverlappingCategoriesWithKeyReplacement(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
					{Key: "pronouns", Value: "she/her", Category: "identity"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Bob", Category: "identity"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 1 {
		t.Fatalf("expected 1 category, got %d", len(merged.Categories))
	}

	cat := merged.Categories[0]
	if cat.Name != "identity" {
		t.Errorf("expected category 'identity', got %q", cat.Name)
	}

	// Both fields should be present: name (overridden) and pronouns (from global).
	if len(cat.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(cat.Fields))
	}

	// Fields are sorted alphabetically: name, pronouns
	nameField := findField(cat.Fields, "name")
	if nameField == nil {
		t.Fatal("expected 'name' field")
	}
	if nameField.Value != "Bob" {
		t.Errorf("expected name = 'Bob' (local override), got %v", nameField.Value)
	}

	pronounsField := findField(cat.Fields, "pronouns")
	if pronounsField == nil {
		t.Fatal("expected 'pronouns' field from global")
	}
	if pronounsField.Value != "she/her" {
		t.Errorf("expected pronouns = 'she/her', got %v", pronounsField.Value)
	}
}

func TestMerge_NonOverlappingCategoriesPreserved(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "contact",
				Fields: []model.Field{
					{Key: "email", Value: "alice@example.com", Category: "contact"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(merged.Categories))
	}

	// Sorted alphabetically: contact, identity
	if merged.Categories[0].Name != "contact" {
		t.Errorf("expected first category 'contact', got %q", merged.Categories[0].Name)
	}
	if merged.Categories[1].Name != "identity" {
		t.Errorf("expected second category 'identity', got %q", merged.Categories[1].Name)
	}
}

func TestMerge_LocalOnlyCategoriesAdded(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "web",
				Fields: []model.Field{
					{Key: "github", Value: "alice", Category: "web"},
					{Key: "blog", Value: "https://alice.dev", Category: "web"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(merged.Categories))
	}

	webCat := findCategory(merged.Categories, "web")
	if webCat == nil {
		t.Fatal("expected 'web' category from local to be present")
	}
	if len(webCat.Fields) != 2 {
		t.Fatalf("expected 2 fields in web, got %d", len(webCat.Fields))
	}
}

func TestMerge_GlobalOnlyCategoriesPreserved(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
			{
				Name: "academic",
				Fields: []model.Field{
					{Key: "institution", Value: "MIT", Category: "academic"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Bob", Category: "identity"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(merged.Categories))
	}

	acadCat := findCategory(merged.Categories, "academic")
	if acadCat == nil {
		t.Fatal("expected global-only 'academic' category to be preserved")
	}
	if len(acadCat.Fields) != 1 {
		t.Fatalf("expected 1 field in academic, got %d", len(acadCat.Fields))
	}
	if acadCat.Fields[0].Value != "MIT" {
		t.Errorf("expected institution = 'MIT', got %v", acadCat.Fields[0].Value)
	}
}

func TestMerge_EmptyLocal(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
			{
				Name: "contact",
				Fields: []model.Field{
					{Key: "email", Value: "alice@example.com", Category: "contact"},
				},
			},
		},
	}

	local := &model.DB{}

	merged := Merge(global, local)

	if len(merged.Categories) != 2 {
		t.Fatalf("expected 2 categories (all from global), got %d", len(merged.Categories))
	}

	// Categories should match global exactly.
	if merged.Categories[0].Name != "contact" {
		t.Errorf("expected first category 'contact', got %q", merged.Categories[0].Name)
	}
	if merged.Categories[1].Name != "identity" {
		t.Errorf("expected second category 'identity', got %q", merged.Categories[1].Name)
	}
}

func TestMerge_EmptyGlobal(t *testing.T) {
	global := &model.DB{}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "web",
				Fields: []model.Field{
					{Key: "github", Value: "alice", Category: "web"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 1 {
		t.Fatalf("expected 1 category (from local), got %d", len(merged.Categories))
	}
	if merged.Categories[0].Name != "web" {
		t.Errorf("expected category 'web', got %q", merged.Categories[0].Name)
	}
}

func TestMerge_BothEmpty(t *testing.T) {
	global := &model.DB{}
	local := &model.DB{}

	merged := Merge(global, local)

	if len(merged.Categories) != 0 {
		t.Errorf("expected 0 categories for both empty, got %d", len(merged.Categories))
	}
}

func TestMerge_CategoriesSortedAlphabetically(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "web",
				Fields: []model.Field{
					{Key: "github", Value: "alice", Category: "web"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "academic",
				Fields: []model.Field{
					{Key: "institution", Value: "MIT", Category: "academic"},
				},
			},
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
		},
	}

	merged := Merge(global, local)

	if len(merged.Categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(merged.Categories))
	}

	expectedOrder := []string{"academic", "identity", "web"}
	for i, cat := range merged.Categories {
		if cat.Name != expectedOrder[i] {
			t.Errorf("category[%d]: expected %q, got %q", i, expectedOrder[i], cat.Name)
		}
	}
}

func TestMerge_FieldsSortedAlphabeticallyInMergedCategory(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "pronouns", Value: "she/her", Category: "identity"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "aka", Value: "Nickname", Category: "identity"},
					{Key: "name", Value: "Alice", Category: "identity"},
				},
			},
		},
	}

	merged := Merge(global, local)

	cat := merged.Categories[0]
	if len(cat.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(cat.Fields))
	}

	expectedKeys := []string{"aka", "name", "pronouns"}
	for i, f := range cat.Fields {
		if f.Key != expectedKeys[i] {
			t.Errorf("field[%d]: expected key %q, got %q", i, expectedKeys[i], f.Key)
		}
	}
}

func TestMerge_LocalOverridesDescriptions(t *testing.T) {
	global := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Alice", Desc: "Full name", Category: "identity"},
				},
			},
		},
	}

	local := &model.DB{
		Categories: []model.Category{
			{
				Name: "identity",
				Fields: []model.Field{
					{Key: "name", Value: "Bob", Desc: "Project name", Category: "identity"},
				},
			},
		},
	}

	merged := Merge(global, local)

	nameField := findField(merged.Categories[0].Fields, "name")
	if nameField == nil {
		t.Fatal("expected 'name' field")
	}
	if nameField.Desc != "Project name" {
		t.Errorf("expected desc from local 'Project name', got %q", nameField.Desc)
	}
}

// -- test helpers --

func findField(fields []model.Field, key string) *model.Field {
	for i, f := range fields {
		if f.Key == key {
			return &fields[i]
		}
	}
	return nil
}

func findCategory(categories []model.Category, name string) *model.Category {
	for i, cat := range categories {
		if cat.Name == name {
			return &categories[i]
		}
	}
	return nil
}
