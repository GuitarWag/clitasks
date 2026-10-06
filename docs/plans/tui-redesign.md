# Plan: TUI and CLI UI redesign

Decision: [Rebuild the TUI on Bubble Tea v2 with a themed component layer](../adr/20261005-213917-rebuild-the-tui-on-bubble-tea.md) (`accepted`).

Status: done (2026-10-06). The build differs from this plan in these places:

- `internal/tui` is one package, not one subpackage per screen. The screens share the model and the geometry, and separate packages added interfaces with one user.
- `tab` switches screens. `1` to `4` jump to board columns, so `1`/`2` are not screen keys.
- Mouse hit-testing uses the board geometry function that the renderer also uses, not `Compositor.Hit`. Overlays and toasts use the compositor.
- The TUI tests drive the model directly (`Update` and `render`), not `teatest/v2`.
- The CLI prints rich output only on a terminal. Pipes keep the plain format with ASCII icons and IDs, because agents parse it. The CLI reads `COLORFGBG` for dark or light instead of querying the terminal, which could stall a command for 2 s.
- The Unicode icons are `○ ◔ ⊘ ●` and `⊙ ▸ ↪`: `◐ ◷ ↳` are missing from common monospace fonts. The Nerd glyphs come from nerd-fonts `glyphnames.json`.
- No VHS tapes. The README uses screenshots from the preview harness (`UI_PREVIEW_DIR`) rendered with `freeze`.
- Task text from `tasks.md` drops control characters at parse time (a security review finding during the work).

User choices (2026-10-05): Nerd Font icons with a fallback, preset themes chosen from the terminal background, kanban cards with a detail pane, and mouse support.

## Goals

1. The board and the timeline look like one designed product, in a dark terminal and in a light terminal.
2. Every action takes one key, and the status bar always shows the keys that work on the current screen.
3. The user never loses context. Overlays float over a dimmed board, and the detail pane follows the selection.
4. The timeline answers the important questions at a glance: what is late, what is next, and what waits on what.
5. The CLI's own output (`board`, `list`, `show`, `timeline`) uses the same renderers.
6. The UI is usable with no color, with no Nerd Font, at 80x24, and with only a keyboard.

Out of scope: a new storage format, multiple boards in one window, undo history, and dragging bars with the mouse.

## Design principles

- **Semantic color only.** No component uses a raw color. It uses a token such as `Accent`, `StatusDone` or `Surface`. Status and priority always show an icon and text as well as a color.
- **One focus.** One element has the accent border or the selection bar at any time.
- **Quiet chrome.** Borders use the `Overlay` tone, and only the focused element uses `Accent`. Use no box inside a box.
- **Truncate, never wrap.** Titles end in `…`. The full title is in the detail pane.
- **Stable layout.** Resizing the window or moving the selection never moves other elements by a half column.

## Visual spec

### Theme tokens (`internal/ui/theme`)

| Token | Use |
|---|---|
| `Base`, `Mantle`, `Crust` | Background layers: app, panels, status bar |
| `Surface0..2` | Cards, selected card, hover |
| `Overlay0..1` | Borders, dividers, weekend shading |
| `Text`, `Subtext`, `Muted` | Primary, secondary and disabled text |
| `Accent` | Focus border, selection bar, key hints |
| `StatusTodo/InProgress/Blocked/Done` | Column header, card stripe, timeline bar |
| `PrioCritical/High/Medium/Low` | Priority chip |
| `Success/Warn/Error/Info` | Toasts and validation |
| `Today` | Today column in the timeline, and a due-today chip |

Presets: Catppuccin Mocha (dark default), Catppuccin Latte (light default), Tokyo Night, and Nord. At startup the app asks for `tea.BackgroundColorMsg`, picks Mocha or Latte, and re-renders. `TASKS_THEME=<name>` or `--theme` overrides the choice. `NO_COLOR` gives a monochrome theme that uses bold, faint and reverse only.

### Icons (`internal/ui/icons`)

| Meaning | Nerd | Unicode | ASCII |
|---|---|---|---|
| todo / in progress / blocked / done | `󰄱 󰔟 󰅙 󰄵` | `○ ◐ ⊘ ●` | `[ ] [>] [!] [x]` |
| priority critical … low | `󰀦 󰁝 󰁔 󰁅` | `▲▲ ▲ ■ ▼` | `!! ! = .` |
| due / start / after / assignee / tag | `󰃭 󰐊 󰌷 󰀄 󰓹` | `◷ ▸ ↳ @ #` | `due start after @ #` |

`TASKS_ICONS=auto|nerd|unicode|ascii`. On `auto`, the app uses Nerd glyphs when `TERM_PROGRAM` is WezTerm or Ghostty. Both terminals ship Nerd Font symbols. Every other terminal gets Unicode glyphs, because a missing glyph shows as a box. Check this list against current terminal docs in slice 1.

### App shell

```
 ▌clitasks  Sprint 12 · 18 tasks                    [ Board ]  Timeline       ◷ 3 due this week
 ─────────────────────────────────────────────────────────────────────────────────────────────
  <screen body>

 ─────────────────────────────────────────────────────────────────────────────────────────────
  NORMAL  tasks.md  /backend                 ←→ column  ↑↓ task  a add  e edit  ? help  q quit
```

- The header shows the board name, the task count, tabs (`1` Board, `2` Timeline, or `tab`), and one alert, such as overdue tasks or after conflicts.
- The status bar shows the mode pill (`NORMAL`, `SEARCH`, `EDIT`), the file name, the active filter, and short help from `bubbles/help`. The short help comes from the keymap of the focused component.
- Toasts appear above the status bar on the right and fade after 3 seconds: "Saved", "Moved to Done", "Reloaded, tasks.md changed on disk".

### Board

Width ≥ 140: four columns and a detail pane on the right (36 columns).
Width 100–139: four columns and a detail pane at the bottom (8 rows).
Width < 100: two columns visible, scrolled with the selection, and `enter` opens the detail as an overlay.
Height < 24: compact cards (one line each).

```
 TODO 5            IN PROGRESS 3 ·2   BLOCKED 1          DONE 9          │ T-MUTT74DN-H6G
 ╭───────────────╮ ╭───────────────╮  ╭───────────────╮  ╭─────────────╮ │ Write the parser
 ▌Write the      │ │ Design schema │  │ Deploy to prod│  │ Setup repo  │ │
 ▌parser         │ │ ▲ high  ◷ 10-09│ │ ▲▲ crit       │  │ ● done      │ │ ◐ In progress  ▲ high
 ▌@bob #backend  │ │ @alice        │  │ ↳ T-…H6G      │  ╰─────────────╯ │ ▸ 2026-09-26 → ◷ 2026-10-09
 ╰───────────────╯ ╰───────────────╯  ╰───────────────╯                  │ ↳ waits for Design schema ✓
                                                                          │ @bob  #backend #parser
```

- A card has a status stripe on the left edge, a title of up to two lines, and one meta line (priority chip, due chip, assignee). The due chip is `Today` when the date is today, `Error` when it is overdue, and `Warn` when it is in less than 3 days.
- The selected card gets `Surface1` and an `Accent` border. Other cards have no border fill.
- The column header shows the count, plus `·n` when the column has tasks that are overdue or in conflict.
- Each column scrolls by itself (a viewport), with `↑ 3 more` / `↓ 2 more` markers.
- An empty column shows one muted line: "Nothing here, press a to add".
- Keys:
  - `h j k l` and the arrows move the selection, `1–4` jump to a column, and `g`/`G` go to the top or bottom.
  - `H`/`L` (shift) move the selected task to the column to the left or right.
  - `enter` toggles the detail pane, `space` opens the status picker, and `/` searches.
- Mouse: a click selects a card, a click on a column header focuses that column, the wheel scrolls the column under the pointer, and a double-click opens the editor.

### Detail pane

The detail pane shows these fields: title, ID, status and priority pills, dates with the relative time ("in 4 days", "2 days late"), the `after` tasks with their status icons, the dependents, tags, assignee, the description (wrapped, scrollable), and the created and updated times. The board and the timeline use the same pane.

### Timeline

```
                         Sep 2026                       │ Oct 2026
                         21 22 23 24 25 26 27 28 29 30  │ 01 02 03 04 05 06 07 08 09 10 11
 ● T-…WPC Design schema  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━   │                ┃
 ◐ T-…H6G Write parser             ⚠━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 ◐ T-…N77 Ship TUI view                                 │             ⚠━━╸░░
 ○ T-…X1A Docs                                          │                ┃    ━━━━━━━━━━━━━
```

The mockup above uses box glyphs to show the shape. The real bars are cells with a background color.

- **Grid renderer:** `internal/ui/timeline` turns a `timeline.Layout` into a grid of cells. Each cell has a kind: `Bar`, `BarOpen` (a fade made of `▓▒░`), `Today`, `Weekend`, `WeekBand`, `Conflict`, `DepHighlight` or `Empty`. The theme maps the kinds to styles. The selected row gets `Surface1` across the full width, and its bar keeps its status color.
- **Header:** two rows. The first row holds months or years, the second holds days or ISO week numbers. A `│` divider marks the start of each month.
- **Today:** a full-height `Today` column, drawn under the bars. On a bar cell it shows as a brighter tone of the bar color.
- **Zoom:** `z` cycles day → week → month. `+` and `-` also zoom. The scale is fixed while the user scrolls.
- **Navigation:**
  - `j k` select a row, `h l` scroll one column, and `{ }` scroll one screen.
  - `.` returns to today, and `f` fits the whole range in the screen.
- **Dependencies:** for the selected row, the bars of its `after` tasks and of its dependents get `DepHighlight`. The label column shows `↳` next to them. A conflict shows `⚠` at the bar start, and the detail pane says which link fails.
- **Grouping:** `b` ("group by") cycles the grouping: none, status, assignee, tag. Each group has a header row with a count.
- **Unscheduled tasks:** a collapsible section at the bottom lists them. The section says which field does not parse.
- **Edit dates:**
  - `<`/`>` move the selected bar one day, `alt+<`/`alt+>` change its due date, and `[`/`]` set the start or due date to today.
  - These keys write through `board.Update`, so the existing validation applies.
- **Mouse:** a click on a bar or a label selects that row, the wheel scrolls rows, and shift with the wheel scrolls horizontally.

### Overlays

All overlays are Lip Gloss layers, centered over a dimmed snapshot of the screen (`Muted` foreground).

- **Task form (Huh v2):**
  - The fields are title, description (text area), status, priority (select with icons), assignee (input with suggestions from existing assignees), tags (multi-select of existing tags plus a free-text field), start and due (input with validation and a shortcut: `today`, `+3d`, `fri`, `next week`), and after (multi-select of tasks with search; the form rejects cycles and shows the reason).
  - Errors show under the field as the user types.
  - `ctrl+s` saves from any field, and `esc` asks before it drops changes.
  - The date shortcuts are parsed by a small `ui/dates` helper with tests. The storage format stays `YYYY-MM-DD`.
- **Status picker:** four rows with icons, opened by `space`. The number keys `1–4` also pick a status.
- **Confirm:** a small dialog. A delete lists the dependents that it will unlink.
- **Search (`/`):** an inline bar in the status bar, not a modal.
  - The results update as the user types, and the matching text is highlighted in cards and labels.
  - The search accepts structured tokens: `@alice`, `#backend`, `!high`, `is:overdue`, `is:blocked`, `due:<7d`.
- **Command palette (`ctrl+k` or `:`):** a fuzzy list of every action with its key. This lets a user learn the keys and run rare actions: change theme, change icons, export Mermaid to the clipboard, reload.
- **Help (`?`):** the full keymap from `bubbles/help`, grouped by screen.

### Live reload

Every 2 seconds, the app checks the modification time of `tasks.md` with `tea.Tick` and `os.Stat`, so no new dependency is needed. When the file changes and no form is open, the app reloads the file, keeps the selection by task ID, and shows a toast. When a form is open, the app shows a warning toast and reloads after the form closes.

### CLI output

`tasks board`, `list`, `show`, `stats` and `timeline` use the same theme and the same renderers from `internal/ui`:
- `board` prints the four columns of compact cards.
- `show` prints the detail pane.
- `timeline` prints the cell grid.

When stdout is not a TTY, or when `NO_COLOR` is set, the output uses the ASCII icons and no color, so pipes and agents still get readable text.

### Accessibility

- No meaning depends on color alone. Status and priority always have an icon and a word or abbreviation.
- The contrast of every text token against `Base` must be at least 4.5:1, and a test computes it for every preset.
- The `ascii` icon set and `NO_COLOR` together give a fully plain UI.
- When the terminal reports that it lost focus, the UI dims (`ReportFocus`).

## Architecture

```
internal/ui/           no Bubble Tea imports; testable as pure functions
  theme/               tokens, presets, picker (background color → preset), contrast test
  icons/               sets and auto choice
  card/                card and detail pane renderers (string in, string out)
  tlgrid/              timeline.Layout → cell grid → styled string
  dates/               relative dates and date shortcuts
internal/tui/          Bubble Tea v2
  app.go               root model: screen routing, overlay stack, toasts, live reload
  shell.go             header, status bar, layout breakpoints
  board/               board screen (columns, viewports, mouse)
  timeline/            timeline screen (zoom, scroll, grouping, date edits)
  overlay/             form (huh), picker, confirm, palette, help
  keys.go              keymaps per screen, for bubbles/help and the palette
internal/timeline/     Build stays; RenderASCII is deleted after slice 6
internal/theme/        deleted after slice 7
```

The root model owns the `*board.Board`. Child screens get the data they need and return commands. They do not save the board themselves. All writes go through one `saveCmd`, which shows a toast on success and an error toast on failure. This also fixes the class of bug where the TUI dropped errors.

## Slices

Each slice is one PR-sized commit series. Each slice ends with `make test` and `make lint` green and the golden files updated.

1. **v2 migration, no visual change.** Move to `charm.land/bubbletea/v2`, `lipgloss/v2` and `bubbles/v2`: `KeyPressMsg`, `View()` returns `tea.View`, and `AltScreen`/`MouseMode` move to the view. All existing TUI tests pass. This separates the framework risk from the design work.
2. **`internal/ui` foundations.**
   - Theme tokens, the four presets, the background-color picker, and `NO_COLOR`.
   - The icon sets and the auto choice.
   - The `--theme` and `--icons` flags, and the `TASKS_THEME` and `TASKS_ICONS` env vars.
   - The contrast test.
3. **App shell.**
   - Header, tabs, status bar with the mode pill and contextual help, toasts, and the layout breakpoints.
   - An overlay stack on the Lip Gloss compositor with a dimmed background.
   - The `saveCmd` error path.
   - Golden tests at 80x24, 120x40 and 200x60.
4. **Board.**
   - Cards, column viewports, the detail pane in three layouts, compact mode, and empty states.
   - Shift moves, number jumps, and mouse support with `Compositor.Hit`.
5. **Overlays.**
   - The Huh form with the date shortcuts and the `after` picker.
   - The status picker, the confirm dialog, inline search with tokens and highlighting, the command palette, and help.
6. **Timeline.**
   - The `tlgrid` renderer, the two-row header, the today column, weekends and week bands, and zoom.
   - Grouping, dependency highlighting, the unscheduled section, the date edit keys, and mouse support.
   - Delete `RenderASCII`.
7. **CLI output and live reload.**
   - Port `board`, `list`, `show`, `stats` and `timeline` to the `internal/ui` renderers. Delete `internal/theme`.
   - Add the mtime polling.
8. **Polish and docs.**
   - Set the window title, add focus dimming, and benchmark render time with 500 tasks (target: under 8 ms for each frame on an M-series Mac).
   - Record VHS tapes (`docs/tapes/*.tape`, which produce GIFs in the README) for the board, the timeline and the form, in a dark and a light theme.
   - Update README, TUI_GUIDE, SKILL.md and CHANGELOG. Move the ADR to `accepted`.

## Keymap rule

A key does the same thing on every screen, or it is used on one screen only. Shared keys:
`a` add, `e` edit, `d` delete, `space` status, `/` search, `ctrl+k` palette, `?` help, `1`/`2` or `tab` screens, `q` quit, `esc` close or clear.
Board only: `H`/`L` move a task between columns, `g`/`G` top or bottom.
Timeline only: `{`/`}`, `<`/`>`, `[`/`]`, `z`, `.`, `f`, `b`.
A test checks the keymaps for collisions.

## Tests

- **`internal/ui`:** table tests and golden strings with the ASCII color profile. One golden file for each preset uses the TrueColor profile, to catch token mistakes.
- **`internal/tui`:** `teatest/v2` flows: add a task, edit its dates, move it with `H`/`L`, search, delete a task that has dependents, zoom and scroll the timeline, and a mouse click on a card and on a bar.
- **Resize matrix:** render every screen at 80x24, 100x30, 140x40 and 220x60 and check that no line is wider than the window.
- **Time zones:** run the timeline and the date-shortcut tests under UTC, America/New_York and Pacific/Kiritimati.

## Risks

- **Huh v2 inside a compositor layer.** Focus and the size of a Huh form may not fit a fixed overlay box. Spike this at the start of slice 5. The fallback is to keep the hand-written form with the new styles.
- **Nerd Font detection.** It is a guess from `TERM_PROGRAM`. If users report boxes, change the default to Unicode.
- **Terminal color.** Some terminals do not report the background color. The fallback is the dark preset, and the command palette can switch the theme at runtime.
- **Size of the change.** Slice 1 keeps the behavior the same, so a slice that goes wrong later can be reverted without losing the framework migration.

## Open questions

- Should the theme chosen in the palette persist? That needs a config file (`~/.config/clitasks/config.toml`). The plan uses only env vars and flags until someone asks.
- Should the board show a WIP limit for each column (for example `IN PROGRESS 3/4`)? That needs a new board-level field. It is not planned.
