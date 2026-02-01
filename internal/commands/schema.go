package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(schemaCmd)
}

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Show field types and metadata",
	Long: `Display the schema of all fields: category, key, inferred type,
description, and example value.

Examples:
  deets schema                  # table output
  deets schema --format json    # JSON array`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		entries := model.BuildSchema(db)

		switch resolveFormat() {
		case "json":
			out, err := model.FormatSchemaJSON(entries)
			if err != nil {
				return err
			}
			fmt.Println(out)
		default: // table
			fmt.Print(model.FormatSchemaTable(entries))
		}
		return nil
	},
}
