package commands

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillContent string

var flagClaudeGlobal bool

func init() {
	claudeInstallCmd.Flags().BoolVar(&flagClaudeGlobal, "global", false, "install to ~/.claude/skills/ (default)")
	claudeUninstallCmd.Flags().BoolVar(&flagClaudeGlobal, "global", false, "uninstall from ~/.claude/skills/")
	claudeCmd.AddCommand(claudeInstallCmd)
	claudeCmd.AddCommand(claudeUninstallCmd)
	rootCmd.AddCommand(claudeCmd)
}

var claudeCmd = &cobra.Command{
	Use:   "claude",
	Short: "Manage Claude Code skill integration",
}

var claudeInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the deets skill for Claude Code",
	Long: `Install the deets skill file so Claude Code knows how to use deets.

By default installs to ~/.claude/skills/deets.md (global).
Use --local to install to .claude/skills/deets.md in the current project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := skillPath()
		if err != nil {
			return err
		}

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}

		if err := os.WriteFile(path, []byte(skillContent), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		if !flagQuiet {
			fmt.Printf("Installed deets skill to %s\n", path)
		}
		return nil
	},
}

var claudeUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the deets skill for Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := skillPath()
		if err != nil {
			return err
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			if !flagQuiet {
				fmt.Printf("No skill file at %s\n", path)
			}
			return nil
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing %s: %w", path, err)
		}

		if !flagQuiet {
			fmt.Printf("Removed deets skill from %s\n", path)
		}
		return nil
	},
}

func skillPath() (string, error) {
	if flagLocal && !flagClaudeGlobal {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, ".claude", "skills", "deets.md"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "skills", "deets.md"), nil
}
