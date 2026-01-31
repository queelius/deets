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

	// Collect and sort category names alphabetically.
	catNames := make([]string, 0, len(raw))
	for name := range raw {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	for _, catName := range catNames {
		catVal := raw[catName]
		catMap, ok := catVal.(map[string]interface{})
		if !ok {
			continue
		}

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

		// Skip empty categories (e.g., section headers with only commented-out fields).
		if len(cat.Fields) > 0 {
			db.Categories = append(db.Categories, cat)
		}
	}

	return db, nil
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
