package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatTable renders a slice of fields as an aligned text table.
//
// If all fields belong to the same category, the Category column is omitted.
// Output example:
//
//	Category    Key       Value
//	────────    ───       ─────
//	identity    name      Alexander Towell
//	identity    aka       Alex Towell
//	web         github    queelius
func FormatTable(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}

	// Determine whether we need the category column.
	multiCat := hasMultipleCategories(fields)

	// Compute column widths.
	catWidth := len("Category")
	keyWidth := len("Key")
	valWidth := len("Value")

	for _, f := range fields {
		if multiCat && len(f.Category) > catWidth {
			catWidth = len(f.Category)
		}
		if len(f.Key) > keyWidth {
			keyWidth = len(f.Key)
		}
		v := FormatValue(f.Value)
		if len(v) > valWidth {
			valWidth = len(v)
		}
	}

	var b strings.Builder

	if multiCat {
		// Header
		fmt.Fprintf(&b, "%-*s    %-*s    %s\n", catWidth, "Category", keyWidth, "Key", "Value")
		// Separator using Unicode box-drawing dash
		fmt.Fprintf(&b, "%-*s    %-*s    %s\n",
			catWidth, repeatRune('\u2500', catWidth),
			keyWidth, repeatRune('\u2500', keyWidth),
			repeatRune('\u2500', valWidth))
		// Rows
		for _, f := range fields {
			fmt.Fprintf(&b, "%-*s    %-*s    %s\n", catWidth, f.Category, keyWidth, f.Key, FormatValue(f.Value))
		}
	} else {
		// Header (no category column)
		fmt.Fprintf(&b, "%-*s    %s\n", keyWidth, "Key", "Value")
		fmt.Fprintf(&b, "%-*s    %s\n",
			keyWidth, repeatRune('\u2500', keyWidth),
			repeatRune('\u2500', valWidth))
		// Rows
		for _, f := range fields {
			fmt.Fprintf(&b, "%-*s    %s\n", keyWidth, f.Key, FormatValue(f.Value))
		}
	}

	return b.String()
}

// FormatJSON serializes the entire DB as a JSON object grouped by category.
//
// Output example:
//
//	{
//	  "identity": {
//	    "name": "Alexander Towell",
//	    "aka": ["Alex Towell"]
//	  },
//	  "web": { ... }
//	}
//
// Fields with _desc keys are excluded from the output.
func FormatJSON(db *DB) (string, error) {
	root := buildCategoryMap(db)
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal DB to JSON: %w", err)
	}
	return string(data), nil
}

// FormatCategoryJSON serializes a single category as a flat JSON object.
// Fields with _desc keys are excluded.
func FormatCategoryJSON(cat Category) (string, error) {
	obj := buildFieldMap(cat.Fields)
	data, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal category %q to JSON: %w", cat.Name, err)
	}
	return string(data), nil
}

// FormatFieldsJSON serializes a slice of fields as JSON. If all fields share
// the same category, a flat object is produced. If fields span multiple
// categories, they are grouped by category name.
func FormatFieldsJSON(fields []Field) (string, error) {
	if len(fields) == 0 {
		data, err := json.MarshalIndent(map[string]interface{}{}, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal empty fields to JSON: %w", err)
		}
		return string(data), nil
	}

	if !hasMultipleCategories(fields) {
		obj := buildFieldMap(fields)
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal fields to JSON: %w", err)
		}
		return string(data), nil
	}

	// Group by category, preserving order.
	grouped := groupByCategory(fields)
	data, err := json.MarshalIndent(grouped, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal grouped fields to JSON: %w", err)
	}
	return string(data), nil
}

// FormatEnv formats the entire DB as environment variable assignments.
//
// Key format: DEETS_<CATEGORY>_<KEY> (uppercased).
// Values are double-quoted. For slice values, elements are comma-separated.
//
// Example:
//
//	DEETS_IDENTITY_NAME="Alexander Towell"
//	DEETS_WEB_GITHUB="queelius"
func FormatEnv(db *DB) string {
	var b strings.Builder
	for _, cat := range db.Categories {
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			envKey := fmt.Sprintf("DEETS_%s_%s",
				strings.ToUpper(cat.Name),
				strings.ToUpper(f.Key))
			b.WriteString(fmt.Sprintf("%s=%q\n", envKey, FormatValue(f.Value)))
		}
	}
	return b.String()
}

// FormatTOML formats the entire DB as a TOML document.
//
// Each category becomes a TOML table header. String values are quoted,
// arrays are rendered as TOML arrays, and numeric types are unquoted.
// _desc fields are excluded.
func FormatTOML(db *DB) string {
	var b strings.Builder
	for i, cat := range db.Categories {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "[%s]\n", cat.Name)
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			b.WriteString(fmt.Sprintf("%s = %s\n", f.Key, tomlValue(f.Value)))
		}
	}
	return b.String()
}

// FormatYAML formats the entire DB as a YAML document.
//
// Each category is a top-level mapping key. String values are unquoted (unless
// they require quoting), arrays use the flow sequence syntax, and numeric types
// are rendered directly. _desc fields are excluded.
func FormatYAML(db *DB) string {
	var b strings.Builder
	for i, cat := range db.Categories {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "%s:\n", cat.Name)
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s: %s\n", f.Key, yamlValue(f.Value)))
		}
	}
	return b.String()
}

// FormatDescTable renders a table of field paths and their descriptions.
//
// Output example:
//
//	Field               Description
//	─────               ───────────
//	identity.name       Full legal name
//	academic.orcid      ORCID persistent digital identifier
func FormatDescTable(fields []Field) string {
	if len(fields) == 0 {
		return ""
	}

	fieldWidth := len("Field")
	descWidth := len("Description")

	for _, f := range fields {
		path := f.Category + "." + f.Key
		if len(path) > fieldWidth {
			fieldWidth = len(path)
		}
		if len(f.Desc) > descWidth {
			descWidth = len(f.Desc)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s    %s\n", fieldWidth, "Field", "Description")
	fmt.Fprintf(&b, "%-*s    %s\n",
		fieldWidth, repeatRune('\u2500', fieldWidth),
		repeatRune('\u2500', descWidth))
	for _, f := range fields {
		path := f.Category + "." + f.Key
		fmt.Fprintf(&b, "%-*s    %s\n", fieldWidth, path, f.Desc)
	}
	return b.String()
}

// FormatDescJSON serializes field descriptions as a JSON object mapping
// "category.key" to description strings.
func FormatDescJSON(fields []Field) (string, error) {
	m := orderedMap{values: make(map[string]interface{})}
	for _, f := range fields {
		path := f.Category + "." + f.Key
		m.keys = append(m.keys, path)
		m.values[path] = f.Desc
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal descriptions to JSON: %w", err)
	}
	return string(data), nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// hasMultipleCategories reports whether the fields span more than one category.
func hasMultipleCategories(fields []Field) bool {
	if len(fields) == 0 {
		return false
	}
	first := fields[0].Category
	for _, f := range fields[1:] {
		if f.Category != first {
			return true
		}
	}
	return false
}

// repeatRune returns a string of n repetitions of the given rune.
func repeatRune(r rune, n int) string {
	return strings.Repeat(string(r), n)
}

// orderedMap is a helper that preserves insertion order when marshaled to JSON.
type orderedMap struct {
	keys   []string
	values map[string]interface{}
}

// MarshalJSON serializes the orderedMap to JSON preserving key order.
func (o orderedMap) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	b.WriteString("{")
	for i, k := range o.keys {
		if i > 0 {
			b.WriteString(",")
		}
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		valJSON, err := json.Marshal(o.values[k])
		if err != nil {
			return nil, err
		}
		b.Write(keyJSON)
		b.WriteString(":")
		b.Write(valJSON)
	}
	b.WriteString("}")
	return []byte(b.String()), nil
}

// buildCategoryMap creates an ordered map of the entire DB for JSON output.
func buildCategoryMap(db *DB) orderedMap {
	om := orderedMap{values: make(map[string]interface{})}
	for _, cat := range db.Categories {
		catMap := buildFieldMap(cat.Fields)
		if len(catMap.keys) > 0 {
			om.keys = append(om.keys, cat.Name)
			om.values[cat.Name] = catMap
		}
	}
	return om
}

// buildFieldMap creates an ordered map from a slice of fields, excluding _desc keys.
func buildFieldMap(fields []Field) orderedMap {
	om := orderedMap{values: make(map[string]interface{})}
	for _, f := range fields {
		if IsDescKey(f.Key) {
			continue
		}
		om.keys = append(om.keys, f.Key)
		om.values[f.Key] = f.Value
	}
	return om
}

// groupByCategory groups fields by their category, preserving order,
// and returns an ordered map suitable for JSON serialization.
func groupByCategory(fields []Field) orderedMap {
	om := orderedMap{values: make(map[string]interface{})}
	seen := make(map[string]bool)
	catFields := make(map[string][]Field)

	for _, f := range fields {
		if !seen[f.Category] {
			seen[f.Category] = true
			om.keys = append(om.keys, f.Category)
		}
		catFields[f.Category] = append(catFields[f.Category], f)
	}
	for _, catName := range om.keys {
		om.values[catName] = buildFieldMap(catFields[catName])
	}
	return om
}

// tomlValue formats a Go value as a TOML value literal.
func tomlValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case []interface{}:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, tomlValue(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []string:
		parts := make([]string, 0, len(val))
		for _, s := range val {
			parts = append(parts, fmt.Sprintf("%q", s))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case int64:
		return fmt.Sprint(val)
	case float64:
		return fmt.Sprint(val)
	case bool:
		return fmt.Sprint(val)
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", v))
	}
}

// yamlValue formats a Go value as a YAML value literal.
func yamlValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		if yamlNeedsQuoting(val) {
			return fmt.Sprintf("%q", val)
		}
		return val
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}
		parts := make([]string, 0, len(val))
		for _, item := range val {
			parts = append(parts, yamlValue(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case []string:
		if len(val) == 0 {
			return "[]"
		}
		parts := make([]string, 0, len(val))
		for _, s := range val {
			if yamlNeedsQuoting(s) {
				parts = append(parts, fmt.Sprintf("%q", s))
			} else {
				parts = append(parts, s)
			}
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case int64:
		return fmt.Sprint(val)
	case float64:
		return fmt.Sprint(val)
	case bool:
		return fmt.Sprint(val)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// yamlNeedsQuoting reports whether a YAML string value requires quoting
// to avoid ambiguity with YAML special values or characters.
func yamlNeedsQuoting(s string) bool {
	if s == "" {
		return true
	}
	// Values that YAML would interpret as special types.
	lower := strings.ToLower(s)
	switch lower {
	case "true", "false", "yes", "no", "on", "off", "null", "~":
		return true
	}
	// If it starts or ends with whitespace, or contains characters that
	// could confuse a YAML parser.
	if s[0] == ' ' || s[len(s)-1] == ' ' {
		return true
	}
	for _, c := range s {
		switch c {
		case ':', '#', '[', ']', '{', '}', ',', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
			return true
		}
	}
	return false
}
