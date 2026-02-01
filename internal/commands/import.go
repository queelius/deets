package commands

import (
	"fmt"
	"strings"

	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

var flagImportDryRun bool

func init() {
	importCmd.Flags().BoolVar(&flagImportDryRun, "dry-run", false, "show what would change without writing")
	rootCmd.AddCommand(importCmd)
}

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import fields from a TOML file",
	Long: `Import fields from a TOML file into the deets store.

Each field in the import file is written to the target file using
line-level editing to preserve formatting. Use --dry-run to preview
changes without writing.

Examples:
  deets import backup.toml             # import into global
  deets import other.toml --local      # import into local
  deets import other.toml --dry-run    # preview changes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importPath := args[0]

		importDB, err := store.LoadFile(importPath)
		if err != nil {
			return fmt.Errorf("loading import file: %w", err)
		}

		if flagImportDryRun {
			return importDryRun(importDB)
		}

		targetPath, err := targetFile()
		if err != nil {
			return err
		}

		count := 0
		for _, cat := range importDB.Categories {
			for _, f := range cat.Fields {
				if model.IsDescKey(f.Key) {
					continue
				}
				val := model.FormatValueTOML(f.Value)
				if err := store.SetValue(targetPath, cat.Name, f.Key, val); err != nil {
					return fmt.Errorf("setting %s.%s: %w", cat.Name, f.Key, err)
				}
				count++
			}
		}

		if !flagQuiet {
			fmt.Printf("Imported %d fields into %s\n", count, targetPath)
		}
		return nil
	},
}

func importDryRun(importDB *model.DB) error {
	// Load existing DB to compare; tolerate missing file but not other errors.
	existingDB, err := loadDB()
	if err != nil && !strings.Contains(err.Error(), "no deets found") {
		return err
	}

	var entries []model.DiffEntry
	for _, cat := range importDB.Categories {
		for _, f := range cat.Fields {
			if model.IsDescKey(f.Key) {
				continue
			}
			path := cat.Name + "." + f.Key
			newVal := model.FormatValue(f.Value)

			entry := model.DiffEntry{
				Path:     path,
				LocalVal: newVal,
			}

			if existingDB != nil {
				existing, ok := existingDB.GetField(path)
				if ok {
					oldVal := model.FormatValue(existing.Value)
					if oldVal == newVal {
						continue // no change
					}
					entry.Status = "change"
					entry.GlobalVal = oldVal
				} else {
					entry.Status = "add"
				}
			} else {
				entry.Status = "add"
			}

			entries = append(entries, entry)
		}
	}

	if len(entries) == 0 {
		if !flagQuiet {
			fmt.Println("No changes to apply.")
		}
		return nil
	}

	switch resolveFormat() {
	case "json":
		out, err := model.FormatDiffJSON(entries)
		if err != nil {
			return err
		}
		fmt.Println(out)
	default:
		fmt.Print(model.FormatDiffTable(entries))
	}
	return nil
}
