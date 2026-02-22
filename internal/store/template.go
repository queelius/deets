package store

// DefaultTemplate is the default me.toml content for `deets init`.
const DefaultTemplate = `# deets — Personal metadata
# Any [category] with any key = "value" is valid.
# Run 'deets populate --all' to auto-fill from your accounts.

[identity]
# name = "Your Name"
# canonical_name = "Your Full Legal Name"
# aka = ["Nickname"]
# pronouns = "they/them"

[contact]
# email = "you@example.com"
# emails = ["you@example.com"]

[academic]
# orcid = "0000-0000-0000-0000"
# institution = "University of..."
# research_interests = ["topic1", "topic2"]
# scholar = "Google Scholar ID"
# degrees = ["BS Computer Science (University, 2020)"]
# phd = "Field of study"

[web]
# handle = "your-username"
# domains = ["example.com"]

[platforms.github]
# url = "https://github.com/your-username"

[platforms.pypi]
# url = "https://pypi.org/user/your-username"

[platforms.orcid]
# url = "https://orcid.org/0000-0000-0000-0000"
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
		"name":           "Full legal name",
		"canonical_name": "Full canonical name including middle name",
		"aka":            "Known aliases and nicknames",
		"pronouns":       "Personal pronouns",
	},
	"contact": {
		"email":  "Primary email address",
		"emails": "All email addresses",
		"phone":  "Phone number",
	},
	"academic": {
		"orcid":              "ORCID persistent digital identifier",
		"institution":        "Academic institution",
		"title":              "Academic title or position",
		"research_interests": "Research interest areas",
		"scholar":            "Google Scholar profile ID",
		"degrees":            "Completed degrees with institution and year",
		"phd":                "PhD field of study",
		"phd_institution":    "PhD institution",
	},
	"web": {
		"handle":  "Primary username/handle across platforms",
		"domains": "Personal domains",
	},
	"platforms.github": {
		"url":  "GitHub profile URL",
		"name": "Display name on GitHub",
		"bio":  "GitHub profile bio",
	},
	"platforms.pypi": {
		"url": "PyPI profile URL",
	},
	"platforms.cran": {},
	"platforms.orcid": {
		"url": "ORCID profile URL",
	},
	"platforms.bluesky": {
		"handle": "Bluesky handle",
		"url":    "Bluesky profile URL",
	},
	"platforms.mastodon": {
		"handle": "Mastodon handle",
		"url":    "Mastodon profile URL",
	},
	"platforms.blog": {
		"url": "Blog URL",
		"alt": "Alternative blog URL",
	},
	"platforms.zenodo": {
		"url": "Zenodo profile URL",
	},
	"platforms.linkedin": {
		"url": "LinkedIn profile URL",
	},
	"platforms.r-universe": {
		"url": "R-universe profile URL",
	},
}
