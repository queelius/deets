// Package store reads TOML files and produces a model.DB. It also provides
// merge logic and line-level TOML editing that preserves comments and formatting.
package store

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/queelius/deets/internal/model"
)

// LoadFile reads a single TOML file at path and parses it into a *model.DB.
// Each top-level key in the TOML is treated as a category name whose value is
// a map of field keys to values. Keys ending in "_desc" are treated as
// descriptions for their companion field (e.g., "email_desc" describes "email").
//
// Nested TOML tables (e.g., [profiles.github]) are flattened into dotted
// category names. BurntSushi/toml parses [profiles.github] as a nested map
// {"profiles": {"github": {...}}}; this function recursively walks the map
// and produces a category named "profiles.github".
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

// collectCategories recursively walks the parsed TOML map and appends
// categories to db. When a table value is itself a map, the parent and
// child keys are joined with a dot to form the category name.
func collectCategories(db *model.DB, prefix string, raw map[string]interface{}) {
	for k, v := range raw {
		m, ok := v.(map[string]interface{})
		if !ok {
			continue // skip non-table top-level values
		}

		fullName := k
		if prefix != "" {
			fullName = prefix + "." + k
		}

		// Separate leaf fields (scalars/arrays) from sub-tables (nested maps).
		leafMap := make(map[string]interface{})
		subTables := make(map[string]interface{})
		for fk, fv := range m {
			if _, isMap := fv.(map[string]interface{}); isMap {
				subTables[fk] = fv
			} else {
				leafMap[fk] = fv
			}
		}

		// If there are leaf fields, create a category for them.
		if len(leafMap) > 0 {
			cat := buildCategory(fullName, leafMap)
			if len(cat.Fields) > 0 {
				db.Categories = append(db.Categories, cat)
			}
		}

		// If there are sub-tables, recurse into them.
		if len(subTables) > 0 {
			subMap := make(map[string]interface{})
			for sk, sv := range subTables {
				subMap[sk] = sv
			}
			collectCategories(db, fullName, subMap)
		}
	}
}

// buildCategory creates a model.Category from a flat TOML table map.
// Keys ending in "_desc" are consumed as descriptions for their companion
// fields and are not included as separate fields.
func buildCategory(catName string, catMap map[string]interface{}) model.Category {
	// Collect non-desc keys and sort alphabetically.
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

		// Look for a companion _desc key in the TOML data.
		if desc, ok := catMap[key+"_desc"]; ok {
			if s, ok := desc.(string); ok {
				f.Desc = s
			}
		}

		// Fall back to DefaultDescriptions if no desc was found.
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

// Load reads the global TOML file and optionally merges it with a local
// override file. If localPath is empty, only the global file is loaded.
func Load(globalPath, localPath string) (*model.DB, error) {
	global, err := LoadFile(globalPath)
	if err != nil {
		return nil, err
	}

	if localPath == "" {
		return global, nil
	}

	local, err := LoadFile(localPath)
	if err != nil {
		return nil, err
	}

	return Merge(global, local), nil
}
