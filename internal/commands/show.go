package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

var flagTOML bool

func init() {
	showCmd.Flags().BoolVar(&flagTOML, "toml", false, "output as raw TOML")
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show [category]",
	Short: "Display metadata",
	Long: `Display all metadata, or a single category.

Examples:
  deets show              # all categories as table
  deets show identity     # single category
  deets show --json       # full JSON dump
  deets show --toml       # raw merged TOML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		// --toml flag
		if flagTOML {
			fmt.Print(model.FormatTOML(db))
			return nil
		}

		// Single category
		if len(args) == 1 {
			cat, ok := db.GetCategory(args[0])
			if !ok {
				return fmt.Errorf("category not found: %s", args[0])
			}

			if flagJSON || !isTTY() {
				out, err := model.FormatCategoryJSON(cat)
				if err != nil {
					return err
				}
				fmt.Println(out)
			} else {
				fields := make([]model.Field, 0, len(cat.Fields))
				for _, f := range cat.Fields {
					if !model.IsDescKey(f.Key) {
						fields = append(fields, f)
					}
				}
				fmt.Print(model.FormatTable(fields))
			}
			return nil
		}

		// All categories
		if flagJSON || !isTTY() {
			out, err := model.FormatJSON(db)
			if err != nil {
				return err
			}
			fmt.Println(out)
		} else {
			fmt.Print(model.FormatTable(db.AllFields()))
		}
		return nil
	},
}
