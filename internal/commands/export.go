package commands

import (
	"fmt"

	"github.com/queelius/deets/internal/model"
	"github.com/spf13/cobra"
)

var (
	flagExportEnv  bool
	flagExportTOML bool
	flagExportYAML bool
)

func init() {
	exportCmd.Flags().BoolVar(&flagExportEnv, "env", false, "export as environment variables")
	exportCmd.Flags().BoolVar(&flagExportTOML, "toml", false, "export as TOML")
	exportCmd.Flags().BoolVar(&flagExportYAML, "yaml", false, "export as YAML")
	rootCmd.AddCommand(exportCmd)
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export metadata in various formats",
	Long: `Export all metadata in a specific format.

Examples:
  deets export --json    # JSON
  deets export --env     # DEETS_IDENTITY_NAME="..." format
  deets export --toml    # raw merged TOML
  deets export --yaml    # YAML`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := loadDB()
		if err != nil {
			return err
		}

		switch {
		case flagExportEnv:
			fmt.Print(model.FormatEnv(db))
		case flagExportTOML:
			fmt.Print(model.FormatTOML(db))
		case flagExportYAML:
			fmt.Print(model.FormatYAML(db))
		case flagJSON:
			out, err := model.FormatJSON(db)
			if err != nil {
				return err
			}
			fmt.Println(out)
		default:
			// Default to JSON
			out, err := model.FormatJSON(db)
			if err != nil {
				return err
			}
			fmt.Println(out)
		}
		return nil
	},
}
