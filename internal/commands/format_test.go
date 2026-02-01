package commands

import (
	"testing"
)

func TestValidateFormat_ValidFormats(t *testing.T) {
	for _, fmt := range []string{"table", "json", "toml", "yaml", "env"} {
		flagFormat = fmt
		if err := validateFormat(); err != nil {
			t.Errorf("validateFormat(%q) returned error: %v", fmt, err)
		}
	}
}

func TestValidateFormat_EmptyIsValid(t *testing.T) {
	flagFormat = ""
	if err := validateFormat(); err != nil {
		t.Errorf("validateFormat(\"\") should not error, got: %v", err)
	}
}

func TestValidateFormat_InvalidFormat(t *testing.T) {
	flagFormat = "xml"
	err := validateFormat()
	if err == nil {
		t.Error("validateFormat(\"xml\") should return error")
	}
}

func TestResolveFormat_ExplicitFlag(t *testing.T) {
	flagFormat = "yaml"
	if got := resolveFormat(); got != "yaml" {
		t.Errorf("resolveFormat() = %q, want %q", got, "yaml")
	}
}

func TestResolveFormat_EmptyDefaultsBasedOnTTY(t *testing.T) {
	flagFormat = ""
	// When running in tests, stdout is not a TTY, so default is "json"
	got := resolveFormat()
	if got != "json" {
		t.Errorf("resolveFormat() in non-TTY = %q, want %q", got, "json")
	}
}
