---
id: 20261004-131735-resolve-the-board-file-from-flag
title: Resolve the board file from flag, then env var, then tasks.md
status: accepted
date: 2026-05-17
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/cli/root.go
  - internal/cli/export.go
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go]
journal_refs: []
tags: [cli]
---

## Context

Backfilled from commits 9faeda4 and d5b66d4. Agents and scripts must point the CLI at a
board in another directory without a `cd`.

## Decision

`resolveFilePath` in `internal/cli/root.go` selects the board file in this order:

1. The persistent `--file` / `-f` flag.
2. The `TASK_BOARD_FILE` environment variable.
3. `tasks.md` in the current directory.

The code reads the flag from `cmd.Flags()`, not from a package-level variable. Because
the root command owns `-f`, no subcommand can use `-f` for a different flag.

## Alternatives considered

- A config file, for example `~/.config/clitasks`. This adds a file format to maintain, and the flag and env var already cover agents and scripts.

## Consequences

- An agent session can set `TASK_BOARD_FILE` one time for all of its commands.
- Tests do not need to reset a global value between runs.
- `export --format` has no short form.
