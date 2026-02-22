# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
go build -o deets ./cmd/deets
go test ./...
go test -cover ./...
go vet ./...
go test ./internal/model/ -run TestQuery   # run a single test by name
```

## Module & Dependencies

- Module: `github.com/queelius/deets` (Go 1.22.2)
- `github.com/BurntSushi/toml` — TOML parsing
- `github.com/spf13/cobra` — CLI framework

## Architecture

```
cmd/deets/main.go            → minimal entrypoint, calls commands.Execute()
internal/commands/            → one file per CLI command (get.go, set.go, populate.go, etc.)
  root.go                    → rootCmd + global flags (--format, --local, --quiet)
  helpers.go                 → ExitError, parsePath(), loadDB(), targetFile()
  populate.go                → deets populate --git/--github/--orcid
internal/config/              → path resolution (~/.deets/ and local walk-up)
internal/model/               → DB/Category/Field types, Query(), Search(), formatting
internal/store/               → TOML Load/Write/Merge, line-level editing, templates
```

### Data flow

1. **Config** resolves paths: global `~/.deets/me.toml` + local `.deets/me.toml` (found by walking up from cwd, stops before $HOME)
2. **Store** loads both TOML files into `model.DB`, then merges: local fields override matching global fields per-category, non-overlapping fields from both are preserved
3. **Model** provides `Query(pattern)` with glob support (`platforms.github`, `platforms.*`, `platforms.*.url`, `*.orcid`) and `Search(query)` for case-insensitive text search across keys, values, and descriptions. Query tries the full pattern as a category name/glob first, then splits on the last dot for field-level matching.
4. **Commands** call `loadDB()` to get the merged DB, then format output (table on TTY, JSON when piped)

### Dotted category names

Category names can contain dots (e.g., `platforms.github`). In TOML, `[platforms.github]` produces nested tables which `LoadFile` flattens into dotted category names. Path resolution splits on the **last** dot: `platforms.github.url` → category `platforms.github`, key `url`.

- `ValidateName(name)` — rejects dots (bare TOML keys only)
- `ValidateCategoryName(name)` — allows dots between valid bare-key segments
- `parsePath(path)` — splits on last dot, validates both parts
- `HasCategory(filePath, name)` — checks if a `[name]` section exists in the TOML file

### Adding a CLI command

Each command lives in its own file under `internal/commands/`. Register via `init()`:

```go
func init() { rootCmd.AddCommand(myCmd) }
var myCmd = &cobra.Command{...}
```

Use `loadDB()` and `targetFile()` from `helpers.go` for read/write operations. Use `resolveFormat()` and respect `flagLocal`/`flagQuiet` global flags.

### Adding a populate source

Populate sources follow a parse/exec separation pattern for testability:

1. An unexported `populateXxx()` function handles exec (CLI calls, HTTP requests) and returns `[]populateEntry`
2. A `parseXxxResponse(data []byte)` function handles pure parsing of the response into entries
3. Add a `--xxx` flag in `populate.go` and wire it into the `RunE` handler and the `--all` block

This lets unit tests call `parseXxxResponse()` with canned JSON without needing real external tools.

## Testing

Test helpers live in `testhelper_test.go` (commands package):

- **`setupTestEnv(t)`** — creates a temp HOME, sets `$HOME`, chdir into it, and **resets all global flags** to defaults. Every command test must call this.
- **`setupTestDB(t)`** — calls `setupTestEnv` then writes a sample `~/.deets/me.toml` with identity, contact, web, academic, and platforms.github categories.
- **`executeCommand(args...)`** — runs the cobra root command, captures real `os.Stdout`/`os.Stderr` via pipe redirection (not `cmd.OutOrStdout()`), returns `(stdout, stderr, error)`.

**Critical**: cobra flags are package-level vars (`flagFormat`, `flagLocal`, etc.). `setupTestEnv` resets them all. If you add a new flag, you must add its reset to `setupTestEnv` or tests will leak state between runs.

Populate tests that need `git config` write a `.gitconfig` in the temp HOME for isolation.

## Key Conventions

- **`_desc` suffix**: Fields like `orcid_desc` hold descriptions and are automatically excluded from query results, show output, and all format functions. Use `model.IsDescKey()` to check.
- **Line-level TOML editing** (`store/writer.go`): `SetValue`/`RemoveValue`/`RemoveCategory` edit TOML line-by-line to preserve comments and formatting. Never rewrite the entire file through marshal/unmarshal for mutations.
- **Exit codes**: 0=success, 1=error, 2=key not found. Commands return `*ExitError` which `main.go` unwraps to set the process exit code.
- **Output heuristic**: `get` prints bare value only for single exact-match results (no globs, format is `table`). Multiple matches → table on TTY, JSON when piped. The `resolveFormat()` function in `root.go` drives format selection.
- **Ordered output**: `model.DB` keeps categories and fields sorted alphabetically. JSON export uses a custom `orderedMap` type to preserve key order.
- **Template defaults** (`store/template.go`): `DefaultDescriptions` map provides fallback descriptions when no explicit `_desc` field exists.
