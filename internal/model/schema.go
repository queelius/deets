package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SchemaField describes a single field's schema metadata.
type SchemaField struct {
	Category    string `json:"category"`
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// InferType returns a human-readable type name for the given value.
func InferType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case []interface{}:
		return "array"
	case []string:
		return "array"
	case int64:
		return "integer"
	case float64:
		return "float"
	case bool:
		return "boolean"
	default:
		return "unknown"
	}
}

// BuildSchema constructs schema entries for every non-desc field in the DB.
func BuildSchema(db *DB) []SchemaField {
	var schema []SchemaField
	for _, cat := range db.Categories {
		for _, f := range cat.Fields {
			if IsDescKey(f.Key) {
				continue
			}
			schema = append(schema, SchemaField{
				Category:    cat.Name,
				Key:         f.Key,
				Type:        InferType(f.Value),
				Description: f.Desc,
				Example:     FormatValue(f.Value),
			})
		}
	}
	return schema
}

// FormatSchemaTable renders a schema table.
func FormatSchemaTable(entries []SchemaField) string {
	if len(entries) == 0 {
		return ""
	}

	catWidth := len("Category")
	keyWidth := len("Key")
	typeWidth := len("Type")
	descWidth := len("Description")
	exWidth := len("Example")

	for _, e := range entries {
		if len(e.Category) > catWidth {
			catWidth = len(e.Category)
		}
		if len(e.Key) > keyWidth {
			keyWidth = len(e.Key)
		}
		if len(e.Type) > typeWidth {
			typeWidth = len(e.Type)
		}
		if len(e.Description) > descWidth {
			descWidth = len(e.Description)
		}
		if len(e.Example) > exWidth {
			exWidth = len(e.Example)
		}
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%-*s    %-*s    %-*s    %-*s    %s\n",
		catWidth, "Category", keyWidth, "Key", typeWidth, "Type", descWidth, "Description", "Example")
	fmt.Fprintf(&b, "%-*s    %-*s    %-*s    %-*s    %s\n",
		catWidth, strings.Repeat("\u2500", catWidth),
		keyWidth, strings.Repeat("\u2500", keyWidth),
		typeWidth, strings.Repeat("\u2500", typeWidth),
		descWidth, strings.Repeat("\u2500", descWidth),
		strings.Repeat("\u2500", exWidth))
	for _, e := range entries {
		fmt.Fprintf(&b, "%-*s    %-*s    %-*s    %-*s    %s\n",
			catWidth, e.Category, keyWidth, e.Key, typeWidth, e.Type, descWidth, e.Description, e.Example)
	}
	return b.String()
}

// FormatSchemaJSON serializes schema entries as a JSON array.
func FormatSchemaJSON(entries []SchemaField) (string, error) {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal schema to JSON: %w", err)
	}
	return string(data), nil
}
