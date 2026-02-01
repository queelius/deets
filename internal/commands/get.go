package commands

import (
	"fmt"
	"strings"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

var (
	flagGetDefault string
	flagGetDesc    bool
	flagGetExists  bool
)

func init() {
	getCmd.Flags().StringVar(&flagGetDefault, "default", "", "fallback value when no match found")
	getCmd.Flags().BoolVar(&flagGetDesc, "desc", false, "include field descriptions in output")
	getCmd.Flags().BoolVar(&flagGetExists, "exists", false, "check existence; exit 0 if found, 2 if not (no output)")
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
  deets get identity.na*           # glob within category
  deets get identity.name --desc   # include description
  deets get foo.bar --default x    # return "x" if not found
  deets get foo.bar --exists       # exit 0/2, no output`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		pattern := args[0]
		fields := db.Query(pattern)

		// --exists: pure existence check, no output
		if flagGetExists {
			if len(fields) == 0 {
				return &ExitError{Code: 2, Message: ""}
			}
			return nil
		}

		if len(fields) == 0 {
			// --default: return default value on no match
			if cmd.Flags().Changed("default") {
				fmt.Println(flagGetDefault)
				return nil
			}
			if strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[") {
				return &ExitError{Code: 2, Message: fmt.Sprintf("field not found: %s", pattern)}
			}
			return &ExitError{Code: 2, Message: fmt.Sprintf("no matches for: %s", pattern)}
		}

		// Use bare value only for exact field paths (no globs, no category-only)
		isExactField := strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[")
		format := resolveFormat()
		if len(fields) == 1 && isExactField && format == "table" {
			if flagGetDesc {
				fmt.Printf("%s\t%s\n", model.FormatValue(fields[0].Value), fields[0].Desc)
			} else {
				fmt.Println(model.FormatValue(fields[0].Value))
			}
			return nil
		}

		// Multiple results or explicit format
		switch format {
		case "json":
			var out string
			if flagGetDesc {
				out, err = model.FormatFieldsJSONWithDesc(fields)
			} else {
				out, err = model.FormatFieldsJSON(fields)
			}
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
			if flagGetDesc {
				fmt.Print(model.FormatTableWithDesc(fields))
			} else {
				fmt.Print(model.FormatTable(fields))
			}
		}
		return nil
	},
}
