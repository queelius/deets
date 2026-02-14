package commands

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

var (
	flagPopulateGit    bool
	flagPopulateGithub bool
	flagPopulateOrcid  bool
	flagPopulateAll    bool
	flagPopulateDryRun bool
	flagPopulateYes    bool
)

func init() {
	populateCmd.Flags().BoolVar(&flagPopulateGit, "git", false, "harvest from local git config")
	populateCmd.Flags().BoolVar(&flagPopulateGithub, "github", false, "harvest from GitHub API (not yet implemented)")
	populateCmd.Flags().BoolVar(&flagPopulateOrcid, "orcid", false, "harvest from ORCID API (not yet implemented)")
	populateCmd.Flags().BoolVar(&flagPopulateAll, "all", false, "enable all available sources")
	populateCmd.Flags().BoolVar(&flagPopulateDryRun, "dry-run", false, "preview changes without writing")
	populateCmd.Flags().BoolVar(&flagPopulateYes, "yes", false, "skip confirmation prompt")
	rootCmd.AddCommand(populateCmd)
}

// populateEntry represents a single field to be populated.
type populateEntry struct {
	category string
	key      string
	value    string
}

var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Auto-populate metadata from external sources",
	Long: `Auto-populate metadata fields from external sources.

At least one source flag must be specified. Use --dry-run to preview
proposed changes, or --yes to skip confirmation.

Examples:
  deets populate --git              # harvest from local git config
  deets populate --git --dry-run    # preview only
  deets populate --git --yes        # skip confirmation
  deets populate --all              # all available sources`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagPopulateAll {
			flagPopulateGit = true
			flagPopulateGithub = true
			flagPopulateOrcid = true
		}

		if !flagPopulateGit && !flagPopulateGithub && !flagPopulateOrcid {
			return fmt.Errorf("no source specified; use --git, --github, --orcid, or --all")
		}

		var entries []populateEntry

		if flagPopulateGit {
			gitEntries, err := populateGit()
			if err != nil {
				return fmt.Errorf("git source: %w", err)
			}
			entries = append(entries, gitEntries...)
		}

		if flagPopulateGithub {
			githubEntries, err := populateGithub()
			if err != nil {
				return fmt.Errorf("github source: %w", err)
			}
			entries = append(entries, githubEntries...)
		}

		if flagPopulateOrcid {
			orcidEntries, err := populateOrcid()
			if err != nil {
				return fmt.Errorf("orcid source: %w", err)
			}
			entries = append(entries, orcidEntries...)
		}

		if len(entries) == 0 {
			if !flagQuiet {
				fmt.Println("No data found from sources.")
			}
			return nil
		}

		// Load existing DB to diff against.
		existingDB, _ := loadDB()

		// Compute proposed changes.
		type change struct {
			entry  populateEntry
			status string // "add" or "change"
			oldVal string // previous value for "change"
		}
		var changes []change

		for _, e := range entries {
			path := e.category + "." + e.key
			if existingDB != nil {
				field, ok := existingDB.GetField(path)
				if ok {
					oldVal := model.FormatValue(field.Value)
					if oldVal == e.value {
						continue // unchanged, skip
					}
					changes = append(changes, change{entry: e, status: "change", oldVal: oldVal})
					continue
				}
			}
			changes = append(changes, change{entry: e, status: "add"})
		}

		if len(changes) == 0 {
			if !flagQuiet {
				fmt.Println("No changes to apply.")
			}
			return nil
		}

		// Display proposed changes.
		for _, c := range changes {
			path := c.entry.category + "." + c.entry.key
			if c.status == "add" {
				fmt.Printf("+ %s = %q\n", path, c.entry.value)
			} else {
				fmt.Printf("~ %s = %q (was: %q)\n", path, c.entry.value, c.oldVal)
			}
		}

		if flagPopulateDryRun {
			return nil
		}

		// Confirm unless --yes is set.
		if !flagPopulateYes {
			fmt.Print("Apply changes? [y/N] ")
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("reading confirmation: %w", err)
			}
			answer = strings.TrimSpace(strings.ToLower(answer))
			if answer != "y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Write changes.
		filePath, err := targetFile()
		if err != nil {
			return err
		}

		for _, c := range changes {
			if err := store.SetValue(filePath, c.entry.category, c.entry.key, c.entry.value); err != nil {
				return fmt.Errorf("setting %s.%s: %w", c.entry.category, c.entry.key, err)
			}
		}

		if !flagQuiet {
			fmt.Printf("Applied %d changes.\n", len(changes))
		}
		return nil
	},
}

// populateGit harvests user.name and user.email from git config.
func populateGit() ([]populateEntry, error) {
	var entries []populateEntry
	name, err := exec.Command("git", "config", "user.name").Output()
	if err == nil && len(strings.TrimSpace(string(name))) > 0 {
		entries = append(entries, populateEntry{"identity", "name", strings.TrimSpace(string(name))})
	}
	email, err := exec.Command("git", "config", "user.email").Output()
	if err == nil && len(strings.TrimSpace(string(email))) > 0 {
		entries = append(entries, populateEntry{"contact", "email", strings.TrimSpace(string(email))})
	}
	return entries, nil
}

// populateGithub harvests metadata from the GitHub API.
// Not yet implemented — will be added in a future task.
func populateGithub() ([]populateEntry, error) {
	return nil, nil
}

// populateOrcid harvests metadata from the ORCID API.
// Not yet implemented — will be added in a future task.
func populateOrcid() ([]populateEntry, error) {
	return nil, nil
}
