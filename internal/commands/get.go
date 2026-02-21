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
	Use:   "get <path> [path...]",
	Short: "Get a metadata value",
	Long: `Get one or more metadata values by path. Supports glob patterns.

Examples:
  deets get identity.name          # single value (bare output)
  deets get academic               # all fields in category
  deets get *.orcid                # find key across categories
  deets get identity.na*           # glob within category
  deets get identity.name --desc   # include description
  deets get foo.bar --default x    # return "x" if not found
  deets get foo.bar --exists       # exit 0/2, no output

  # Multiple paths in one call
  deets get identity.name contact.email academic.orcid`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		multiPath := len(args) > 1

		// Collect fields from all paths
		var fields []model.Field
		for _, pattern := range args {
			matched := db.Query(pattern)

			if len(matched) == 0 {
				// --exists: any missing path → exit 2
				if flagGetExists {
					return &ExitError{Code: 2, Message: ""}
				}

				// --default: synthesize a field with the default value
				if cmd.Flags().Changed("default") {
					cat, key := "", pattern
					if dot := strings.LastIndex(pattern, "."); dot != -1 {
						cat, key = pattern[:dot], pattern[dot+1:]
					}
					fields = append(fields, model.Field{
						Category: cat,
						Key:      key,
						Value:    flagGetDefault,
					})
					continue
				}

				// No match, no default → error
				if strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[") {
					return &ExitError{Code: 2, Message: fmt.Sprintf("field not found: %s", pattern)}
				}
				return &ExitError{Code: 2, Message: fmt.Sprintf("no matches for: %s", pattern)}
			}

			fields = append(fields, matched...)
		}

		// --exists with all paths found: success, no output
		if flagGetExists {
			return nil
		}

		// Single-path bare value: exact field path with one result
		if !multiPath {
			pattern := args[0]
			isExactField := strings.Contains(pattern, ".") && !strings.ContainsAny(pattern, "*?[")
			format := resolveFormat()
			if len(fields) == 1 && isExactField && (format == "table" || format == "json") {
				if flagGetDesc {
					fmt.Printf("%s\t%s\n", model.FormatValue(fields[0].Value), fields[0].Desc)
				} else {
					fmt.Println(model.FormatValue(fields[0].Value))
				}
				return nil
			}
		}

		// Structured output
		format := resolveFormat()
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
