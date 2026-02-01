package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(keysCmd)
}

var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "List all field paths",
	Long: `List every field path in the database, one per line.

Examples:
  deets keys                  # one per line
  deets keys --format json    # JSON array`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		fields := db.AllFields()
		paths := make([]string, 0, len(fields))
		for _, f := range fields {
			paths = append(paths, f.Category+"."+f.Key)
		}

		switch resolveFormat() {
		case "json":
			data, err := json.MarshalIndent(paths, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		default: // table
			for _, p := range paths {
				fmt.Println(p)
			}
		}
		return nil
	},
}
