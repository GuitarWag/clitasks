---
id: 20261004-131728-port-clitasks-from-typescript-to-go
title: Port clitasks from TypeScript to Go
status: accepted
date: 2026-05-16
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - cmd/**
  - internal/**
  - go.mod
supersedes: []
superseded_by: []
depends_on: []
journal_refs: []
tags: [architecture, language]
---

## Context

Backfilled from commit 9faeda4 (2026-05-16) and PR #1.

The first versions (1.x and 2.x) were TypeScript on Node. The TUI used `blessed`.
The CHANGELOG for 2.0.1 to 2.1.0 records repeated fixes for `blessed` input
widgets that did not show or accept values. Users installed the tool through npm.

## Decision

Rewrite the tool in Go as one static binary:

- `spf13/cobra` for the CLI commands.
- `charmbracelet/bubbletea`, `bubbles` and `lipgloss` for the TUI.
- Shared styles in `internal/theme`, used by the CLI and the TUI.

Keep behavior the same as the TypeScript version for every command and TUI key, with two
intentional changes:

- `export --format` has no short flag, because the root command uses `-f` for `--file`.
- In the TUI, `h` moves left only. `?` is the only help key.

## Alternatives considered

- Keep TypeScript and replace `blessed` with a different TUI library. This keeps the Node
  install and the startup cost.

## Consequences

- Users install with `go install`, and the binary starts faster.
- The repo has one language. The Node toolchain and `package.json` are gone.
- Scripts that used `export -f <format>` or `h` for help must change.
