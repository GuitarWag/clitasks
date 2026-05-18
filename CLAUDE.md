# CLITasks

CLI task management tool with Markdown storage, usable by humans and AI agents.

## Tech Stack

- Go (1.22+)
- `spf13/cobra` for CLI
- `charmbracelet/bubbletea` + `bubbles` + `lipgloss` for the TUI
- Markdown files as storage (no database)

## Build & Dev

```bash
make build         # build bin/tasks (runs `make sync-skill` first)
make test          # run all tests with -race
make lint          # golangci-lint
make install       # go install ./cmd/tasks
make run ARGS="list"   # go run ./cmd/tasks list

go run ./cmd/tasks tui     # launch the TUI directly
```

## Project Structure

- `cmd/tasks/` — binary entry point (`main.go`)
- `internal/model/` — Task/Board types and enums
- `internal/storage/` — Markdown parser, renderer, atomic writer
- `internal/board/` — Board service over a `Store`
- `internal/export/` — JSON / CSV / summary exports
- `internal/cli/` — Cobra commands; `SKILL.md` is embedded here via `go:embed`
- `internal/tui/` — Bubble Tea Model / Update / View and modals

The canonical `SKILL.md` lives at the repo root. `make sync-skill` (and every `make build`/`test`) copies it into `internal/cli/SKILL.md` for embedding.

## Git Rules

- **Never add `Co-Authored-By` trailers to commit messages.**
- Write concise commit messages focused on the "why."
