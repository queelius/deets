package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export metadata in various formats",
	Long: `Export all metadata in a specific format.

Examples:
  deets export --format json    # JSON (default)
  deets export --format env     # DEETS_IDENTITY_NAME="..." format
  deets export --format toml    # raw merged TOML
  deets export --format yaml    # YAML`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		// Export defaults to JSON when resolveFormat() returns "table",
		// since export is inherently structured output.
		format := resolveFormat()
		if format == "table" {
			format = "json"
		}

		switch format {
		case "env":
			fmt.Print(model.FormatEnv(db))
		case "toml":
			fmt.Print(model.FormatTOML(db))
		case "yaml":
			fmt.Print(model.FormatYAML(db))
		default: // json
			out, err := model.FormatJSON(db)
			if err != nil {
				return err
			}
			fmt.Println(out)
		}
		return nil
	},
}
