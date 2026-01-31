# deets — Personal Metadata CLI

A self-describing, TOML-backed personal metadata store.

## Build & Test

```bash
go build -o deets ./cmd/deets
go test ./...
go test -cover ./...
go vet ./...
```

## Architecture

- `cmd/deets/main.go` — entrypoint, calls cobra Execute()
- `internal/commands/` — one file per CLI command, root.go has global flags
- `internal/config/` — path resolution (~/.deets/, local walk-up discovery)
- `internal/store/` — TOML read/write, merge logic, templates
- `internal/model/` — DB/Category/Field types, formatting, glob query

## Key Conventions

- Module: `github.com/queelius/deets`
- Data file: `~/.deets/me.toml` (global), `.deets/me.toml` (local override)
- `_desc` suffix fields are descriptions, filtered from normal output
- Line-level TOML editing preserves comments and formatting
- Exit codes: 0=success, 1=error, 2=key not found
- Output: bare value for single get, table on TTY, JSON when piped

## Dependencies

- `github.com/BurntSushi/toml` — TOML parsing
- `github.com/spf13/cobra` — CLI framework
