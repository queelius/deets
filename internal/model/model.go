// Package model defines the core data structures and query methods for the
// deets personal metadata database. It represents a collection of categorized
// fields, each holding a key-value pair with an optional description.
package model

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Field represents a single metadata entry within a category.
type Field struct {
	// Key is the field name (e.g., "name", "email").
	Key string
	// Value is the field's value, which may be a string, []string, number, etc.
	Value interface{}
	// Desc is the human-readable description of this field.
	Desc string
	// Category is the name of the category this field belongs to.
	Category string
}

// Category represents a named group of related fields.
type Category struct {
	// Name is the category identifier (e.g., "identity", "web").
	Name string
	// Fields is the ordered list of fields within this category.
	Fields []Field
}

// DB is the top-level container for the entire metadata database,
// organized as an ordered list of categories.
type DB struct {
	// Categories is the ordered list of all categories in the database.
	Categories []Category
}

// GetField retrieves a single field by its "category.key" path.
// Returns the field and true if found, or a zero Field and false otherwise.
func (db *DB) GetField(path string) (Field, bool) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) != 2 {
		return Field{}, false
	}
	catName, key := parts[0], parts[1]

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

// Query performs a glob-based query against the database fields.
//
// Supported patterns:
//   - "category.key"    — exact match for a specific field
//   - "category" or "category.*" — all fields in the named category (excluding _desc fields)
//   - "*.key"           — find a key across all categories
//   - "category.prefix*" — glob match within a category
//
// The function uses filepath.Match for glob semantics and always excludes
// _desc fields from results.
func (db *DB) Query(pattern string) []Field {
	var results []Field

	// If pattern has no dot, treat it as "category" shorthand for "category.*"
	if !strings.Contains(pattern, ".") {
		// Check if this matches a category name exactly
		for _, cat := range db.Categories {
			if cat.Name == pattern {
				for _, f := range cat.Fields {
					if !IsDescKey(f.Key) {
						results = append(results, f)
					}
				}
				return results
			}
		}
		// If it doesn't match a category, try it as a glob against category names
		for _, cat := range db.Categories {
			matched, err := filepath.Match(pattern, cat.Name)
			if err == nil && matched {
				for _, f := range cat.Fields {
					if !IsDescKey(f.Key) {
						results = append(results, f)
					}
				}
			}
		}
		return results
	}

	parts := strings.SplitN(pattern, ".", 2)
	catPattern, keyPattern := parts[0], parts[1]

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

// GetCategory retrieves a category by name.
// Returns the category and true if found, or a zero Category and false otherwise.
func (db *DB) GetCategory(name string) (Category, bool) {
	for _, cat := range db.Categories {
		if cat.Name == name {
			return cat, true
		}
	}
	return Category{}, false
}

// CategoryNames returns the names of all categories in their original order.
func (db *DB) CategoryNames() []string {
	names := make([]string, 0, len(db.Categories))
	for _, cat := range db.Categories {
		names = append(names, cat.Name)
	}
	return names
}

// Search performs a case-insensitive search across all field keys, values,
// and descriptions, returning every field that contains the query string.
// Results exclude _desc fields.
func (db *DB) Search(query string) []Field {
	var results []Field
	q := strings.ToLower(query)

	for _, cat := range db.Categories {
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			if containsLower(f.Key, q) ||
				containsLower(FormatValue(f.Value), q) ||
				containsLower(f.Desc, q) {
				results = append(results, f)
			}
		}
	}
	return results
}

// AllFields returns every field in the database, excluding _desc fields,
// in category order.
func (db *DB) AllFields() []Field {
	var results []Field
	for _, cat := range db.Categories {
		for _, f := range cat.Fields {
			if !IsDescKey(f.Key) {
				results = append(results, f)
			}
		}
	}
	return results
}

// DescribeField returns the description for the field identified by the
// "category.key" path. If the field has no description, an empty string
// is returned.
func (db *DB) DescribeField(path string) string {
	f, ok := db.GetField(path)
	if !ok {
		return ""
	}
	return f.Desc
}

// DescribeCategory returns all fields within the named category that have
// a non-empty description.
func (db *DB) DescribeCategory(name string) []Field {
	cat, ok := db.GetCategory(name)
	if !ok {
		return nil
	}
	var results []Field
	for _, f := range cat.Fields {
		if IsDescKey(f.Key) {
			continue
		}
		if f.Desc != "" {
			results = append(results, f)
		}
	}
	return results
}

// AllDescriptions returns every field across the entire database that has
// a non-empty description, excluding _desc fields.
func (db *DB) AllDescriptions() []Field {
	var results []Field
	for _, cat := range db.Categories {
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			if f.Desc != "" {
				results = append(results, f)
			}
		}
	}
	return results
}

// FormatValue converts a field value to a human-readable string for display.
//
// Formatting rules:
//   - string: returned as-is
//   - []interface{}: elements joined with ", "
//   - []string: elements joined with ", "
//   - int64/float64: formatted with fmt.Sprint
//   - fallback: formatted with fmt.Sprintf("%v", v)
func FormatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, fmt.Sprintf("%v", item))
		}
		return strings.Join(parts, ", ")
	case []string:
		return strings.Join(val, ", ")
	case int64:
		return fmt.Sprint(val)
	case float64:
		return fmt.Sprint(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// IsDescKey reports whether the given key is a description companion field,
// identified by the "_desc" suffix.
func IsDescKey(key string) bool {
	return strings.HasSuffix(key, "_desc")
}

// BaseKey strips the "_desc" suffix from a key if present, returning the
// base field name.
func BaseKey(key string) string {
	return strings.TrimSuffix(key, "_desc")
}

// containsLower checks whether s (lowercased) contains the already-lowered
// substring q.
func containsLower(s, q string) bool {
	return strings.Contains(strings.ToLower(s), q)
}
