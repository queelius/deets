package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/queelius/deets/internal/config"
	"github.com/queelius/deets/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a new deets metadata file",
	Long:  "Create ~/.deets/me.toml from a template, or .deets/me.toml with --local.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagLocal {
			return initLocal()
		}
		return initGlobal()
	},
}

func initGlobal() error {
	if err := config.EnsureGlobalDir(); err != nil {
		return fmt.Errorf("creating global directory: %w", err)
	}

	path := config.GlobalFile()
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	if err := os.WriteFile(path, []byte(store.DefaultTemplate), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	if !flagQuiet {
		fmt.Printf("Created %s\n", path)
		fmt.Println("Edit it to add your personal details.")
	}
	return nil
}

func initLocal() error {
	if err := config.EnsureLocalDir(); err != nil {
		return fmt.Errorf("creating local directory: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	path := filepath.Join(cwd, config.DirName, config.FileName)

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists", path)
	}

	if err := os.WriteFile(path, []byte(store.LocalTemplate), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	if !flagQuiet {
		fmt.Printf("Created %s\n", path)
	}
	return nil
}
