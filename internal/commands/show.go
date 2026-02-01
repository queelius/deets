package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show [category]",
	Short: "Display metadata",
	Long: `Display all metadata, or a single category.

Examples:
  deets show                    # all categories as table
  deets show identity           # single category
  deets show --format json      # full JSON dump
  deets show --format toml      # raw merged TOML
  deets show --format yaml      # YAML output`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		format := resolveFormat()

		// Single category
		if len(args) == 1 {
			cat, ok := db.GetCategory(args[0])
			if !ok {
				return fmt.Errorf("category not found: %s", args[0])
			}

			switch format {
			case "json":
				out, err := model.FormatCategoryJSON(cat)
				if err != nil {
					return err
				}
				fmt.Println(out)
			case "toml":
				catDB := &model.DB{Categories: []model.Category{cat}}
				fmt.Print(model.FormatTOML(catDB))
			case "yaml":
				catDB := &model.DB{Categories: []model.Category{cat}}
				fmt.Print(model.FormatYAML(catDB))
			case "env":
				catDB := &model.DB{Categories: []model.Category{cat}}
				fmt.Print(model.FormatEnv(catDB))
			default: // table
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
		switch format {
		case "json":
			out, err := model.FormatJSON(db)
			if err != nil {
				return err
			}
			fmt.Println(out)
		case "toml":
			fmt.Print(model.FormatTOML(db))
		case "yaml":
			fmt.Print(model.FormatYAML(db))
		case "env":
			fmt.Print(model.FormatEnv(db))
		default: // table
			fmt.Print(model.FormatTable(db.AllFields()))
		}
		return nil
	},
}
