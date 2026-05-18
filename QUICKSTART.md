# Quick Start Guide

## Installation

```bash
make build       # produces bin/tasks
# or
make install     # installs `tasks` into $GOBIN
```

Examples below use the installed `tasks` command. If you only built locally, substitute `./bin/tasks`.

## 5-Minute Tutorial

### 1. Initialize your board

```bash
tasks init --name "My Project"
```

This creates `tasks.md` in your current directory.

### 2. Add some tasks

```bash
tasks add "Setup project structure" -p high
tasks add "Write tests" -p medium -a yourname
tasks add "Deploy to staging" -p low -t deployment
```

### 3. View your board

```bash
tasks board
```

### 4. Start working on a task

```bash
# Use the task ID from the board
tasks start T-ABC123
```

### 5. Mark it as complete

```bash
tasks complete T-ABC123
```

### 6. View the board again

```bash
tasks board
```

## Interactive TUI

```bash
tasks tui
```

Use `hjkl` or arrow keys to move, `a` to add, `e` to edit, `s` to change status, `f` to filter, `r` to reload, `?` for help, `q` to quit.

## For AI Agents

AI agents can call the CLI like any user:

```bash
tasks board
tasks add "Implement authentication" -a claude -t backend
tasks start T-XYZ789
tasks complete T-XYZ789
```

They can also read `tasks.md` directly to load context.

## Human-Readable Format

`tasks.md` is plain Markdown:

```markdown
# Board: My Project

## TODO
- [ ] [T-ABC123] **Setup project structure** `priority:high`

## IN PROGRESS
_No tasks_

## DONE
_No tasks_

## BLOCKED
_No tasks_
```

You can edit it by hand if you prefer.

## Next Steps

- [README.md](README.md) — full reference
- [TUI_GUIDE.md](TUI_GUIDE.md) — keybindings and modes
- `tasks --help`
