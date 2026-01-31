package config

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

func TestConstants(t *testing.T) {
	if DirName != ".deets" {
		t.Errorf("DirName = %q, want %q", DirName, ".deets")
	}
	if FileName != "me.toml" {
		t.Errorf("FileName = %q, want %q", FileName, "me.toml")
	}
}

// ---------------------------------------------------------------------------
// GlobalDir
// ---------------------------------------------------------------------------

func TestGlobalDir_NonEmpty(t *testing.T) {
	dir := GlobalDir()
	if dir == "" {
		t.Fatal("GlobalDir() returned empty string")
	}
	if filepath.Base(dir) != DirName {
		t.Errorf("GlobalDir() = %q, want basename %q", dir, DirName)
	}
}

func TestGlobalDir_EndsWithDeets(t *testing.T) {
	dir := GlobalDir()
	if !hasSuffix(dir, DirName) {
		t.Errorf("GlobalDir() = %q does not end with %q", dir, DirName)
	}
}

// ---------------------------------------------------------------------------
// GlobalFile
// ---------------------------------------------------------------------------

func TestGlobalFile_NonEmpty(t *testing.T) {
	f := GlobalFile()
	if f == "" {
		t.Fatal("GlobalFile() returned empty string")
	}
	if filepath.Base(f) != FileName {
		t.Errorf("GlobalFile() = %q, want basename %q", f, FileName)
	}
}

func TestGlobalFile_EndsWithMeToml(t *testing.T) {
	f := GlobalFile()
	if !hasSuffix(f, FileName) {
		t.Errorf("GlobalFile() = %q does not end with %q", f, FileName)
	}
}

func TestGlobalFile_ContainsGlobalDir(t *testing.T) {
	dir := GlobalDir()
	f := GlobalFile()
	if filepath.Dir(f) != dir {
		t.Errorf("GlobalFile() dir = %q, want %q", filepath.Dir(f), dir)
	}
}

// ---------------------------------------------------------------------------
// FindLocalDir
// ---------------------------------------------------------------------------

func TestFindLocalDir_NotFound(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	got := FindLocalDir()
	if got != "" {
		t.Errorf("FindLocalDir() = %q, want empty string (no .deets anywhere)", got)
	}
}

func TestFindLocalDir_FoundInCurrentDir(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	chdir(t, tmp)

	got := FindLocalDir()
	if got != deetsDir {
		t.Errorf("FindLocalDir() = %q, want %q", got, deetsDir)
	}
}

func TestFindLocalDir_FoundInParentDir(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(tmp, "subdir")
	if err := os.Mkdir(child, 0755); err != nil {
		t.Fatal(err)
	}
	chdir(t, child)

	got := FindLocalDir()
	if got != deetsDir {
		t.Errorf("FindLocalDir() = %q, want %q", got, deetsDir)
	}
}

func TestFindLocalDir_FoundInGrandparentDir(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(tmp, "a", "b")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	chdir(t, child)

	got := FindLocalDir()
	if got != deetsDir {
		t.Errorf("FindLocalDir() = %q, want %q", got, deetsDir)
	}
}

func TestFindLocalDir_FileNotDir(t *testing.T) {
	// If .deets exists but is a regular file (not a directory), it should not
	// be treated as a valid local config directory.
	tmp := t.TempDir()
	deetsFile := filepath.Join(tmp, DirName)
	if err := os.WriteFile(deetsFile, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, tmp)

	got := FindLocalDir()
	if got != "" {
		t.Errorf("FindLocalDir() = %q, want empty string (.deets is a file, not a dir)", got)
	}
}

func TestFindLocalDir_StopsAtHomeDir(t *testing.T) {
	// The function checks `if dir == home { break }` BEFORE looking for .deets,
	// so ~/.deets/ must NOT be returned as a local directory.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot determine home directory: %v", err)
	}

	// Ensure ~/.deets/ actually exists so the test is meaningful.
	homeDeetsDir := filepath.Join(home, DirName)
	created := false
	if _, err := os.Stat(homeDeetsDir); os.IsNotExist(err) {
		if err := os.Mkdir(homeDeetsDir, 0755); err != nil {
			t.Skipf("cannot create %s for test: %v", homeDeetsDir, err)
		}
		created = true
	}
	if created {
		t.Cleanup(func() { os.Remove(homeDeetsDir) })
	}

	// Create a child directory inside $HOME that has no .deets/ of its own.
	child := filepath.Join(home, "config_test_tmp_subdir")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(child) })

	chdir(t, child)

	got := FindLocalDir()
	if got == homeDeetsDir {
		t.Errorf("FindLocalDir() = %q; it should NOT return ~/.deets/ as a local directory", got)
	}
}

// ---------------------------------------------------------------------------
// FindLocalFile
// ---------------------------------------------------------------------------

func TestFindLocalFile_Exists(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	meFile := filepath.Join(deetsDir, FileName)
	if err := os.WriteFile(meFile, []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, tmp)

	got := FindLocalFile()
	if got != meFile {
		t.Errorf("FindLocalFile() = %q, want %q", got, meFile)
	}
}

func TestFindLocalFile_DirExistsButNoFile(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	// No me.toml created inside .deets/
	chdir(t, tmp)

	got := FindLocalFile()
	if got != "" {
		t.Errorf("FindLocalFile() = %q, want empty string (no me.toml in .deets/)", got)
	}
}

func TestFindLocalFile_NoDirAtAll(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	got := FindLocalFile()
	if got != "" {
		t.Errorf("FindLocalFile() = %q, want empty string (no .deets/ at all)", got)
	}
}

// ---------------------------------------------------------------------------
// ResolvePaths
// ---------------------------------------------------------------------------

func TestResolvePaths_WithLocal(t *testing.T) {
	tmp := t.TempDir()
	deetsDir := filepath.Join(tmp, DirName)
	if err := os.Mkdir(deetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	meFile := filepath.Join(deetsDir, FileName)
	if err := os.WriteFile(meFile, []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}
	chdir(t, tmp)

	p, err := ResolvePaths()
	if err != nil {
		t.Fatalf("ResolvePaths() error: %v", err)
	}

	if p.GlobalDir == "" {
		t.Error("ResolvePaths().GlobalDir is empty")
	}
	if p.GlobalFile == "" {
		t.Error("ResolvePaths().GlobalFile is empty")
	}
	if p.LocalDir != deetsDir {
		t.Errorf("ResolvePaths().LocalDir = %q, want %q", p.LocalDir, deetsDir)
	}
	if p.LocalFile != meFile {
		t.Errorf("ResolvePaths().LocalFile = %q, want %q", p.LocalFile, meFile)
	}
	if !p.HasLocal {
		t.Error("ResolvePaths().HasLocal = false, want true")
	}
}

func TestResolvePaths_WithoutLocal(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	p, err := ResolvePaths()
	if err != nil {
		t.Fatalf("ResolvePaths() error: %v", err)
	}

	if p.GlobalDir == "" {
		t.Error("ResolvePaths().GlobalDir is empty")
	}
	if p.GlobalFile == "" {
		t.Error("ResolvePaths().GlobalFile is empty")
	}
	if p.LocalDir != "" {
		t.Errorf("ResolvePaths().LocalDir = %q, want empty", p.LocalDir)
	}
	if p.LocalFile != "" {
		t.Errorf("ResolvePaths().LocalFile = %q, want empty", p.LocalFile)
	}
	if p.HasLocal {
		t.Error("ResolvePaths().HasLocal = true, want false")
	}
}

func TestResolvePaths_GlobalPathsAlwaysPopulated(t *testing.T) {
	// Regardless of local state, global paths must always be set.
	tmp := t.TempDir()
	chdir(t, tmp)

	p, err := ResolvePaths()
	if err != nil {
		t.Fatalf("ResolvePaths() error: %v", err)
	}

	home, _ := os.UserHomeDir()
	wantGlobalDir := filepath.Join(home, DirName)
	wantGlobalFile := filepath.Join(home, DirName, FileName)

	if p.GlobalDir != wantGlobalDir {
		t.Errorf("ResolvePaths().GlobalDir = %q, want %q", p.GlobalDir, wantGlobalDir)
	}
	if p.GlobalFile != wantGlobalFile {
		t.Errorf("ResolvePaths().GlobalFile = %q, want %q", p.GlobalFile, wantGlobalFile)
	}
}

// ---------------------------------------------------------------------------
// EnsureGlobalDir
// ---------------------------------------------------------------------------

func TestEnsureGlobalDir_CreatesDir(t *testing.T) {
	// We cannot easily redirect $HOME, so we just verify the function does
	// not error and that GlobalDir() exists afterwards.
	if err := EnsureGlobalDir(); err != nil {
		t.Fatalf("EnsureGlobalDir() error: %v", err)
	}

	dir := GlobalDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("GlobalDir %q does not exist after EnsureGlobalDir: %v", dir, err)
	}
	if !info.IsDir() {
		t.Errorf("GlobalDir %q is not a directory", dir)
	}
}

func TestEnsureGlobalDir_Idempotent(t *testing.T) {
	if err := EnsureGlobalDir(); err != nil {
		t.Fatalf("EnsureGlobalDir() first call error: %v", err)
	}
	if err := EnsureGlobalDir(); err != nil {
		t.Fatalf("EnsureGlobalDir() second call error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// EnsureLocalDir
// ---------------------------------------------------------------------------

func TestEnsureLocalDir_CreatesDir(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	if err := EnsureLocalDir(); err != nil {
		t.Fatalf("EnsureLocalDir() error: %v", err)
	}

	expected := filepath.Join(tmp, DirName)
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("%q does not exist after EnsureLocalDir: %v", expected, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", expected)
	}
}

func TestEnsureLocalDir_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	if err := EnsureLocalDir(); err != nil {
		t.Fatalf("EnsureLocalDir() first call error: %v", err)
	}
	if err := EnsureLocalDir(); err != nil {
		t.Fatalf("EnsureLocalDir() second call error: %v", err)
	}

	expected := filepath.Join(tmp, DirName)
	info, err := os.Stat(expected)
	if err != nil {
		t.Fatalf("%q does not exist after second EnsureLocalDir: %v", expected, err)
	}
	if !info.IsDir() {
		t.Errorf("%q is not a directory", expected)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// chdir changes the working directory to dir and restores it when the test
// finishes.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd(): %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("os.Chdir(%q): %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Logf("warning: could not restore cwd to %q: %v", orig, err)
		}
	})
}

// hasSuffix checks whether path ends with the given suffix component.
func hasSuffix(path, suffix string) bool {
	return len(path) >= len(suffix) &&
		path[len(path)-len(suffix):] == suffix
}
