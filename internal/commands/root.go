package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	flagFormat string
	flagLocal  bool
	flagQuiet  bool
)

// validFormats lists all recognized output format names.
var validFormats = map[string]bool{
	"table": true,
	"json":  true,
	"toml":  true,
	"yaml":  true,
	"env":   true,
}

var rootCmd = &cobra.Command{
	Use:           "deets",
	Short:         "Personal metadata CLI",
	Long:          "A self-describing, TOML-backed personal metadata store.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return validateFormat()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagFormat, "format", "", "output format: table, json, toml, yaml, env")
	rootCmd.PersistentFlags().BoolVar(&flagLocal, "local", false, "operate on local .deets/me.toml")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress informational messages")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// resolveFormat returns the effective output format for the current invocation.
// If --format was explicitly set, that value is returned. Otherwise, TTY
// detection drives the default: "table" on a terminal, "json" when piped.
func resolveFormat() string {
	if flagFormat != "" {
		return flagFormat
	}
	if isTTY() {
		return "table"
	}
	return "json"
}

// validateFormat checks that the --format flag (if given) is a known format.
func validateFormat() error {
	if flagFormat == "" {
		return nil
	}
	if !validFormats[flagFormat] {
		return fmt.Errorf("unknown format %q: expected table, json, toml, yaml, or env", flagFormat)
	}
	return nil
}

// isTTY reports whether stdout is connected to a terminal.
func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
