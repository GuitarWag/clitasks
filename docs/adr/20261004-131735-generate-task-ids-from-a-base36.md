---
id: 20261004-131735-generate-task-ids-from-a-base36
title: Generate task IDs from a base36 timestamp and a random suffix
status: accepted
date: 2026-05-17
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/board/id.go
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go]
journal_refs: []
tags: [storage]
---

## Context

Backfilled from commits 9faeda4 and d5b66d4. Users and agents type task IDs in commands,
so the IDs must be short. IDs must also sort by creation time and not collide on one board.

## Decision

`newID` in `internal/board/id.go` returns `T-<ms>-<sfx>`:

- `<ms>` is the Unix time in milliseconds, in uppercase base36.
- `<sfx>` is three characters from `0-9A-Z`, from `math/rand/v2` (PCG).

## Alternatives considered

- UUIDs. These are unique but too long to type.
- Sequential integers. These need a counter in the file, and they collide when two
  branches add tasks and then merge.

## Consequences

- The IDs are about 14 characters and sort by creation time.
- Two IDs collide only if both are created in the same millisecond with the same suffix
  (1 in 46,656). The RNG is not cryptographic, which is fine because IDs are not secrets.
