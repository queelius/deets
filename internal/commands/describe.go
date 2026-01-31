package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(describeCmd)
}

var describeCmd = &cobra.Command{
	Use:   "describe [path] [description]",
	Short: "Show or set field descriptions",
	Long: `Show or set field descriptions.

Examples:
  deets describe                          # all descriptions
  deets describe identity                 # descriptions in category
  deets describe academic.orcid           # single field description
  deets describe web.mastodon "Mastodon handle"  # set a description`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Setting a description
		if len(args) == 2 {
			return setDescription(args[0], args[1])
		}

		db, err := loadDB()
		if err != nil {
			return err
		}

		var fields []model.Field

		switch len(args) {
		case 0:
			// All descriptions
			fields = db.AllDescriptions()
		case 1:
			path := args[0]
			if strings.Contains(path, ".") {
				// Single field description
				desc := db.DescribeField(path)
				if desc == "" {
					fmt.Fprintf(os.Stderr, "no description for: %s\n", path)
					os.Exit(2)
				}
				fmt.Println(desc)
				return nil
			}
			// Category descriptions
			fields = db.DescribeCategory(path)
		}

		if len(fields) == 0 {
			fmt.Fprintln(os.Stderr, "no descriptions found")
			os.Exit(2)
		}

		if flagJSON || !isTTY() {
			out, err := model.FormatDescJSON(fields)
			if err != nil {
				return err
			}
			fmt.Println(out)
		} else {
			fmt.Print(model.FormatDescTable(fields))
		}
		return nil
	},
}

func setDescription(path, desc string) error {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid path %q: expected category.key", path)
	}

	filePath, err := targetFile()
	if err != nil {
		return err
	}

	return store.SetValue(filePath, parts[0], parts[1]+"_desc", desc)
}
