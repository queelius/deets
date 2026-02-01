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
internal/commands/            → one file per CLI command (get.go, set.go, etc.)
  root.go                    → rootCmd + global flags (--format, --local, --quiet)
  helpers.go                 → ExitError, parsePath(), loadDB(), targetFile()
internal/config/              → path resolution (~/.deets/ and local walk-up)
internal/model/               → DB/Category/Field types, Query(), Search(), formatting
internal/store/               → TOML Load/Write/Merge, line-level editing, templates
```

### Data flow

1. **Config** resolves paths: global `~/.deets/me.toml` + local `.deets/me.toml` (found by walking up from cwd, stops before $HOME)
2. **Store** loads both TOML files into `model.DB`, then merges: local fields override matching global fields per-category, non-overlapping fields from both are preserved
3. **Model** provides `Query(pattern)` with glob support (`identity.*`, `*.orcid`, `web.git*`) and `Search(query)` for case-insensitive text search across keys, values, and descriptions
4. **Commands** call `loadDB()` to get the merged DB, then format output (table on TTY, JSON when piped)

### Adding a CLI command

Each command lives in its own file under `internal/commands/`. Register via `init()`:

```go
func init() { rootCmd.AddCommand(myCmd) }
var myCmd = &cobra.Command{...}
```

Use `loadDB()` and `targetFile()` from `helpers.go` for read/write operations. Use `resolveFormat()` and respect `flagLocal`/`flagQuiet` global flags.

## Key Conventions

- **`_desc` suffix**: Fields like `orcid_desc` hold descriptions and are automatically excluded from query results, show output, and all format functions. Use `model.IsDescKey()` to check.
- **Line-level TOML editing** (`store/writer.go`): `SetValue`/`RemoveValue`/`RemoveCategory` edit TOML line-by-line to preserve comments and formatting. Never rewrite the entire file through marshal/unmarshal for mutations.
- **Exit codes**: 0=success, 1=error, 2=key not found
- **Output heuristic**: `get` prints bare value only for single exact-match results (no globs, format is `table`). Multiple matches → table on TTY, JSON when piped. The `resolveFormat()` function in `root.go` drives format selection.
- **Ordered output**: `model.DB` keeps categories and fields sorted alphabetically. JSON export uses a custom `orderedMap` type to preserve key order.
- **Template defaults** (`store/template.go`): `DefaultDescriptions` map provides fallback descriptions when no explicit `_desc` field exists.
