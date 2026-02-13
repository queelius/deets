# Agent-Native Overhaul Design

**Date:** 2026-02-13
**Status:** Approved
**Version target:** v0.8

## Motivation

deets exists primarily so coding agents (Claude Code, etc.) can query personal metadata without asking the user repeatedly. The current schema grew organically and has structural problems that make agent consumption harder than it should be:

- Emails scattered across categories with ad-hoc naming (`contact.email`, `packages.pypi_email`)
- No way to query "all emails" or "the email for PyPI" without knowing every key name
- Platform identities (GitHub, PyPI, ORCID) spread across `[web]`, `[packages]`, `[publications]`
- Manual population is high friction — the user already has this data in git config, GitHub, ORCID
- The skill file is 107 lines of examples that go stale when fields change

## Design Principles

1. **Agent-first:** Structure data around how agents query it ("what's my GitHub identity?"), not how humans categorize it
2. **Schema is the API:** Agents discover available fields via `deets schema --format json`, not by reading a skill file
3. **Zero-friction population:** `deets populate` harvests from sources you already use
4. **Clean break:** No backwards compatibility with the old `[web]`/`[packages]` schema. This is a v0.8 restructure.

---

## 1. Platform-Centric Schema

### Structure

```toml
# Core identity — canonical source of truth
[identity]
name = "Alexander Towell"
aka = ["Alex Towell"]
pronouns = "he/him"

[contact]
email = "lex@metafunctor.com"

[academic]
orcid = "0000-0001-6443-9897"
institution = "Southern Illinois University Edwardsville"
research_interests = ["information retrieval", "Bayesian statistics"]
scholar = "E9mnFzQAAAAJ"

[education]
degrees = ["MS Computer Science (SIUE, 2015)", "MS Mathematics (SIUE, 2023)"]
phd = "Computer Science"
phd_institution = "Southern Illinois University Edwardsville"

# Platform profiles — agent-queryable, self-contained contexts
[profiles.github]
username = "queelius"
name = "Alex Towell"
email = "queelius@gmail.com"
url = "https://github.com/queelius"

[profiles.pypi]
username = "queelius"
name = "Alex Towell"
email = "lex@metafunctor.com"
url = "https://pypi.org/user/queelius"

[profiles.cran]
name = "Alexander Towell"
email = "lex@metafunctor.com"

[profiles.orcid]
id = "0000-0001-6443-9897"
name = "Alexander Towell"
email = "atowell@siue.edu"
url = "https://orcid.org/0000-0001-6443-9897"

[profiles.bluesky]
handle = "queelius.bsky.social"
url = "https://bsky.app/profile/queelius.bsky.social"

[profiles.mastodon]
handle = "@queelius@mastodon.social"
url = "https://mastodon.social/@queelius"

[profiles.blog]
url = "https://metafunctor.com"
alt = "https://queelius.github.io"

[profiles.zenodo]
username = "queelius"
```

### Agent query patterns

```bash
deets get profiles.github           # all GitHub identity fields
deets get profiles.pypi.email       # exact: "lex@metafunctor.com"
deets get profiles.*.email          # all platform emails
deets get profiles.*.url            # all profile URLs
deets get identity.name             # canonical name
```

### Design decisions

- **Duplication is intentional.** The same email may appear under multiple profiles. This is a feature: `deets get profiles.pypi.email` is unambiguous. The populate command keeps them in sync.
- **Every platform gets a `url` field** where applicable. Agents building profile pages can glob `profiles.*.url`.
- **Old categories removed.** `[web]`, `[packages]`, `[publications]` are replaced by `[profiles.*]`. The `[publications]` category's `scholar` field moves to `[academic]`.

---

## 2. Parser Changes

### Problem

`[profiles.github]` in TOML produces nested maps:
```go
{"profiles": {"github": {"username": "queelius", ...}}}
```

The current `LoadFile` expects one level of nesting (top-level key = category, value = flat map of fields). Nested maps are silently skipped.

### Fix

In `LoadFile`, when iterating a category's values, if a value is itself a `map[string]interface{}`, flatten it: concatenate the parent key and child key with a dot to form the category name.

```
"profiles" + "github" → category "profiles.github" with fields {username, name, email, url}
```

This is ~15 lines of code. The line-level writer already handles dotted category names correctly because `fmt.Sprintf("[%s]", category)` produces `[profiles.github]`, which BurntSushi parses back as a nested table. Round-trip works.

The flattening is general — `[a.b.c]` becomes category `a.b.c` with flat fields. No depth limit needed.

### ValidateName impact

The `ValidateName` function currently rejects dots. Dotted category names like `profiles.github` are composed of two valid bare keys joined by a dot. The validation needs to allow dots in category names when they come from TOML parsing (they represent table nesting), while still rejecting them in user-supplied key names.

Approach: `ValidateName` stays strict (no dots). `parsePath` splits on the first dot and validates each part. The store functions that receive dotted category names from internal code paths (like populate) bypass `parsePath` validation — `SetValue` validates the category with a separate function that allows dots.

---

## 3. Interactive Harvest (`deets populate`)

### Command interface

```bash
deets populate --git              # local git config
deets populate --github           # GitHub API via gh CLI
deets populate --orcid            # ORCID public API
deets populate --all              # all available sources
deets populate --github --dry-run # preview only
deets populate --github --yes     # skip confirmation
```

### Flow

1. **Fetch** data from the source
2. **Map** to deets schema fields
3. **Diff** against current `me.toml` — show additions and changes
4. **Confirm** — user approves before writing (unless `--yes`)
5. **Write** via `store.SetValue` (preserves formatting)

### Sources

**`--git` (zero auth, local):**
- `git config user.name` → `identity.name`
- `git config user.email` → `contact.email`

**`--github` (requires `gh` CLI):**
- `gh api user` → JSON with login, name, email, bio, blog, company, location
- Maps to: `profiles.github.username`, `.name`, `.email`, `.url`
- Also populates `identity.name` and `contact.email` if empty

**`--orcid` (public API, no auth):**
- `GET https://pub.orcid.org/v3.0/{orcid}/person` (reads ORCID from `academic.orcid`)
- Maps to: `profiles.orcid.*`, `academic.institution`

### Constraints

- No secrets/tokens stored in me.toml
- No writes without confirmation (unless `--yes`)
- Requires `academic.orcid` to be set before `--orcid` works
- Requires `gh` CLI to be installed and authed for `--github`

---

## 4. Schema-Driven Skill

### New skill (~40 lines)

```markdown
---
name: deets
description: >
  Use when you need personal metadata about the user — name, email, ORCID,
  GitHub username, affiliations, education, or any other personal details. Also
  use when populating author fields, git identity, paper metadata, profile info,
  or personalized content.
---

# deets — Personal Metadata CLI

Query the user's personal metadata store. Fields are organized into core
categories (identity, contact, academic, education) and platform profiles
(profiles.github, profiles.pypi, profiles.orcid, etc.).

## Discovery

Run `deets schema --format json` to see all available fields with types,
descriptions, and example values. This is the authoritative source of what
data exists.

## Common Queries

    ```bash
    # Single value (bare output, pipe-friendly)
    deets get identity.name
    deets get contact.email
    deets get profiles.github.username

    # Platform context (all fields for a platform)
    deets get profiles.github
    deets get profiles.pypi

    # Cross-platform queries
    deets get profiles.*.email          # all platform emails
    deets get profiles.*.url            # all profile URLs

    # With fallback (never fails)
    deets get academic.scholar --default ""

    # Structured output
    deets show --format json            # full dump
    deets export --format env           # DEETS_IDENTITY_NAME="..." format
    ```

## Output Conventions

- Single `get`: bare value, no decoration (pipe-friendly)
- Multiple matches: JSON when piped, table on TTY
- `--format`: table, json, toml, yaml, env
- Exit code 2 = not found
```

### Why this works

- `deets schema --format json` is the live field inventory — always current
- Skill teaches query patterns, not specific fields
- 40 lines instead of 107 — less context window cost
- Skill stays stable as user adds profiles or custom categories

---

## 5. Expanded Descriptions

### Goal

Every field in `deets schema` output has a non-empty description. No blank descriptions.

### Approach

Expand `DefaultDescriptions` in `store/template.go` to cover all `profiles.*` sub-keys:

```go
"profiles.github": {
    "username": "GitHub username",
    "name":     "Display name on GitHub",
    "email":    "Email associated with GitHub account",
    "url":      "GitHub profile URL",
},
"profiles.pypi": {
    "username": "PyPI username",
    "name":     "Author name for PyPI packages",
    "email":    "Email for PyPI packages",
    "url":      "PyPI profile URL",
},
// ... etc for each platform
```

Also add missing descriptions for existing fields like `scholar`, `blog`, `r_universe`.

---

## 6. Updated Init Template

The `deets init` template gets restructured to match the new schema. All common fields pre-commented with descriptions:

```toml
# deets — Personal metadata
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
# name = "Display Name"
# email = "github@example.com"

[profiles.pypi]
# username = "your-username"

[profiles.orcid]
# id = "0000-0000-0000-0000"
```

The template prominently mentions `deets populate --all` as the fast path.

---

## Summary of Code Changes

| Area | Files | Scope |
|------|-------|-------|
| Parser (nested tables) | `internal/store/store.go` | ~15 lines changed in `LoadFile` |
| Validation (dotted categories) | `internal/store/writer.go` | Update `ValidateName` / add `ValidateCategoryName` |
| Populate command | `internal/commands/populate.go` (new) | ~200-300 lines |
| Populate sources | `internal/populate/` (new package) | ~150 lines per source |
| Skill file | `internal/commands/skill.md` | Rewrite (~40 lines) |
| Init template | `internal/store/template.go` | Rewrite template + expand `DefaultDescriptions` |
| Schema output | `internal/model/` | Ensure nested categories display as `profiles.X.Y` paths |
| Tests | Various `*_test.go` | New tests for populate, parser, nested schema |

## What Does NOT Change

- All existing commands (`get`, `set`, `rm`, `show`, `export`, etc.)
- Line-level TOML editing
- Local override system
- Output format system (table/json/toml/yaml/env)
- Exit code conventions
- `--format`, `--local`, `--quiet` flags
