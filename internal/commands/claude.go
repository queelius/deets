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

By default installs to ~/.claude/skills/deets/SKILL.md (global).
Use --local to install to .claude/skills/deets/SKILL.md in the current project.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := skillPath()
		if err != nil {
			return err
		}

		// Clean up old flat-file format if it exists
		if oldPath, err := oldSkillPath(); err == nil {
			if _, err := os.Stat(oldPath); err == nil {
				os.Remove(oldPath)
			}
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

		// Clean up old flat-file format if it exists
		if oldPath, err := oldSkillPath(); err == nil {
			if _, err := os.Stat(oldPath); err == nil {
				os.Remove(oldPath)
				if !flagQuiet {
					fmt.Printf("Removed old skill file %s\n", oldPath)
				}
			}
		}

		// Remove the skill directory (e.g., ~/.claude/skills/deets/)
		skillDir := filepath.Dir(path)
		if _, err := os.Stat(skillDir); os.IsNotExist(err) {
			if !flagQuiet {
				fmt.Printf("No skill directory at %s\n", skillDir)
			}
			return nil
		}

		if err := os.RemoveAll(skillDir); err != nil {
			return fmt.Errorf("removing %s: %w", skillDir, err)
		}

		if !flagQuiet {
			fmt.Printf("Removed deets skill from %s\n", skillDir)
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
		return filepath.Join(cwd, ".claude", "skills", "deets", "SKILL.md"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "skills", "deets", "SKILL.md"), nil
}

// oldSkillPath returns the legacy flat-file path for transition cleanup.
func oldSkillPath() (string, error) {
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
