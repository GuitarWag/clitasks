---
id: 20261004-132518-compute-one-timeline-layout-for-cli
title: Compute one timeline layout for CLI, TUI and Mermaid
status: proposed
date: 2026-10-04
branch: chore/adrlog
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/timeline/**
  - internal/export/**
  - internal/tui/**
supersedes: []
superseded_by: []
depends_on: [20261004-132518-add-start-and-after-fields-to]
journal_refs: []
tags: [timeline, architecture]
---

## Context

The timeline appears in three places: `tasks timeline`, a TUI view and
`tasks export --format gantt` (Mermaid). If each place computes bar dates itself,
the three outputs will differ.

## Decision

A new package, `internal/timeline`, computes the layout. The package has no I/O.
`Build(tasks []model.Task, today time.Time) Layout` returns:

- One bar for each task that has a date, sorted by start date, then by ID.
- Bar start: `Start`, else the date of `CreatedAt`.
- Bar end: `DueDate`. A task in `done` with no due date ends on the date of
  `UpdatedAt`. Other tasks with no due date end on `today` and have `Open: true`.
- `Unscheduled`: tasks whose `DueDate` is not a `YYYY-MM-DD` date.
- `Warnings`: each `after` link where the task starts before the
  end of the task that it waits for.
- `From` and `To`: the date range of all bars.

The `after` links do not move bars. The tool shows a dependency conflict as a
warning and does not change the schedule.

The renderers:

- ASCII (`internal/timeline/ascii.go`). The CLI and the TUI both use it, with a width argument.
- Mermaid (`internal/export`, format `gantt`). It writes explicit dates and
  groups the bars by status in sections. It does not write Mermaid `after` syntax, because the
  dates already include the order and Mermaid draws no arrows.

## Alternatives considered

- Move the start of each task to the end of the tasks it waits for. This hides
  the real planned dates, and the user cannot see the conflict.
- Let each surface compute its own bars. This is less code at first, but the outputs drift apart.

## Consequences

- Tests for dates and ordering go in one package and use no terminal.
- The TUI and the CLI show the same bars for the same width.
