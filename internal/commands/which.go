package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/queelius/deets/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(whichCmd)
}

var whichCmd = &cobra.Command{
	Use:   "which",
	Short: "Show resolved file paths and merge status",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := config.ResolvePaths()
		if err != nil {
			return err
		}

		if flagJSON {
			data, err := json.MarshalIndent(map[string]interface{}{
				"global_dir":  paths.GlobalDir,
				"global_file": paths.GlobalFile,
				"local_dir":   paths.LocalDir,
				"local_file":  paths.LocalFile,
				"has_local":   paths.HasLocal,
				"global_exists": fileExists(paths.GlobalFile),
			}, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		fmt.Printf("Global: %s", paths.GlobalFile)
		if fileExists(paths.GlobalFile) {
			fmt.Println(" (exists)")
		} else {
			fmt.Println(" (not found)")
		}

		if paths.HasLocal {
			fmt.Printf("Local:  %s (active override)\n", paths.LocalFile)
		} else if paths.LocalDir != "" {
			fmt.Printf("Local:  %s (dir exists, no me.toml)\n", paths.LocalDir)
		} else {
			fmt.Println("Local:  none")
		}

		return nil
	},
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
