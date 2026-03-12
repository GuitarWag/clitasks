# Quick Start Guide

## Installation

```bash
npm install
npm run build
```

## 5-Minute Tutorial

### 1. Initialize your board

```bash
npm run dev -- init --name "My Project"
```

This creates `tasks.md` in your current directory.

### 2. Add some tasks

```bash
npm run dev -- add "Setup project structure" -p high
npm run dev -- add "Write tests" -p medium -a yourname
npm run dev -- add "Deploy to staging" -p low -t deployment
```

### 3. View your board

```bash
npm run dev -- board
```

### 4. Start working on a task

```bash
# Use the task ID from the board
npm run dev -- start T-ABC123
```

### 5. Mark it as complete

```bash
npm run dev -- complete T-ABC123
```

### 6. View the board again

```bash
npm run dev -- board
```

## For AI Agents

AI agents like Claude can use this tool to manage tasks:

```bash
# AI reads current state
npm run dev -- board

# AI creates task for its work
npm run dev -- add "Implement authentication" -a claude -t backend

# AI starts the task
npm run dev -- start T-XYZ789

# AI completes the task
npm run dev -- complete T-XYZ789
```

The AI can also read the `tasks.md` file directly to understand project context.

## Human-Readable Format

Open `tasks.md` in any text editor to see your tasks in a clean Markdown format:

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

You can even edit this file manually if you prefer!

## Next Steps

- Read the full [README.md](README.md) for all features
- Check out [example-tasks.md](example-tasks.md) for a complete example
- Run `npm run dev -- --help` to see all commands
