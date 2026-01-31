package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(getCmd)
}

var getCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get a metadata value",
	Long: `Get a metadata value by path. Supports glob patterns.

Examples:
  deets get identity.name          # single value
  deets get academic               # all fields in category
  deets get *.orcid                # find key across categories
  deets get identity.na*           # glob within category`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		pattern := args[0]
		fields := db.Query(pattern)

		if len(fields) == 0 {
			// Try exact field lookup for better error message
			if strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[") {
				fmt.Fprintf(os.Stderr, "field not found: %s\n", pattern)
			} else {
				fmt.Fprintf(os.Stderr, "no matches for: %s\n", pattern)
			}
			os.Exit(2)
		}

		// Use bare value only for exact field paths (no globs, no category-only)
		isExactField := strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[")
		if len(fields) == 1 && isExactField && !flagJSON {
			fmt.Println(model.FormatValue(fields[0].Value))
			return nil
		}

		// Multiple results
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
