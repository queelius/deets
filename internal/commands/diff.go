package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/config"
	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show differences between global and local files",
	Long: `Compare fields in the local .deets/me.toml against the global
~/.deets/me.toml. Shows overrides and local-only fields.

Examples:
  deets diff                  # table output
  deets diff --format json    # JSON output`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := config.FindLocalFile()
		if localPath == "" {
			return fmt.Errorf("no local .deets/me.toml found")
		}

		globalPath := config.GlobalFile()
		globalDB, err := store.LoadFile(globalPath)
		if err != nil {
			return fmt.Errorf("loading global file: %w", err)
		}

		localDB, err := store.LoadFile(localPath)
		if err != nil {
			return fmt.Errorf("loading local file: %w", err)
		}

		entries := computeDiff(globalDB, localDB)

		if len(entries) == 0 {
			if !flagQuiet {
				fmt.Println("No differences.")
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
		default: // table
			fmt.Print(model.FormatDiffTable(entries))
		}
		return nil
	},
}

// computeDiff compares global and local DBs and returns diff entries.
func computeDiff(globalDB, localDB *model.DB) []model.DiffEntry {
	var entries []model.DiffEntry

	for _, cat := range localDB.Categories {
		for _, f := range cat.Fields {
			if model.IsDescKey(f.Key) {
				continue
			}
			path := cat.Name + "." + f.Key
			localVal := model.FormatValue(f.Value)

			globalField, found := globalDB.GetField(path)
			if found {
				globalVal := model.FormatValue(globalField.Value)
				if globalVal != localVal {
					entries = append(entries, model.DiffEntry{
						Path:      path,
						Status:    "override",
						GlobalVal: globalVal,
						LocalVal:  localVal,
					})
				}
			} else {
				entries = append(entries, model.DiffEntry{
					Path:     path,
					Status:   "local-only",
					LocalVal: localVal,
				})
			}
		}
	}

	return entries
}
