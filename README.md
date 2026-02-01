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
deets set web.github "queelius"
deets set academic.orcid "0000-0002-1234-5678"
deets describe academic.orcid "ORCID persistent digital identifier"

# Get values (great for scripts)
deets get identity.name          # → Alexander Towell
deets get web.github             # → queelius
name=$(deets get identity.name)  # pipe-friendly bare output
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
deets get identity.name          # single value, bare output
deets get academic               # all fields in category
deets get *.orcid                # find key across all categories
deets get identity.na*           # glob within category
deets get identity.name --desc   # include field description
deets get foo.bar --default x    # return "x" if not found
deets get foo.bar --exists       # exit 0 if found, 2 if not (no output)
```

Single exact matches output bare values (pipe-friendly). Multiple matches show a table on TTY, JSON when piped.

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
deets describe web.mastodon "Mastodon handle"  # set a description
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
name_desc = "Full legal name"
aka = ["Alex Towell"]
pronouns = "he/him"

[contact]
email = "alex@example.com"

[web]
github = "queelius"
blog = "https://example.com"

[academic]
orcid = "0000-0002-1234-5678"
orcid_desc = "ORCID persistent digital identifier"
research_interests = ["information retrieval", "Bayesian statistics"]
```

Any `[category]` with any `key = "value"` is valid. Add `_desc` suffix for self-describing fields.

### Local Overrides

Create `.deets/me.toml` in any project directory to override global fields:

```bash
deets init --local
deets set --local contact.email "project@example.com"
```

Local keys replace matching global keys within categories. Discovery walks up from cwd.

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
