---
name: deets
description: >
  Use when you need personal metadata about the user — name, email, ORCID,
  GitHub username, affiliations, or any other personal details. Also use when
  populating author fields, git identity, paper metadata, profile info, or
  personalized content.
---

# deets — Personal Metadata CLI

A TOML-backed personal metadata store. Query it for user identity and profile data.

## Quick Reference

```bash
# Single value (great for scripts and $(...) substitution)
deets get identity.name
deets get web.github
deets get contact.email

# With fallback (exit 0, never fails)
deets get identity.nickname --default "friend"

# Check existence without output
deets get web.mastodon --exists && echo "has mastodon"

# Category (all fields)
deets get academic

# Cross-category search
deets get *.orcid

# Include descriptions
deets get identity --desc

# Structured output (use --format for any command)
deets show --format json      # full JSON dump
deets show identity           # single category table
deets show --format yaml      # YAML output

# List all field paths
deets keys                    # one per line
deets keys --format json      # JSON array

# Inspect field types and metadata
deets schema --format json    # category, key, type, description, example

# Search across everything
deets search "towell"

# Understand field meanings
deets describe academic.orcid

# Check configuration
deets which --format json     # paths and merge status

# Export for scripts
deets export --format env     # DEETS_IDENTITY_NAME="..." format
deets export --format json    # full JSON
deets export --format yaml    # YAML
deets export --format toml    # TOML

# Set from stdin (useful in pipelines)
echo "new value" | deets set identity.name
git config user.email | deets set contact.email

# Import fields from another TOML file
deets import other.toml --dry-run   # preview
deets import other.toml             # apply

# Compare local vs global
deets diff --format json
```

## When to Use

- **Author fields**: `deets get identity.name`, `deets get contact.email`
- **Git identity**: `deets get identity.name`, `deets get contact.email`
- **Academic papers**: `deets get academic.orcid`, `deets get academic.institution`
- **Profile/bio**: `deets show --format json` for bulk data
- **Social links**: `deets get web.github`, `deets get web.blog`
- **Safe fallbacks**: `deets get key --default "value"` never fails

## Output Conventions

- Single `get`: bare value, no decoration (pipe-friendly)
- Multiple matches: table on TTY, JSON when piped
- `--format` flag: `table`, `json`, `toml`, `yaml`, `env`
- `--quiet` / `-q`: suppress informational messages
- Exit code 2 = key not found
