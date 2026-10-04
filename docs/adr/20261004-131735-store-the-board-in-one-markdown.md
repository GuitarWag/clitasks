---
id: 20261004-131735-store-the-board-in-one-markdown
title: Store the board in one Markdown file
status: accepted
date: 2026-05-16
branch: main
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/storage/**
  - internal/model/**
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go]
journal_refs: []
tags: [storage]
---

## Context

Backfilled. The TypeScript version already used Markdown, and the Go port kept the format.

The tool is for humans and AI agents. Both must be able to read and edit the board
without the CLI, and the board must give a readable diff in git.

## Decision

Store the whole board in one Markdown file, `tasks.md` by default. Use no database.

- One `##` section for each status. One list line for each task:
  `- [ ] [<id>] **<title>**` followed by `` `priority:` ``, `` `assignee:` ``, `` `tags:` ``
  and `` `due:` `` fields.
- Put timestamps on a `> Created: ... | Updated: ...` line under the task.
- The parser also reads the legacy format, which has `Created` and `Updated` on two lines.
- The parser accepts a metadata line only if it starts with an RFC 3339 date (`\d{4}-\d{2}-\d{2}T`).
  A description that contains the words "Created:" or "Updated:" stays part of the description.
- A missing file reads as an empty default board.

## Alternatives considered

- SQLite or JSON. These are more reliable to parse, but people cannot read or edit them
  easily, and they give poor diffs.

## Consequences

- Every write reads and renders the whole file. This is acceptable for a personal board.
- A hand edit that breaks the line format can drop a task without an error.
- A change to the format must keep the parser able to read files from earlier versions.
