package store

import (
	"sort"

	"github.com/queelius/deets/internal/model"
)

// Merge merges a local override DB into a global DB and returns a new DB.
// Local keys replace matching global keys within each category. Non-overlapping
// keys from both are preserved. Categories that exist only in local or only in
// global are included. The result is sorted alphabetically by category and by
// field key within each category.
func Merge(global, local *model.DB) *model.DB {
	// Index global categories by name for efficient lookup.
	globalIdx := make(map[string]int, len(global.Categories))
	for i, cat := range global.Categories {
		globalIdx[cat.Name] = i
	}

	// Index local categories by name.
	localIdx := make(map[string]int, len(local.Categories))
	for i, cat := range local.Categories {
		localIdx[cat.Name] = i
	}

	// Collect all unique category names.
	catSet := make(map[string]struct{})
	for _, cat := range global.Categories {
		catSet[cat.Name] = struct{}{}
	}
	for _, cat := range local.Categories {
		catSet[cat.Name] = struct{}{}
	}

	catNames := make([]string, 0, len(catSet))
	for name := range catSet {
		catNames = append(catNames, name)
	}
	sort.Strings(catNames)

	merged := &model.DB{}

	for _, catName := range catNames {
		gIdx, gOK := globalIdx[catName]
		lIdx, lOK := localIdx[catName]

		switch {
		case gOK && !lOK:
			// Category only in global — keep as-is.
			merged.Categories = append(merged.Categories, global.Categories[gIdx])

		case !gOK && lOK:
			// Category only in local — add it.
			merged.Categories = append(merged.Categories, local.Categories[lIdx])

		case gOK && lOK:
			// Both have this category — merge at key level.
			merged.Categories = append(merged.Categories, mergeCategory(
				global.Categories[gIdx],
				local.Categories[lIdx],
			))
		}
	}

	return merged
}

// mergeCategory merges fields from a local category into a global category.
// Local fields override global fields with the same key. All other fields are
// preserved and the result is sorted alphabetically by key.
func mergeCategory(global, local model.Category) model.Category {
	// Build a map of global fields.
	fieldMap := make(map[string]model.Field, len(global.Fields))
	for _, f := range global.Fields {
		fieldMap[f.Key] = f
	}

	// Local fields override globals.
	for _, f := range local.Fields {
		fieldMap[f.Key] = f
	}

	// Collect and sort keys.
	keys := make([]string, 0, len(fieldMap))
	for k := range fieldMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	cat := model.Category{Name: global.Name}
	for _, k := range keys {
		cat.Fields = append(cat.Fields, fieldMap[k])
	}
	return cat
}
