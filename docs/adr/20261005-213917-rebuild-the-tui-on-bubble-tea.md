---
id: 20261005-213917-rebuild-the-tui-on-bubble-tea
title: Rebuild the TUI on Bubble Tea v2 with a themed component layer
status: proposed
date: 2026-10-05
branch: feat/tui-redesign
worktree: clitasks
session: 386254d3-bd79-4037-9422-519eabacf631
affects:
  - internal/tui/**
  - internal/ui/**
  - internal/theme/**
  - internal/timeline/**
  - internal/cli/**
supersedes: []
superseded_by: []
depends_on: [20261004-131728-port-clitasks-from-typescript-to-go, 20261004-132518-compute-one-timeline-layout-for-cli]
journal_refs: []
tags: [tui, architecture]
---

## Context

The TUI works, but it looks unfinished, and the timeline view is hard to read:

- All colors are hard-coded ANSI 16-color values in `internal/theme`, so the TUI does not match the terminal's light or dark theme.
- Modals replace the whole screen, so the user loses sight of the board while they edit.
- The timeline renders plain strings, so it cannot color a single cell, shade weekends, or show today behind a bar.
- The add and edit form has no date picker or `after` picker, and it shows errors only after the last step.
- There is no mouse support, no detail pane, and no reload when an agent changes `tasks.md`.

Bubble Tea v2, Lip Gloss v2, Bubbles v2 and Huh v2 are stable. Lip Gloss v2 adds layers with a compositor and hit-testing. Bubble Tea v2 reports the terminal background color and has a declarative `View`.

## Decision

Rebuild the TUI on Bubble Tea v2. Keep Go and the existing `board`, `storage` and `timeline.Build` packages.

- A new package `internal/ui` holds the parts that do not depend on Bubble Tea:
  - semantic theme tokens with four presets (Catppuccin Mocha, Catppuccin Latte, Tokyo Night, Nord);
  - icon sets (Nerd Font, Unicode, ASCII);
  - the card and the timeline grid renderers.
- The CLI's own output (`board`, `list`, `show`, `timeline`) uses the same `internal/ui` renderers, so the CLI and the TUI look the same.
- The TUI chooses a dark or light preset from `tea.BackgroundColorMsg`. `TASKS_THEME`, `TASKS_ICONS` and `tasks tui --theme/--icons` override the choice. `NO_COLOR` is respected.
- Overlays (form, picker, help, palette) are Lip Gloss layers over a dimmed copy of the screen. Mouse hit-testing uses `Compositor.Hit` with layer IDs.
- Forms use Huh v2 inside an overlay. Each field validates as the user types.
- The timeline renderer returns cells with a semantic kind (bar, open end, today, weekend, conflict). The theme maps each kind to a style. `timeline.Build` stays as it is.

## Alternatives considered

- **ratatui (Rust).** This needs a second language or a full port, and ratatui has no Gantt widget. The problems are in our renderer, not in the framework.
- **Keep Bubble Tea v1 and restyle.** v1 has no layer compositor and no background color detection. Floating overlays and adaptive themes would need code that v2 already ships.
- **Hand-written forms.** We already have hand-written forms. A date field, a multi-select `after` picker and inline validation are what Huh provides.

## Consequences

- New direct dependencies: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`, `charm.land/bubbles/v2` and `charm.land/huh/v2`. The v1 Charm modules are removed after the migration.
- `internal/theme` is replaced by `internal/ui/theme`.
- `timeline.RenderASCII` is replaced by the cell grid. Plain-text output stays available with the ASCII icons and no color.
- TUI tests move to golden files rendered with the ASCII color profile, plus `teatest/v2` flow tests.
- Wrong Nerd Font glyphs show as boxes, so the icon default must be careful. See the plan.
