# CLITasks

CLI task management tool with Markdown storage, usable by humans and AI agents.

## Tech Stack

- Go (1.26+)
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
- `internal/timeline/` — Gantt layout (bars, conflicts), no rendering
- `internal/ui/` — renderers shared by the TUI and CLI output: `theme` (presets, contrast-tested), `icons`, `card` (card, detail), `tlgrid` (timeline grid), `dates`
- `internal/tui/` — Bubble Tea app: `app.go` (model, writes, reload, frame), `board.go`, `timeline.go`, `overlay.go`, `form.go` (huh)

Review UI changes as images: `UI_PREVIEW_DIR=/tmp/p go test -run TestPreview ./internal/tui/ ./internal/ui/...` writes ANSI frames; render them with `charmbracelet/freeze`.

The canonical `SKILL.md` lives at the repo root. `make sync-skill` (and every `make build`/`test`) copies it into `internal/cli/SKILL.md` for embedding.

## Git Rules

- **Never add `Co-Authored-By` trailers to commit messages.**
- Write concise commit messages focused on the "why."

## Decisions (adrlog)

Record design decisions with the `adrlog` CLI. Records go to `docs/adr/`, state to `.adrlog/`.

- Before a change in an area, run `adrlog list --affects '<glob>'` and read the matching records.
- After a decision that changes behavior, storage format, CLI surface, or a dependency, run `adrlog new "<title>" --status accepted --affects '<glob>' --tags <tag>`, then fill Context / Decision / Consequences in the new file.
- To replace a decision, pass `--supersedes <id>`. Do not edit the old record's decision.
- When the stop hook asks about a change that holds no decision, run `adrlog ack --none`.
- Run `adrlog lint` before you commit a record.
