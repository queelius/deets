package store

// DefaultTemplate is the default me.toml content for `deets init`.
const DefaultTemplate = `# deets — Personal metadata
# Edit this file to add your personal details.
# Any [category] with any key = "value" is valid.
# Add _desc suffix for self-describing fields:
#   orcid = "0000-..."
#   orcid_desc = "ORCID persistent digital identifier"

[identity]
# name = "Your Name"
# name_desc = "Full legal name"
# aka = ["Nickname"]
# aka_desc = "Known aliases and nicknames"
# pronouns = "they/them"

[contact]
# email = "you@example.com"
# email_desc = "Primary email address"

[web]
# github = "username"
# blog = "https://example.com"

[academic]
# orcid = "0000-0000-0000-0000"
# orcid_desc = "ORCID persistent digital identifier"
# institution = "University of..."
# title = "..."
# research_interests = ["topic1", "topic2"]

[education]
# degrees = ["BS Computer Science (University, 2020)"]
# degrees_desc = "Completed degrees with institution and year"
# field = "Computer Science"
# institution = "University of..."
`

// LocalTemplate is the minimal template for local overrides.
const LocalTemplate = `# deets — Local project overrides
# Keys here override matching keys from ~/.deets/me.toml
# Only include fields you want to override for this project.
`

// DefaultDescriptions provides built-in fallback descriptions for well-known
// fields, keyed by category then field name.
var DefaultDescriptions = map[string]map[string]string{
	"identity": {
		"name":     "Full legal name",
		"aka":      "Known aliases and nicknames",
		"pronouns": "Personal pronouns",
	},
	"contact": {
		"email": "Primary email address",
		"phone": "Phone number",
	},
	"web": {
		"github":   "GitHub username",
		"blog":     "Personal blog URL",
		"website":  "Personal website URL",
		"mastodon": "Mastodon handle",
		"twitter":  "Twitter/X handle",
		"linkedin": "LinkedIn profile URL",
		"bluesky":  "Bluesky handle",
	},
	"academic": {
		"orcid":              "ORCID persistent digital identifier",
		"institution":        "Academic institution",
		"title":              "Academic title or position",
		"research_interests": "Research interest areas",
		"scholar":            "Google Scholar ID",
	},
	"education": {
		"degrees":     "Completed degrees with institution and year",
		"field":       "Primary field of study",
		"institution": "Degree-granting institution",
	},
}
