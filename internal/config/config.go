package config

import (
	"os"
	"path/filepath"
)

const (
	// DirName is the name of the deets directory.
	DirName = ".deets"

	// FileName is the name of the data file.
	FileName = "me.toml"
)

// Paths holds the resolved paths for global and local deets directories.
type Paths struct {
	GlobalDir  string // path to ~/.deets/
	GlobalFile string // path to ~/.deets/me.toml
	LocalDir   string // path to local .deets/ (empty if not found)
	LocalFile  string // path to local .deets/me.toml (empty if not found)
	HasLocal   bool   // whether a local override exists
}

// GlobalDir returns the path to ~/.deets/.
func GlobalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, DirName)
}

// GlobalFile returns the path to ~/.deets/me.toml.
func GlobalFile() string {
	dir := GlobalDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, FileName)
}

// FindLocalDir walks up from the current working directory looking for a
// .deets/ directory. It stops at the user's home directory or the filesystem
// root. Returns an empty string if no .deets/ directory is found.
func FindLocalDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		// Stop before checking the home directory — ~/.deets/ is the global store.
		if dir == home {
			break
		}

		candidate := filepath.Join(dir, DirName)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root.
			break
		}
		dir = parent
	}

	return ""
}

// FindLocalFile returns the path to me.toml inside the local .deets/ directory
// found by FindLocalDir. Returns an empty string if no local directory is found
// or if me.toml does not exist inside it.
func FindLocalFile() string {
	localDir := FindLocalDir()
	if localDir == "" {
		return ""
	}

	file := filepath.Join(localDir, FileName)
	info, err := os.Stat(file)
	if err != nil || info.IsDir() {
		return ""
	}
	return file
}

// ResolvePaths resolves all deets paths and populates a Paths struct.
// Returns an error only if the home directory cannot be determined.
func ResolvePaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	p := Paths{
		GlobalDir:  filepath.Join(home, DirName),
		GlobalFile: filepath.Join(home, DirName, FileName),
	}

	p.LocalDir = FindLocalDir()
	if p.LocalDir != "" {
		p.LocalFile = FindLocalFile()
		p.HasLocal = p.LocalFile != ""
	}

	return p, nil
}

// EnsureGlobalDir creates ~/.deets/ if it does not already exist.
func EnsureGlobalDir() error {
	dir := GlobalDir()
	if dir == "" {
		_, err := os.UserHomeDir()
		return err
	}
	return os.MkdirAll(dir, 0755)
}

// EnsureLocalDir creates .deets/ in the current working directory if it does
// not already exist.
func EnsureLocalDir() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(cwd, DirName), 0755)
}
