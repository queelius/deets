package store

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// --- DefaultTemplate tests ---

func TestDefaultTemplate_IsValidTOML(t *testing.T) {
	// The template contains commented-out keys. Strip comment lines to get
	// a parseable TOML document with section headers only.
	lines := strings.Split(DefaultTemplate, "\n")
	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	tomlContent := strings.Join(cleaned, "\n")

	var raw map[string]interface{}
	if err := toml.Unmarshal([]byte(tomlContent), &raw); err != nil {
		t.Fatalf("DefaultTemplate is not valid TOML after stripping comments: %v", err)
	}

	// Verify expected sections exist in the stripped template.
	expectedSections := []string{"identity", "contact", "academic", "education", "profiles"}
	for _, section := range expectedSections {
		if _, ok := raw[section]; !ok {
			t.Errorf("expected section %q in DefaultTemplate", section)
		}
	}
}

func TestDefaultTemplate_IsNotEmpty(t *testing.T) {
	if strings.TrimSpace(DefaultTemplate) == "" {
		t.Error("DefaultTemplate should not be empty")
	}
}

func TestDefaultTemplate_ContainsExpectedSections(t *testing.T) {
	expectedSections := []string{"[identity]", "[contact]", "[academic]", "[education]", "[profiles.github]", "[profiles.pypi]", "[profiles.orcid]"}
	for _, section := range expectedSections {
		if !strings.Contains(DefaultTemplate, section) {
			t.Errorf("DefaultTemplate should contain %q", section)
		}
	}
}

func TestDefaultTemplate_ContainsInstructions(t *testing.T) {
	if !strings.Contains(DefaultTemplate, "deets") {
		t.Error("DefaultTemplate should reference 'deets'")
	}
	if !strings.Contains(DefaultTemplate, "populate") {
		t.Error("DefaultTemplate should mention populate command")
	}
}

func TestLocalTemplate_IsNotEmpty(t *testing.T) {
	if strings.TrimSpace(LocalTemplate) == "" {
		t.Error("LocalTemplate should not be empty")
	}
}

func TestLocalTemplate_ContainsOverrideInstruction(t *testing.T) {
	if !strings.Contains(LocalTemplate, "override") {
		t.Error("LocalTemplate should mention overrides")
	}
}

// --- DefaultDescriptions tests ---

func TestDefaultDescriptions_HasExpectedCategories(t *testing.T) {
	expectedCategories := []string{
		"identity", "contact", "academic", "education",
		"profiles.github", "profiles.pypi", "profiles.cran",
		"profiles.orcid", "profiles.bluesky", "profiles.mastodon",
		"profiles.blog", "profiles.zenodo", "profiles.linkedin",
	}
	for _, cat := range expectedCategories {
		if _, ok := DefaultDescriptions[cat]; !ok {
			t.Errorf("DefaultDescriptions should have category %q", cat)
		}
	}
}

func TestDefaultDescriptions_HasExpectedKeys(t *testing.T) {
	tests := []struct {
		category string
		key      string
	}{
		{"identity", "name"},
		{"identity", "aka"},
		{"identity", "pronouns"},
		{"contact", "email"},
		{"contact", "phone"},
		{"academic", "orcid"},
		{"academic", "institution"},
		{"academic", "title"},
		{"academic", "research_interests"},
		{"academic", "scholar"},
		{"education", "degrees"},
		{"education", "field"},
		{"education", "institution"},
		{"education", "phd"},
		{"education", "phd_institution"},
		{"profiles.github", "username"},
		{"profiles.github", "name"},
		{"profiles.github", "email"},
		{"profiles.github", "url"},
		{"profiles.github", "bio"},
		{"profiles.pypi", "username"},
		{"profiles.pypi", "name"},
		{"profiles.pypi", "email"},
		{"profiles.pypi", "url"},
		{"profiles.cran", "name"},
		{"profiles.cran", "email"},
		{"profiles.orcid", "id"},
		{"profiles.orcid", "name"},
		{"profiles.orcid", "email"},
		{"profiles.orcid", "url"},
		{"profiles.bluesky", "handle"},
		{"profiles.bluesky", "url"},
		{"profiles.mastodon", "handle"},
		{"profiles.mastodon", "url"},
		{"profiles.blog", "url"},
		{"profiles.blog", "alt"},
		{"profiles.zenodo", "username"},
		{"profiles.zenodo", "url"},
		{"profiles.linkedin", "url"},
	}

	for _, tc := range tests {
		catDescs, ok := DefaultDescriptions[tc.category]
		if !ok {
			t.Errorf("missing category %q in DefaultDescriptions", tc.category)
			continue
		}
		desc, ok := catDescs[tc.key]
		if !ok {
			t.Errorf("missing key %q in DefaultDescriptions[%q]", tc.key, tc.category)
			continue
		}
		if desc == "" {
			t.Errorf("empty description for DefaultDescriptions[%q][%q]", tc.category, tc.key)
		}
	}
}

func TestDefaultDescriptions_AllDescriptionsNonEmpty(t *testing.T) {
	for cat, fields := range DefaultDescriptions {
		for key, desc := range fields {
			if desc == "" {
				t.Errorf("DefaultDescriptions[%q][%q] has empty description", cat, key)
			}
		}
	}
}

func TestDefaultDescriptions_IdentityCategoryContents(t *testing.T) {
	identity, ok := DefaultDescriptions["identity"]
	if !ok {
		t.Fatal("missing 'identity' in DefaultDescriptions")
	}

	if identity["name"] != "Full legal name" {
		t.Errorf("expected identity.name = 'Full legal name', got %q", identity["name"])
	}
	if identity["aka"] != "Known aliases and nicknames" {
		t.Errorf("expected identity.aka = 'Known aliases and nicknames', got %q", identity["aka"])
	}
	if identity["pronouns"] != "Personal pronouns" {
		t.Errorf("expected identity.pronouns = 'Personal pronouns', got %q", identity["pronouns"])
	}
}

func TestDefaultDescriptions_ContactCategoryContents(t *testing.T) {
	contact, ok := DefaultDescriptions["contact"]
	if !ok {
		t.Fatal("missing 'contact' in DefaultDescriptions")
	}

	if contact["email"] != "Primary email address" {
		t.Errorf("expected contact.email = 'Primary email address', got %q", contact["email"])
	}
	if contact["phone"] != "Phone number" {
		t.Errorf("expected contact.phone = 'Phone number', got %q", contact["phone"])
	}
}

func TestDefaultDescriptions_ProfilesGitHubCategoryContents(t *testing.T) {
	gh, ok := DefaultDescriptions["profiles.github"]
	if !ok {
		t.Fatal("missing 'profiles.github' in DefaultDescriptions")
	}

	if gh["username"] != "GitHub username" {
		t.Errorf("expected profiles.github.username = 'GitHub username', got %q", gh["username"])
	}
	if gh["url"] != "GitHub profile URL" {
		t.Errorf("expected profiles.github.url = 'GitHub profile URL', got %q", gh["url"])
	}
	if gh["bio"] != "GitHub profile bio" {
		t.Errorf("expected profiles.github.bio = 'GitHub profile bio', got %q", gh["bio"])
	}
}

func TestDefaultDescriptions_AcademicCategoryContents(t *testing.T) {
	academic, ok := DefaultDescriptions["academic"]
	if !ok {
		t.Fatal("missing 'academic' in DefaultDescriptions")
	}

	if academic["orcid"] != "ORCID persistent digital identifier" {
		t.Errorf("expected academic.orcid = 'ORCID persistent digital identifier', got %q", academic["orcid"])
	}
	if academic["institution"] != "Academic institution" {
		t.Errorf("expected academic.institution = 'Academic institution', got %q", academic["institution"])
	}
	if academic["scholar"] != "Google Scholar profile ID" {
		t.Errorf("expected academic.scholar = 'Google Scholar profile ID', got %q", academic["scholar"])
	}
}

func TestDefaultDescriptions_EducationCategoryContents(t *testing.T) {
	education, ok := DefaultDescriptions["education"]
	if !ok {
		t.Fatal("missing 'education' in DefaultDescriptions")
	}

	if education["degrees"] != "Completed degrees with institution and year" {
		t.Errorf("expected education.degrees = 'Completed degrees with institution and year', got %q", education["degrees"])
	}
	if education["field"] != "Primary field of study" {
		t.Errorf("expected education.field = 'Primary field of study', got %q", education["field"])
	}
	if education["institution"] != "Degree-granting institution" {
		t.Errorf("expected education.institution = 'Degree-granting institution', got %q", education["institution"])
	}
	if education["phd"] != "PhD field of study" {
		t.Errorf("expected education.phd = 'PhD field of study', got %q", education["phd"])
	}
	if education["phd_institution"] != "PhD institution" {
		t.Errorf("expected education.phd_institution = 'PhD institution', got %q", education["phd_institution"])
	}
}
