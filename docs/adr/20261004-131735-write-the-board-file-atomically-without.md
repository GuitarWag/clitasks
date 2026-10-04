---
id: 20261004-131735-write-the-board-file-atomically-without
title: Write the board file atomically without a directory fsync
status: accepted
date: 2026-05-17
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/storage/render.go
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go, 20261004-131735-store-the-board-in-one-markdown]
journal_refs: []
tags: [storage, durability]
---

## Context

Backfilled from commits 9faeda4 and d5b66d4. The TypeScript version wrote `tasks.md` in
place, so an interrupt during a write could leave a truncated board.

## Decision

`atomicWrite` in `internal/storage/render.go`:

1. Creates a temp file `.tasks-*.md` in the same directory.
2. Copies the permission bits of the existing file to the temp file.
3. Writes and fsyncs the temp file.
4. Renames the temp file over the target.

The parent directory is not fsynced.

## Alternatives considered

- Fsync the directory after the rename. This gives full durability on a crash but costs one
  more syscall and platform-specific code.

## Consequences

- An interrupt never leaves a partial `tasks.md`.
- A shared file, for example one that is group-readable, keeps its mode.
- A power loss after the rename and before the OS flushes the directory can lose the new file.
  This is acceptable for a personal CLI.
- The CLI has no lock. Two concurrent writers can lose an update (last write wins).
