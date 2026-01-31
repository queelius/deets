package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/queelius/deets/internal/config"
	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
)

// loadDB loads the merged metadata database (global + optional local).
func loadDB() (*model.DB, error) {
	globalPath := config.GlobalFile()
	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no deets found; run 'deets init' first")
	}

	localPath := config.FindLocalFile()
	return store.Load(globalPath, localPath)
}

// targetFile returns the TOML file path to write to, based on --local flag.
func targetFile() (string, error) {
	if flagLocal {
		if err := config.EnsureLocalDir(); err != nil {
			return "", err
		}
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Join(cwd, config.DirName, config.FileName), nil
	}

	if err := config.EnsureGlobalDir(); err != nil {
		return "", err
	}
	return config.GlobalFile(), nil
}
