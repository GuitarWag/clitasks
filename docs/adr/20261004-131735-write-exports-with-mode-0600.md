---
id: 20261004-131735-write-exports-with-mode-0600
title: Write exports with mode 0600
status: accepted
date: 2026-05-17
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/cli/export.go
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go]
journal_refs: []
tags: [security]
---

## Context

Backfilled from commit d5b66d4 (PR review). `tasks export --output <file>` writes JSON,
CSV or a summary. The output can contain assignee names.

## Decision

`tasks export` writes the output file with mode `0600`.

## Alternatives considered

- Use `0644`, the default for most tools. This is simpler, but other local users can then read assignee names.

## Consequences

- Other users on the same machine cannot read an export.
- To share an export, the user must change its mode.
- This mode is different from `tasks.md`, which keeps its existing mode
  (see the atomic write record).
