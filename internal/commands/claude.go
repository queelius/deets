package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

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

		fmt.Printf("Installed deets skill to %s\n", path)
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
			fmt.Printf("No skill file at %s\n", path)
			return nil
		}

		if err := os.Remove(path); err != nil {
			return fmt.Errorf("removing %s: %w", path, err)
		}

		fmt.Printf("Removed deets skill from %s\n", path)
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

const skillContent = `---
name: deets
description: >
  Use when you need personal metadata about the user — name, email, ORCID,
  GitHub username, affiliations, or any other personal details. Also use when
  populating author fields, git identity, paper metadata, profile info, or
  personalized content.
---

# deets — Personal Metadata CLI

A TOML-backed personal metadata store. Query it for user identity and profile data.

## Quick Reference

` + "```" + `bash
# Single value (great for scripts and $(...) substitution)
deets get identity.name
deets get web.github
deets get contact.email

# Category (all fields)
deets get academic

# Cross-category search
deets get *.orcid

# Structured output
deets show --json         # full JSON dump
deets show identity       # single category table

# Search across everything
deets search "towell"

# Understand field meanings
deets describe academic.orcid

# Check configuration
deets which --json        # paths and merge status

# Export for scripts
deets export --env        # DEETS_IDENTITY_NAME="..." format
deets export --json       # full JSON
` + "```" + `

## When to Use

- **Author fields**: ` + "`" + `deets get identity.name` + "`" + `, ` + "`" + `deets get contact.email` + "`" + `
- **Git identity**: ` + "`" + `deets get identity.name` + "`" + `, ` + "`" + `deets get contact.email` + "`" + `
- **Academic papers**: ` + "`" + `deets get academic.orcid` + "`" + `, ` + "`" + `deets get academic.institution` + "`" + `
- **Profile/bio**: ` + "`" + `deets show --json` + "`" + ` for bulk data
- **Social links**: ` + "`" + `deets get web.github` + "`" + `, ` + "`" + `deets get web.blog` + "`" + `

## Output Conventions

- Single ` + "`" + `get` + "`" + `: bare value, no decoration (pipe-friendly)
- Multiple matches: table on TTY, JSON when piped
- ` + "`" + `--json` + "`" + ` flag forces JSON on any read command
- Exit code 2 = key not found
`
