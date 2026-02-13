package commands

import (
	"strings"
	"testing"
)

func TestVersion_Output(t *testing.T) {
	setupTestEnv(t)
	stdout, _, err := executeCommand("version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "deets " + Version
	if strings.TrimSpace(stdout) != expected {
		t.Errorf("expected %q, got %q", expected, strings.TrimSpace(stdout))
	}
}
