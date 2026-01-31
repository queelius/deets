package commands

import (
	"fmt"
	"strings"

	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set <category.key> <value>",
	Short: "Set a metadata value",
	Long: `Set a metadata value. Creates the category if it doesn't exist.

Examples:
  deets set identity.name "Alexander Towell"
  deets set cooking.fav "lasagna"          # creates [cooking]
  deets set identity.aka '["Alex Towell"]' # array value`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		value := args[1]

		parts := strings.SplitN(path, ".", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid path %q: expected category.key", path)
		}

		filePath, err := targetFile()
		if err != nil {
			return err
		}

		return store.SetValue(filePath, parts[0], parts[1], value)
	},
}
