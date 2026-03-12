# CLI Tasks - Scrum/Kanban Task Management

A command-line task management system that uses Markdown files as storage. Perfect for both human developers and AI agents working on projects.

## Features

- **Markdown-based storage**: Tasks are stored in human-readable `.md` files
- **Kanban board**: Organize tasks in TODO, IN PROGRESS, DONE, and BLOCKED columns
- **Rich task metadata**: Priority, assignee, tags, due dates, descriptions
- **CLI interface**: Fast and efficient command-line operations
- **Interactive TUI**: Beautiful terminal UI with keyboard navigation (NEW!)
- **Real-time filtering**: Search tasks across all columns instantly (NEW!)
- **Quick actions**: Keyboard shortcuts for common operations (NEW!)
- **AI-friendly**: Both humans and AI agents can read and edit the same task board
- **Portable**: No database required, just a simple Markdown file

## Installation

```bash
npm install
npm run build
npm link  # Optional: to use 'tasks' command globally
```

## Usage

### Interactive TUI (Recommended for Daily Use)

Launch the beautiful Terminal UI for visual task management:

```bash
tasks tui
```

Features:
- 📊 **Column-based kanban view** with color coding
- ⌨️ **Keyboard navigation** with arrow keys or Vim-style (hjkl)
- ✏️ **In-place editing** of tasks
- 🔍 **Real-time filtering** across all columns
- ⚡ **Quick actions** menu for status changes
- 📋 **Task details** shown on selection

See [TUI_GUIDE.md](TUI_GUIDE.md) for complete documentation.

### Command-Line Interface

For scripts, automation, or quick operations:

### Initialize a Board

```bash
tasks init --name "My Project" --description "Project task board"
```

This creates a `tasks.md` file in the current directory.

### Add Tasks

```bash
# Basic task
tasks add "Implement login feature"

# Task with full metadata
tasks add "Fix authentication bug" \
  -d "JWT tokens are expiring too quickly" \
  -p high \
  -a alice \
  -t backend,security \
  --due 2026-02-15
```

Options:
- `-d, --description <desc>`: Task description
- `-p, --priority <priority>`: Priority (low|medium|high|critical)
- `-a, --assignee <name>`: Assignee name
- `-t, --tags <tags>`: Comma-separated tags
- `--due <date>`: Due date (YYYY-MM-DD)

### View Tasks

```bash
# View kanban board
tasks board

# List all tasks
tasks list

# List with filters
tasks list --status todo
tasks list --priority high
tasks list --assignee alice
tasks list --tags backend

# Show detailed task information
tasks show T-ML31897Y-TKP
```

### Update Tasks

```bash
# Update task details
tasks update T-ML31897Y-TKP \
  -t "Implement authentication API" \
  -p critical \
  -a bob

# Move task to different status
tasks move T-ML31897Y-TKP in-progress

# Shortcuts for common status changes
tasks start T-ML31897Y-TKP     # Move to in-progress
tasks complete T-ML31897Y-TKP   # Move to done
tasks block T-ML31897Y-TKP      # Move to blocked
```

### Delete Tasks

```bash
tasks delete T-ML31897Y-TKP
```

### Board Information

```bash
tasks info
```

### Board Statistics

```bash
# View comprehensive statistics
tasks stats
```

Shows:
- Status breakdown (TODO, IN PROGRESS, DONE, BLOCKED)
- Priority distribution
- Assignee workload
- Completion rate

### Export Data

```bash
# Export as JSON (default)
tasks export

# Export as CSV
tasks export --format csv

# Export summary
tasks export --format summary

# Save to file
tasks export --format json -o backup.json
tasks export --format csv -o report.csv
```

### Using Custom File

By default, tasks are stored in `tasks.md`. You can specify a different file:

```bash
tasks -f project-tasks.md board
tasks --file sprint-1.md add "New feature"

# Or use environment variable
export TASK_BOARD_FILE=my-tasks.md
tasks board
```

## Markdown Format

The task board is stored in a human and AI-readable Markdown format:

```markdown
# Board: Project Alpha
> Description: Main development board
> Created: 2026-02-01T00:58:17.796Z | Updated: 2026-02-01T00:58:22.510Z

## TODO

- [ ] [T-ABC123] **Implement user authentication** `priority:high` `assignee:alice` `tags:backend,security` `due:2026-02-15`
  > Add JWT-based authentication system
  > Created: 2026-02-01T00:58:22.510Z | Updated: 2026-02-01T00:58:22.510Z

## IN PROGRESS

- [>] [T-DEF456] **Create database schema** `priority:medium` `assignee:bob` `tags:backend,database`
  > Design and implement PostgreSQL schema
  > Created: 2026-02-01T00:58:22.510Z | Updated: 2026-02-01T00:58:22.510Z

## DONE

- [x] [T-GHI789] **Setup project repository** `priority:high` `assignee:alice` `tags:devops`
  > Initialize Git repo and CI/CD
  > Created: 2026-02-01T00:58:22.510Z | Updated: 2026-02-01T00:58:22.510Z

## BLOCKED

- [!] [T-JKL012] **Deploy to production** `priority:critical` `assignee:bob` `tags:devops,deployment`
  > Waiting for infrastructure approval
  > Created: 2026-02-01T00:58:22.510Z | Updated: 2026-02-01T00:58:22.510Z
```

## Use Cases

### For Human Developers

```bash
# Morning routine: check what's on the board
tasks board

# Start working on a task
tasks start T-ABC123

# Check your assigned tasks
tasks list --assignee yourname

# Mark task as complete
tasks complete T-ABC123
```

### For AI Agents

AI agents can use the same CLI commands:

```bash
# AI reads the board to understand current work
tasks board

# AI creates a task for work it's about to do
tasks add "Implement error handling" -a claude -t backend

# AI updates task status as it works
tasks start T-XYZ789
# ... do work ...
tasks complete T-XYZ789
```

AI agents can also directly read/parse the `tasks.md` file to understand the project context.

### For Teams

```bash
# View team workload
tasks list --detailed

# Check blocked items
tasks list --status blocked

# Filter by sprint tag
tasks list --tags sprint-1

# View critical items
tasks list --priority critical

# View board statistics
tasks stats

# Export data for reporting
tasks export --format csv -o sprint-report.csv
tasks export --format json -o tasks-backup.json
```

## Development

```bash
# Run in development mode
npm run dev -- board

# Build
npm run build

# Run built version
npm start -- board
```

## TypeScript API

You can also use the task board programmatically:

```typescript
import { TaskBoard } from './board';

const board = new TaskBoard('tasks.md');

// Add a task
const task = board.addTask('Implement feature', {
  priority: 'high',
  assignee: 'alice',
  tags: ['backend'],
});

// Move task
board.moveTask(task.id, 'in-progress');

// List tasks
const inProgressTasks = board.listTasks({ status: 'in-progress' });

// View board
console.log(board.getTasksByStatus());
```

## License

ISC
