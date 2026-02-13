package commands

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestDescribe_AllDescriptions(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("describe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should include at least one description from the test data.
	if !strings.Contains(stdout, "Full legal name") {
		t.Errorf("expected description in output, got %q", stdout)
	}
}

func TestDescribe_SingleCategory(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("describe", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Full legal name") {
		t.Errorf("expected name description, got %q", stdout)
	}
}

func TestDescribe_SingleField(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	stdout, _, err := executeCommand("describe", "academic.orcid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "ORCID persistent digital identifier") {
		t.Errorf("expected orcid description, got %q", stdout)
	}
}

func TestDescribe_FieldNotFound(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("describe", "identity.nonexistent")
	if err == nil {
		t.Fatal("expected error for no description")
	}
	var exitErr *ExitError
	if !errors.As(err, &exitErr) || exitErr.Code != 2 {
		t.Errorf("expected ExitError with code 2, got %v", err)
	}
}

func TestDescribe_SetDescription(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"

	// Set a description.
	_, _, err := executeCommand("describe", "web.mastodon", "Mastodon handle")
	if err != nil {
		t.Fatalf("unexpected error setting description: %v", err)
	}

	// Read it back — mastodon key itself doesn't exist in the test data, but
	// we can set the description. The describe command reads from _desc field.
	// Since the mastodon key itself doesn't exist, the describe command won't
	// find it via AllDescriptions — but we can verify the file was written.
	// Let's set both the key and description, then read back.
	_, _, err = executeCommand("set", "web.mastodon", "@user@mastodon.social")
	if err != nil {
		t.Fatalf("unexpected error setting field: %v", err)
	}

	stdout, _, err := executeCommand("describe", "web.mastodon")
	if err != nil {
		t.Fatalf("unexpected error reading description: %v", err)
	}
	if !strings.Contains(stdout, "Mastodon handle") {
		t.Errorf("expected set description, got %q", stdout)
	}
}

func TestDescribe_DefaultDescriptionFallback(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	// web.website has no explicit _desc in test data, but DefaultDescriptions
	// has a fallback for it: "Personal website URL".
	stdout, _, err := executeCommand("describe", "web.website")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout, "Personal website URL") {
		t.Errorf("expected default description fallback, got %q", stdout)
	}
}

func TestDescribe_InvalidPath(t *testing.T) {
	setupTestDB(t)
	flagFormat = "table"
	_, _, err := executeCommand("describe", "a b.key", "desc")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
}

func TestDescribe_FormatJSON(t *testing.T) {
	setupTestDB(t)
	flagFormat = "json"
	stdout, _, err := executeCommand("describe", "identity")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !json.Valid([]byte(strings.TrimSpace(stdout))) {
		t.Errorf("expected valid JSON, got %q", stdout)
	}
}
