# Plan: timeline and Gantt view

Decisions:

- [Add start and after fields to the task line](../adr/20261004-132518-add-start-and-after-fields-to.md)
- [Compute one timeline layout for CLI, TUI and Mermaid](../adr/20261004-132518-compute-one-timeline-layout-for-cli.md)

Both records are `proposed`. Change them to `accepted` when slice 1 and slice 3 merge.

## Scope

- `tasks timeline`: an ASCII Gantt chart on stdout.
- TUI: the `t` key switches between the board and the timeline.
- `tasks export --format gantt`: a Mermaid `gantt` block.
- New task fields: `start` (date) and `after` (task IDs).

Out of scope: automatic scheduling from `after`, hours or times of day, a
milestone type, and `after` editing in the TUI form (the CLI only, for now).

## Slices

Each slice is one commit. Each slice passes `make test` and `make lint` before the next one starts.

### 1. Model and storage: `start` and `after`

Files: `internal/model/task.go`, `internal/storage/parse.go`, `internal/storage/render.go`.

- Add `Start string` and `After []string` (JSON `start,omitempty`, `after,omitempty`).
- Parse `` `start:...` `` and `` `after:a,b` ``. Render them after `due:`.
- Add `model.ParseDate(s) (time.Time, error)` for `YYYY-MM-DD` (`time.DateOnly`).

Verify:

- A round trip of a task with both fields gives the same task.
- A file without the fields parses the same as before (existing tests pass).
- The existing legacy-format tests still pass.

### 2. Board service and CLI flags

Files: `internal/board/board.go`, `internal/cli/add.go`, `internal/cli/update.go`,
`internal/cli/show.go`, `internal/cli/root.go` (`renderTask`).

- `AddInput` and `UpdateInput` get `Start` and `After`.
- `Add` and `Update` validate the following. Each failure returns an error, and the CLI prints it:
  - `start` and `due` use `YYYY-MM-DD`.
  - `start` is not after `due`.
  - Each `after` ID exists, is not the task itself, and makes no cycle (DFS over `After`).
- `Update` with `Status` set to in-progress sets `Start` to today if `Start` is empty.
  `tasks start` uses this path through `Move`.
- `Delete` removes the ID from the `After` list of other tasks. It returns the changed IDs, and the CLI prints them.
- Flags: `--start`, `--after` (comma-separated, like `--tags`). `update --after ""` clears the list.
- `show` and `list --detailed` print `Start` and `After`.

Verify (table tests in `internal/board`):

- An invalid date is rejected for `start` and for `due`.
- `start` after `due` is rejected.
- An unknown ID is rejected, a self reference is rejected, and an A→B→A cycle is rejected.
- `Move` to in-progress sets `Start` one time and does not overwrite it.
- `Delete` removes the dangling `after` references.

Risk: `add --due "next friday"` works today, and after this slice it fails. Put this in CHANGELOG under "Changed".

### 3. `internal/timeline` layout

Files: `internal/timeline/timeline.go`, `internal/timeline/timeline_test.go`.

`Build(tasks, today) Layout` follows the rules in the layout ADR. The function has no I/O and no styles.

Verify:

- Each bar-end rule has a test: due set, done without due, open.
- A task with a non-date due goes to `Unscheduled`.
- An `after` conflict gives a warning, and the bar dates do not change.
- Sort order: by start date, then by ID.

### 4. ASCII renderer and `tasks timeline`

Files: `internal/timeline/ascii.go`, `internal/cli/timeline.go`, `internal/cli/root.go`.

- `RenderASCII(l Layout, width int, today time.Time) string`. A label column
  (ID and truncated title) and a chart column. The renderer chooses the scale (day or week) to fit
  the range in the chart width. It marks today with `|`, and an open end with `>`.
- It takes no lipgloss styles. The CLI and the TUI color the output after the render, so the
  renderer stays easy to test.
- Command: `tasks timeline [--status] [--assignee] [--tag] [--width N]`.
  The command reuses `board.Filter`. The default width is the terminal width, or 100 if stdout is not a TTY.
- The command prints the warnings and the unscheduled tasks under the chart.

Verify: golden string tests for a fixed `today` and a fixed width. One test for
each scale. One test where the range is longer than the width.

### 5. Mermaid export

Files: `internal/export/export.go`, `internal/cli/export.go`.

- `FormatGantt = "gantt"`. Output: `gantt`, `dateFormat YYYY-MM-DD`, `title`,
  one `section` for each status that has bars, and one line for each bar:
  `<title> :<status-tag>, <mermaid-id>, <start>, <end>`.
  - The tag is `done` for done tasks, `active` for in-progress tasks and `crit` for blocked tasks.
- `<mermaid-id>` is the task ID with `-` changed to `_`. Before you merge, check in the Mermaid
  live editor if Mermaid accepts the original ID.
- In the title, escape `:` and `#`, because Mermaid parses them.
- Update the `--format` help text and the error message.

Verify: a golden test. Paste the output into the Mermaid validator one time before you merge.

### 6. TUI timeline view

Files: `internal/tui/model.go`, `internal/tui/keys.go`, `internal/tui/modals.go`, a new `internal/tui/timeline.go`.

- `modeTimeline`. The `t` key toggles it from the board view.
- The view uses `timeline.RenderASCII` with `m.width`. `j`/`k` select a row,
  `h`/`l` scroll the range one week, `e` opens the existing edit form, and `t`/`esc` go back.
- The view uses the board filter (`applyFilter`).
- The add and edit wizard gets a `Start` step after `Due`. Both steps validate the date.
- Add `t` to the footer and the help screen.

Verify: `Update`/`View` tests in the same style as `tui_test.go`. These include the toggle, the selection
clamp on resize, and the edit from the timeline.

### 7. Docs

- `SKILL.md`: the `timeline` command, `--start`, `--after`, and `--format gantt`. Run `make sync-skill`.
- `README.md`, `TUI_GUIDE.md`, `CHANGELOG.md` (new minor version, with the `--due` validation under "Changed").
- Change the two ADRs to `accepted`. Run `adrlog index` and `adrlog lint`.

## Open questions

- Week scale: does a week start on Monday or on Sunday? The plan uses Monday (ISO).
- Should `tasks timeline` hide `done` tasks by default? The plan shows all tasks. Use `--status` to filter.
