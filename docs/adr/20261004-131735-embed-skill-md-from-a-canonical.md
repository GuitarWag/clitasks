---
id: 20261004-131735-embed-skill-md-from-a-canonical
title: Embed SKILL.md from a canonical root copy
status: accepted
date: 2026-05-17
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - SKILL.md
  - internal/cli/embed.go
  - internal/cli/skill.go
  - Makefile
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go]
journal_refs: []
tags: [agents, build]
---

## Context

Backfilled from commits 9faeda4 and d5b66d4. `tasks claude` and `tasks codex` install a
skill file that tells an agent how to use the CLI. The binary must work after
`go install` with no repo checkout, and `go:embed` cannot read a file outside its package
directory.

## Decision

- The canonical `SKILL.md` is at the repo root.
- `make sync-skill` copies it to `internal/cli/SKILL.md`, which `go:embed` reads.
  `build`, `install`, `run`, `test` and `cover` run `sync-skill` first.
- `internal/cli/SKILL.md` is in `.gitignore`.
- The install writes to `~/.<tool>/skills/tasks-cli/SKILL.md` (global) or
  `.<tool>/skills/tasks-cli/SKILL.md` (local). This path matches the path that other
  skills use, so the install does not overwrite them.

## Alternatives considered

- Commit the copy in `internal/cli/`. This removes the make step but lets the two files
  drift apart.

## Consequences

- A bare `go test ./...` or `go build` on a clean checkout fails until `make sync-skill`
  runs. The README documents this.
- Edit only the root `SKILL.md`.
