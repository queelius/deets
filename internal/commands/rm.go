package commands

import (
	"fmt"
	"strings"

	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(rmCmd)
}

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

		// Validate the path early: if it has no dots it must be a valid
		// category name; if it has dots, at least one valid interpretation
		// (category or category.key) must parse.
		if !strings.Contains(path, ".") {
			if err := store.ValidateCategoryName(path); err != nil {
				return fmt.Errorf("invalid path %q: %w", path, err)
			}
			return store.RemoveCategory(filePath, path)
		}

		// Path contains dots. Check if the full path is a category name
		// (e.g. "profiles.github") before treating it as category.key.
		if err := store.ValidateCategoryName(path); err == nil {
			if store.HasCategory(filePath, path) {
				return store.RemoveCategory(filePath, path)
			}
		}

		// Not a category — treat as a field path.
		cat, key, err := parsePath(path)
		if err != nil {
			return err
		}
		return store.RemoveValue(filePath, cat, key)
	},
}
