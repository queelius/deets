package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(categoriesCmd)
}

var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List category names",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		names := db.CategoryNames()

		if flagJSON || !isTTY() {
			data, err := json.MarshalIndent(names, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			for _, name := range names {
				fmt.Println(name)
			}
		}
		return nil
	},
}
