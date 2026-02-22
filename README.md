# deets

A self-describing, TOML-backed personal metadata store. Unix-philosophy CLI tool for making personal details (name, email, ORCID, GitHub, affiliations, etc.) instantly available to coding agents and scripts.

## Install

```bash
go install github.com/queelius/deets/cmd/deets@latest
```

Or build from source:

```bash
git clone https://github.com/queelius/deets
cd deets
go build -o deets ./cmd/deets
```

## Quick Start

```bash
# Initialize your metadata file
deets init

# Set some values
deets set identity.name "Alexander Towell"
deets set identity.aka '["Alex Towell"]'
deets set platforms.github.url "https://github.com/queelius"
deets set academic.orcid "0000-0001-6443-9897"

# Or auto-populate from external sources
deets populate --git             # harvest from git config
deets populate --github          # harvest from GitHub API (requires gh CLI)
deets populate --all             # all available sources

# Get values (great for scripts)
deets get identity.name               # → Alexander Towell
deets get platforms.github.url        # → https://github.com/queelius
deets get platforms.github             # → all GitHub platform fields
deets get platforms.*.url              # → all platform URLs
name=$(deets get identity.name)       # pipe-friendly bare output
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--format <fmt>` | Output format: `table`, `json`, `toml`, `yaml`, `env` |
| `--local` | Operate on local `.deets/me.toml` instead of global |
| `--quiet` / `-q` | Suppress informational messages |

When `--format` is not set, output defaults to `table` on a TTY and `json` when piped.

## Usage

### Get

```bash
deets get identity.name               # single value, bare output
deets get platforms.github             # all fields in a dotted category
deets get platforms.github.url        # field in a dotted category
deets get platforms.*.url             # cross-platform glob
deets get *.orcid                     # find key across all categories
deets get identity.name --desc        # include field description
deets get foo.bar --default x         # return "x" if not found
deets get foo.bar --exists            # exit 0 if found, 2 if not (no output)
```

Single exact matches output bare values (pipe-friendly). Multiple matches show a table on TTY, JSON when piped.

Glob patterns supported:
- `category` or `category.*` — all fields in a category (e.g., `platforms.github`)
- `platforms.*` — all fields across dotted sub-categories
- `platforms.*.url` — specific key across dotted sub-categories
- `*.key` — find a key across all categories (e.g., `*.orcid`)

### Show

```bash
deets show                       # table of all categories
deets show identity              # single category
deets show --format json         # full JSON dump
deets show --format toml         # raw merged TOML
deets show --format yaml         # YAML output
```

### Set / Remove

```bash
deets set identity.name "Alex Towell"
deets set cooking.fav "lasagna"  # creates [cooking] automatically
echo "piped" | deets set identity.name    # value from stdin
cat bio.txt | deets set identity.bio -    # explicit stdin with "-"
deets rm contact.phone           # remove a field
deets rm cooking                 # remove entire category
deets rm platforms.github         # remove a dotted category
deets rm platforms.github.bio    # remove field in dotted category
```

### Search

```bash
deets search "towell"            # search keys, values, and descriptions
```

### Describe

```bash
deets describe                   # all descriptions
deets describe identity          # descriptions in category
deets describe academic.orcid    # single field description
deets describe platforms.mastodon.handle "Mastodon handle"  # set a description
```

### Keys

```bash
deets keys                       # list all field paths, one per line
deets keys --format json         # as a JSON array
```

### Export

```bash
deets export                     # JSON (default, even on TTY)
deets export --format env        # DEETS_IDENTITY_NAME="..." format
deets export --format toml       # raw merged TOML
deets export --format yaml       # YAML
```

### Import

```bash
deets import backup.toml             # import into global store
deets import other.toml --local      # import into local store
deets import other.toml --dry-run    # preview changes without writing
```

### Diff

```bash
deets diff                       # compare local vs global (table)
deets diff --format json         # JSON output
```

### Schema

```bash
deets schema                     # show field types and metadata
deets schema --format json       # JSON output
```

### Populate

```bash
deets populate --git             # harvest name/email from git config
deets populate --github          # harvest from GitHub API (requires gh CLI)
deets populate --orcid           # harvest from ORCID API (requires academic.orcid)
deets populate --all             # all available sources
deets populate --github --dry-run  # preview without writing
deets populate --all --yes       # skip confirmation prompt
```

### Other

```bash
deets edit                       # open ~/.deets/me.toml in $EDITOR
deets edit --local               # open local override
deets which                      # show resolved paths, merge status
deets categories                 # list category names
deets version                    # print version
deets completion bash            # shell completions
```

## Data Format

### `~/.deets/me.toml`

```toml
[identity]
name = "Alexander Towell"
aka = ["Alex Towell", "queelius"]
pronouns = "he/him"

[contact]
email = "lex@metafunctor.com"
emails = ["lex@metafunctor.com", "queelius@gmail.com"]

[academic]
orcid = "0000-0001-6443-9897"
institution = "Southern Illinois University Edwardsville"
research_interests = ["information retrieval", "Bayesian statistics"]
degrees = ["MS Computer Science (SIUE, 2015)", "MS Mathematics (SIUE, 2023)"]

[web]
handle = "queelius"
domains = ["metafunctor.com"]

[platforms.github]
url = "https://github.com/queelius"
name = "Alex Towell"

[platforms.pypi]
url = "https://pypi.org/user/queelius"
```

Any `[category]` with any `key = "value"` is valid. Platform entries use `[platforms.name]` sections, enabling cross-platform queries like `deets get platforms.*.url`. Shared identity info (handle, email, name) lives in top-level categories to avoid redundancy.

### Local Overrides

Create `.deets/me.toml` in any project directory to override global fields:

```bash
deets init --local
deets set --local contact.email "project@example.com"
```

Local keys replace matching global keys within categories. Discovery walks up from cwd.

### Merge Behavior

When a local `.deets/me.toml` is present, fields are merged at key-level granularity:

- Local fields **override** matching global fields within the same category
- Non-overlapping fields from both files are **preserved**
- Categories only in global → included as-is; categories only in local → included as-is

```
# ~/.deets/me.toml (global)      # .deets/me.toml (local)
[identity]                        [identity]
name = "Alex Towell"             email_work = "alex@corp.com"
pronouns = "he/him"
                                  [contact]
[contact]                         email = "project@example.com"
email = "alex@example.com"

# Merged result:
# identity.name        → "Alex Towell"       (global, no override)
# identity.pronouns    → "he/him"            (global, no override)
# identity.email_work  → "alex@corp.com"     (local addition)
# contact.email        → "project@example.com" (local overrides global)
```

## Claude Code Integration

Install the deets skill so Claude Code knows how to query your metadata:

```bash
deets claude install             # install to ~/.claude/skills/
deets claude install --local     # install to .claude/skills/
deets claude uninstall           # remove the skill
```

## Exit Codes

- `0` — success
- `1` — error
- `2` — key/field not found
