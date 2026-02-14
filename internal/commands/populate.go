package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	populateCmd.Flags().BoolVar(&flagPopulateGithub, "github", false, "harvest from GitHub API via gh CLI")
	populateCmd.Flags().BoolVar(&flagPopulateOrcid, "orcid", false, "harvest from ORCID public API")
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
  deets populate --github           # harvest from GitHub profile
  deets populate --orcid            # harvest from ORCID public API
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

// populateGithub harvests metadata from the GitHub API via `gh api user`.
func populateGithub() ([]populateEntry, error) {
	data, err := exec.Command("gh", "api", "user").Output()
	if err != nil {
		return nil, fmt.Errorf("running 'gh api user': %w (is gh CLI installed and authenticated?)", err)
	}
	return parseGitHubUser(data)
}

// parseGitHubUser parses the JSON from `gh api user` into populate entries.
func parseGitHubUser(data []byte) ([]populateEntry, error) {
	var user struct {
		Login   string `json:"login"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Bio     string `json:"bio"`
		Blog    string `json:"blog"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("parsing GitHub response: %w", err)
	}

	var entries []populateEntry

	if user.Login != "" {
		entries = append(entries, populateEntry{"profiles.github", "username", user.Login})
	}
	if user.Name != "" {
		entries = append(entries, populateEntry{"profiles.github", "name", user.Name})
		// Also suggest for identity.name if not empty
		entries = append(entries, populateEntry{"identity", "name", user.Name})
	}
	if user.Email != "" {
		entries = append(entries, populateEntry{"profiles.github", "email", user.Email})
		// Also suggest for contact.email
		entries = append(entries, populateEntry{"contact", "email", user.Email})
	}
	if user.HTMLURL != "" {
		entries = append(entries, populateEntry{"profiles.github", "url", user.HTMLURL})
	}
	if user.Bio != "" {
		entries = append(entries, populateEntry{"profiles.github", "bio", user.Bio})
	}
	if user.Blog != "" {
		entries = append(entries, populateEntry{"profiles.blog", "url", user.Blog})
	}

	return entries, nil
}

// populateOrcid harvests metadata from the ORCID public API.
// Requires academic.orcid to be set in the database first.
func populateOrcid() ([]populateEntry, error) {
	db, err := loadDB()
	if err != nil {
		return nil, fmt.Errorf("loading database: %w", err)
	}

	field, ok := db.GetField("academic.orcid")
	if !ok {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "orcid: academic.orcid not set; skipping (set it with: deets set academic.orcid \"0000-...\")")
		}
		return nil, nil
	}
	orcid := model.FormatValue(field.Value)

	url := fmt.Sprintf("https://pub.orcid.org/v3.0/%s/person", orcid)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching ORCID profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ORCID API returned %s for %s", resp.Status, orcid)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading ORCID response: %w", err)
	}

	return parseOrcidPerson(data, orcid)
}

// parseOrcidPerson parses the ORCID person API response into populate entries.
func parseOrcidPerson(data []byte, orcid string) ([]populateEntry, error) {
	var person struct {
		Name struct {
			GivenNames struct {
				Value string `json:"value"`
			} `json:"given-names"`
			FamilyName struct {
				Value string `json:"value"`
			} `json:"family-name"`
		} `json:"name"`
		Emails struct {
			Email []struct {
				Email string `json:"email"`
			} `json:"email"`
		} `json:"emails"`
	}
	if err := json.Unmarshal(data, &person); err != nil {
		return nil, fmt.Errorf("parsing ORCID response: %w", err)
	}

	var entries []populateEntry

	given := person.Name.GivenNames.Value
	family := person.Name.FamilyName.Value
	if given != "" || family != "" {
		fullName := strings.TrimSpace(given + " " + family)
		entries = append(entries, populateEntry{"profiles.orcid", "name", fullName})
	}

	entries = append(entries, populateEntry{"profiles.orcid", "id", orcid})
	entries = append(entries, populateEntry{"profiles.orcid", "url", "https://orcid.org/" + orcid})

	if len(person.Emails.Email) > 0 && person.Emails.Email[0].Email != "" {
		entries = append(entries, populateEntry{"profiles.orcid", "email", person.Emails.Email[0].Email})
	}

	return entries, nil
}
