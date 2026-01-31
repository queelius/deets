package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/queelius/deets/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(editCmd)
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open metadata file in $EDITOR",
	Long:  "Open ~/.deets/me.toml in $EDITOR, or .deets/me.toml with --local.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var path string
		if flagLocal {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			path = filepath.Join(cwd, config.DirName, config.FileName)
		} else {
			path = config.GlobalFile()
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("%s does not exist; run 'deets init' first", path)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi"
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}
