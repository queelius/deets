package commands

import (
	"testing"
)

func TestParsePath_Valid(t *testing.T) {
	tests := []struct {
		input    string
		wantCat  string
		wantKey  string
	}{
		{"identity.name", "identity", "name"},
		{"web.github", "web", "github"},
		{"my-cat.my-key", "my-cat", "my-key"},
		{"a.b", "a", "b"},
	}
	for _, tt := range tests {
		cat, key, err := parsePath(tt.input)
		if err != nil {
			t.Errorf("parsePath(%q) returned error: %v", tt.input, err)
			continue
		}
		if cat != tt.wantCat || key != tt.wantKey {
			t.Errorf("parsePath(%q) = (%q, %q), want (%q, %q)", tt.input, cat, key, tt.wantCat, tt.wantKey)
		}
	}
}

func TestParsePath_Invalid(t *testing.T) {
	tests := []struct {
		input string
		desc  string
	}{
		{"noperiod", "no dot separator"},
		{".key", "empty category"},
		{"category.", "empty key"},
		{"", "empty string"},
		{"a.b.c", "key contains dot (b.c fails validation)"},
		{"my category.key", "space in category"},
		{"cat.my key", "space in key"},
		{"evil].name", "bracket in category"},
		{"cat.evil]", "bracket in key"},
	}
	for _, tt := range tests {
		_, _, err := parsePath(tt.input)
		if err == nil {
			t.Errorf("parsePath(%q) should error (%s), got nil", tt.input, tt.desc)
		}
	}
}

func TestExitError_WithMessage(t *testing.T) {
	err := &ExitError{Code: 2, Message: "not found"}
	if err.Error() != "not found" {
		t.Errorf("expected 'not found', got %q", err.Error())
	}
}

func TestExitError_WithoutMessage(t *testing.T) {
	err := &ExitError{Code: 1}
	if err.Error() != "exit code 1" {
		t.Errorf("expected 'exit code 1', got %q", err.Error())
	}
}
