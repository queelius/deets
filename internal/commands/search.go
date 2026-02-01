package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search keys, values, and descriptions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		fields := db.Search(args[0])
		if len(fields) == 0 {
			return &ExitError{Code: 2, Message: fmt.Sprintf("no matches for: %s", args[0])}
		}

		switch resolveFormat() {
		case "json":
			out, err := model.FormatFieldsJSON(fields)
			if err != nil {
				return err
			}
			fmt.Println(out)
		case "toml":
			db := model.FieldsToDB(fields)
			fmt.Print(model.FormatTOML(db))
		case "yaml":
			db := model.FieldsToDB(fields)
			fmt.Print(model.FormatYAML(db))
		case "env":
			db := model.FieldsToDB(fields)
			fmt.Print(model.FormatEnv(db))
		default: // table
			fmt.Print(model.FormatTable(fields))
		}
		return nil
	},
}
