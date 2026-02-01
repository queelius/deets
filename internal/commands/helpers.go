package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/queelius/deets/internal/config"
	"github.com/queelius/deets/internal/model"
	"github.com/queelius/deets/internal/store"
)

// ExitError represents a command failure with a specific exit code.
// Commands return this instead of calling os.Exit() directly, so the
// error can be handled (and tested) at the top level in main.go.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

// parsePath splits a "category.key" path and validates both parts are non-empty.
func parsePath(path string) (category, key string, err error) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid path %q: expected category.key", path)
	}
	return parts[0], parts[1], nil
}

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
