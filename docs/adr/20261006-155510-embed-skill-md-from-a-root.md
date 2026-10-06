---
id: 20261006-155510-embed-skill-md-from-a-root
title: Embed SKILL.md from a root package and use the v3 module path
status: accepted
date: 2026-10-06
branch: fix/go-install
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - skill.go
  - go.mod
  - Makefile
  - internal/cli/skill.go
supersedes: [20261004-131735-embed-skill-md-from-a-canonical]
superseded_by: []
depends_on: []
journal_refs: []
tags: [build, release]
---

## Context

After the `v3.0.0` tag, `go install github.com/GuitarWag/clitasks/cmd/tasks@v3.0.0`
failed for two reasons:

- A module at major version 2 or higher must end its path in `/vN`. The module
  path was `github.com/GuitarWag/clitasks`, so Go rejected the v3 tag.
- The binary embedded `internal/cli/SKILL.md`, a copy that `make sync-skill`
  made and `.gitignore` excluded. A remote checkout has no copy, so the embed
  failed. `go install ...@latest` never worked since the Go port.

## Decision

- The module path is `github.com/GuitarWag/clitasks/v3`.
- A root package, `clitasks` (`skill.go`), embeds the root `SKILL.md`. A
  package can embed a file in its own directory, so no copy is needed.
  `internal/cli` imports `clitasks.Skill`.
- `make sync-skill`, the copied file and its `.gitignore` entry are removed.
- `cmd/tasks` reports the module version from the build info when `make` did
  not set it through `-ldflags`.

## Alternatives considered

- Commit the copy in `internal/cli/`. Two copies drift apart.
- Move `SKILL.md` into `internal/cli/`. The root file is where people and the
  `tasks claude` docs expect it.

## Consequences

- `go install github.com/GuitarWag/clitasks/v3/cmd/tasks@latest` works.
- `go build` and `go test ./...` work on a clean checkout with no make step.
- The `v3.0.0` tag stays broken for `go install`; `v3.0.1` is the first
  installable release.
