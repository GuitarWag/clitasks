# Task Management with `tasks` CLI

Use the `tasks` CLI to manage project work via a `tasks.md` Markdown file in the working directory. All commands run via Bash.

## Storage

Tasks live in `tasks.md` in the current directory (created automatically on first use). You can also specify a different file with `-f <path>` or `TASK_BOARD_FILE` env var.

## Commands

### View the board

```bash
tasks board                  # Kanban board view
tasks list                   # List all tasks
tasks list -s todo           # Filter by status: todo | in-progress | done | blocked
tasks list -p high           # Filter by priority: low | medium | high | critical
tasks list -a alice          # Filter by assignee
tasks list -t backend        # Filter by tag
tasks list --detailed        # Include descriptions and timestamps
tasks show <task-id>         # Show one task in detail
tasks info                   # Board metadata
tasks stats                  # Status/priority/assignee breakdown
tasks timeline               # Gantt chart of tasks by start and due date
tasks timeline -s in-progress -w 120   # Filter (-s, -a, -t) and set the width
```

`tasks timeline` also lists `after` conflicts (a task that starts before a task it waits for ends) and tasks with a due date that is not YYYY-MM-DD.

### Create tasks

```bash
tasks add "Title"
tasks add "Title" -d "Description" -p high -a claude -t backend,api --due 2026-04-01
tasks add "Title" --start 2026-03-20 --due 2026-04-01 --after T-ML31897Y-TKP
```

Options:
- `-d` description
- `-p` priority (low | medium | high | critical) — default: medium
- `-a` assignee
- `-t` comma-separated tags
- `--due` date in YYYY-MM-DD. The CLI rejects other formats.
- `--start` date in YYYY-MM-DD. It must not be after `--due`.
- `--after` comma-separated IDs of tasks that this task waits for. The CLI rejects unknown IDs and cycles.

### Update tasks

```bash
tasks update <task-id> -t "New title"
tasks update <task-id> -d "New description"
tasks update <task-id> -p critical -a bob --tags backend,urgent --due 2026-05-01
tasks update <task-id> --start 2026-04-20 --after T-AAA-111,T-BBB-222
tasks update <task-id> --after ""       # Clear a field with an empty value
```

### Change status

```bash
tasks start <task-id>        # Move to in-progress
tasks complete <task-id>     # Move to done
tasks block <task-id>        # Move to blocked
tasks move <task-id> todo    # Move to any status
```

A move to in-progress sets the start date to today if the task has no start date.

### Delete

```bash
tasks delete <task-id>
```

Delete also removes the task ID from the `after` list of other tasks.

### Export

```bash
tasks export                              # JSON to stdout
tasks export --format csv                 # CSV to stdout
tasks export --format summary             # Human-readable summary
tasks export --format gantt               # Mermaid gantt chart
tasks export --format json -o backup.json # Write to file
```

`--format` has no short flag. `-f` is the board file.

### Initialize a named board

```bash
tasks init -n "Sprint 1" -d "Sprint 1 tasks"
```

## Task IDs

Every task gets a unique ID like `T-ML31897Y-TKP`. Use this ID for all update/move/delete/show commands. Get IDs from `tasks board` or `tasks list`.

## Workflow for AI agents

1. **Before starting work**: Run `tasks board` to see current state
2. **Plan work**: Create tasks with `tasks add` — use `-d` for context, `-a claude` to self-assign
3. **Track progress**: `tasks start <id>` when beginning, `tasks complete <id>` when done
4. **Document blockers**: `tasks block <id>` then `tasks update <id> -d "Blocked: reason"`
5. **Prioritize**: Focus on `critical` and `high` priority tasks first

## Markdown format

The `tasks.md` file is human-readable Markdown. You can also read/parse it directly:

```markdown
## TODO

- [ ] [T-ABC123] **Task title** `priority:high` `assignee:alice` `tags:backend,api` `due:2026-04-01`
  > Task description here
  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-02T00:00:00.000Z

## IN PROGRESS

- [>] [T-DEF456] **Another task** `priority:medium` `due:2026-04-10` `start:2026-04-01` `after:T-ABC123`

## DONE

- [x] [T-GHI789] **Completed task** `priority:low`

## BLOCKED

- [!] [T-JKL012] **Blocked task** `priority:critical`
```

Status markers: `[ ]` todo, `[>]` in-progress, `[x]` done, `[!]` blocked.
