package commands

import (
	"fmt"
	"os"

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
			fmt.Fprintf(os.Stderr, "no matches for: %s\n", args[0])
			os.Exit(2)
		}

		if flagJSON || !isTTY() {
			out, err := model.FormatFieldsJSON(fields)
			if err != nil {
				return err
			}
			fmt.Println(out)
		} else {
			fmt.Print(model.FormatTable(fields))
		}
		return nil
	},
}
