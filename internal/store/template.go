package store

// DefaultTemplate is the default me.toml content for `deets init`.
const DefaultTemplate = `# deets — Personal metadata
# Any [category] with any key = "value" is valid.
# Run 'deets populate --all' to auto-fill from your accounts.

[identity]
# name = "Your Name"
# aka = ["Nickname"]
# pronouns = "they/them"

[contact]
# email = "you@example.com"

[academic]
# orcid = "0000-0000-0000-0000"
# institution = "University of..."
# research_interests = ["topic1", "topic2"]
# scholar = "Google Scholar ID"

[education]
# degrees = ["BS Computer Science (University, 2020)"]
# phd = "Field of study"

[profiles.github]
# username = "your-username"

[profiles.pypi]
# username = "your-username"

[profiles.orcid]
# id = "0000-0000-0000-0000"
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
	"academic": {
		"orcid":              "ORCID persistent digital identifier",
		"institution":        "Academic institution",
		"title":              "Academic title or position",
		"research_interests": "Research interest areas",
		"scholar":            "Google Scholar profile ID",
	},
	"education": {
		"degrees":         "Completed degrees with institution and year",
		"field":           "Primary field of study",
		"institution":     "Degree-granting institution",
		"phd":             "PhD field of study",
		"phd_institution": "PhD institution",
	},
	"profiles.github": {
		"username": "GitHub username",
		"name":     "Display name on GitHub",
		"email":    "Email associated with GitHub account",
		"url":      "GitHub profile URL",
		"bio":      "GitHub profile bio",
	},
	"profiles.pypi": {
		"username": "PyPI username",
		"name":     "Author name for PyPI packages",
		"email":    "Email for PyPI packages",
		"url":      "PyPI profile URL",
	},
	"profiles.cran": {
		"name":  "Author name for CRAN packages",
		"email": "Email for CRAN packages",
	},
	"profiles.orcid": {
		"id":    "ORCID identifier",
		"name":  "Name registered with ORCID",
		"email": "Email registered with ORCID",
		"url":   "ORCID profile URL",
	},
	"profiles.bluesky": {
		"handle": "Bluesky handle",
		"url":    "Bluesky profile URL",
	},
	"profiles.mastodon": {
		"handle": "Mastodon handle",
		"url":    "Mastodon profile URL",
	},
	"profiles.blog": {
		"url": "Blog URL",
		"alt": "Alternative blog URL",
	},
	"profiles.zenodo": {
		"username": "Zenodo username",
		"url":      "Zenodo profile URL",
	},
	"profiles.linkedin": {
		"url": "LinkedIn profile URL",
	},
}
