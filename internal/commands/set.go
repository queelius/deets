package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(setCmd)
}

var setCmd = &cobra.Command{
	Use:   "set <category.key> [value]",
	Short: "Set a metadata value",
	Long: `Set a metadata value. Creates the category if it doesn't exist.

The value can be provided as a second argument, piped via stdin, or with
"-" as the value to read from stdin explicitly.

Examples:
  deets set identity.name "Alexander Towell"
  deets set cooking.fav "lasagna"          # creates [cooking]
  deets set identity.aka '["Alex Towell"]' # array value
  echo "piped" | deets set identity.name   # value from stdin
  cat file.txt | deets set identity.bio -  # explicit stdin`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		cat, key, err := parsePath(path)
		if err != nil {
			return err
		}

		var value string

		switch {
		case len(args) == 2 && args[1] != "-":
			value = args[1]
		case len(args) == 2 && args[1] == "-":
			// Explicit stdin sentinel
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			value = strings.TrimRight(string(data), "\n")
		case len(args) == 1:
			if isTTY() {
				return fmt.Errorf("value argument required (or pipe from stdin)")
			}
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}
			value = strings.TrimRight(string(data), "\n")
		}

		filePath, err := targetFile()
		if err != nil {
			return err
		}

		return store.SetValue(filePath, cat, key, value)
	},
}
