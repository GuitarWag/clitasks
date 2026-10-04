---
id: 20261004-132518-add-start-and-after-fields-to
title: Add start and after fields to the task line
status: accepted
date: 2026-10-04
branch: chore/adrlog
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/model/**
  - internal/storage/**
  - internal/board/**
supersedes: []
superseded_by: []
depends_on: [20261004-131735-store-the-board-in-one-markdown]
journal_refs: []
tags: [storage, timeline]
---

## Context

The timeline feature needs a start and an end date for each task. It also needs
the order between tasks. The task model has none of these:

- `DueDate` is a free-form string. `add --due` says `YYYY-MM-DD` in its help text,
  but no code checks it.
- No field holds a start date. `CreatedAt` is the date of the record, not the date
  the work started.
- No field links one task to another.

## Decision

Add two optional fields to `model.Task` and to the task line:

- `Start string`, rendered as `` `start:YYYY-MM-DD` ``.
- `After []string`, rendered as `` `after:T-AAA-111,T-BBB-222` ``.

Rules:

- `--start` and `--due` accept only `YYYY-MM-DD`. The CLI rejects other values.
  The parser still reads a non-date `due:` from older files and keeps it.
- `tasks start <id>`, and every other move to in-progress, sets `Start` to today
  (local date) if `Start` is empty. An overdue task keeps an empty `Start`,
  because today is after its due date.
- `After` holds IDs of tasks on the same board. `Add` and `Update` reject an
  unknown ID, a self reference and a cycle.
- `Delete` removes the deleted ID from the `After` list of every other task.
- If `Start` and `DueDate` are both set, `Start` must not be after `DueDate`.

## Alternatives considered

- Use `CreatedAt` as the start. This needs no format change, but a task that waited in
  todo for weeks shows a long bar for the wait time.
- Record the date of each status change. This holds more data and changes the
  format more, and existing tasks have no history.
- Use one free-form `deps:` field and no validation. A bad ID then shows up only
  in the chart.

## Consequences

- Files written by this version are still readable by earlier versions. Earlier parsers
  ignore unknown backtick fields, but they drop `start` and `after` when they write the file.
- A board with an old free-form due date still loads. That task shows in the
  timeline as unscheduled.
- `Delete` can now change other tasks. The CLI prints the IDs that it changed.
