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
  deets rm contact.phone     # remove a field
  deets rm cooking           # remove entire category`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		filePath, err := targetFile()
		if err != nil {
			return err
		}

		if strings.Contains(path, ".") {
			parts := strings.SplitN(path, ".", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return fmt.Errorf("invalid path %q: expected category.key", path)
			}
			return store.RemoveValue(filePath, parts[0], parts[1])
		}

		return store.RemoveCategory(filePath, path)
	},
}
