package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// executeCommand runs a cobra command with the given args and captures output.
// It captures actual os.Stdout and os.Stderr since commands use fmt.Print*
// which writes to os.Stdout directly, not to cmd.OutOrStdout().
func executeCommand(args ...string) (stdout, stderr string, err error) {
	// Capture stdout
	origStdout := os.Stdout
	rOut, wOut, _ := os.Pipe()
	os.Stdout = wOut

	// Capture stderr
	origStderr := os.Stderr
	rErr, wErr, _ := os.Pipe()
	os.Stderr = wErr

	rootCmd.SetArgs(args)
	err = rootCmd.Execute()

	// Close write ends and read the captured output
	wOut.Close()
	wErr.Close()
	os.Stdout = origStdout
	os.Stderr = origStderr

	var stdoutBuf, stderrBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(&stdoutBuf, rOut)
	}()
	go func() {
		defer wg.Done()
		io.Copy(&stderrBuf, rErr)
	}()
	wg.Wait()

	return stdoutBuf.String(), stderrBuf.String(), err
}

// setupTestEnv creates an isolated home directory.
// It sets HOME to a temp dir and changes CWD into it so that
// config.FindLocalDir() won't walk up to the real home directory.
// Returns the home directory path.
func setupTestEnv(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)

	// Change CWD into the temp home so FindLocalDir() doesn't
	// walk into the real user's ~/.deets/.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getting cwd: %v", err)
	}
	if err := os.Chdir(home); err != nil {
		t.Fatalf("chdir to temp home: %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Reset global flags to defaults before each test.
	flagFormat = ""
	flagLocal = false
	flagQuiet = false
	flagGetDefault = ""
	flagGetDesc = false
	flagGetExists = false
	flagImportDryRun = false

	return home
}

// setupTestDB creates an isolated test environment and initializes a
// deets database with sample data. Returns the home directory path.
func setupTestDB(t *testing.T) string {
	t.Helper()
	home := setupTestEnv(t)

	deetsDir := filepath.Join(home, ".deets")
	if err := os.MkdirAll(deetsDir, 0755); err != nil {
		t.Fatalf("creating deets dir: %v", err)
	}

	toml := `[identity]
name = "Alexander Towell"
name_desc = "Full legal name"
aka = ["Alex Towell", "Alex T"]

[contact]
email = "alex@example.com"
email_desc = "Primary email"

[web]
github = "queelius"
github_desc = "GitHub username"
website = "https://example.com"

[academic]
orcid = "0000-0001-2345-6789"
orcid_desc = "ORCID persistent digital identifier"
gpa = 3.95
topics = ["statistics", "machine learning"]
`
	if err := os.WriteFile(filepath.Join(deetsDir, "me.toml"), []byte(toml), 0644); err != nil {
		t.Fatalf("writing test TOML: %v", err)
	}

	return home
}
