package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	flagJSON  bool
	flagLocal bool
)

var rootCmd = &cobra.Command{
	Use:   "deets",
	Short: "Personal metadata CLI",
	Long:  "A self-describing, TOML-backed personal metadata store.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagLocal, "local", false, "operate on local .deets/me.toml")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// isTTY reports whether stdout is connected to a terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
