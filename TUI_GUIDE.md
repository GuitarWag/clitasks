# Terminal UI (TUI) Guide

## Overview

The Task Board TUI provides a beautiful, interactive terminal interface for managing your tasks. Navigate with arrow keys, edit tasks in-place, and get a real-time kanban view.

## Launching the TUI

There are three ways to launch the TUI:

```bash
# Method 1: Via tasks command
tasks tui

# Method 2: Direct executable
tasks-tui

# Method 3: With custom file
tasks -f myboard.md tui
```

## Interface Layout

```
┌─────────────────────────────────────────────────────────────┐
│        Project Alpha - Main development board              │
│  Tasks: 7 | File: tasks.md                                 │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│        Filter: backend (Press Esc to clear)                │
└─────────────────────────────────────────────────────────────┘
┌──────────┬──────────────┬──────────┬──────────┐
│   TODO   │ IN PROGRESS  │   DONE   │ BLOCKED  │
│   (3)    │     (1)      │   (2)    │   (1)    │
├──────────┼──────────────┼──────────┼──────────┤
│          │              │          │          │
│  ● Task1 │  ▶ Task2     │  ● Task5 │  ● Task7 │
│  ● Task3 │    @alice    │  ● Task6 │          │
│  ● Task4 │    #backend  │          │          │
│          │              │          │          │
└──────────┴──────────────┴──────────┴──────────┘
┌─────────────────────────────────────────────────────────────┐
│  ↑/↓ Navigate | ←/→ Column | e Edit | a Add | q Quit       │
└─────────────────────────────────────────────────────────────┘
```

## Keyboard Shortcuts

### Navigation
- **↑** or **k** - Move up in current column
- **↓** or **j** - Move down in current column
- **←** or **h** - Move to previous column
- **→** or **l** - Move to next column

### Task Management
- **a** - Add new task
- **e** - Edit selected task
- **d** - Delete selected task (with confirmation)
- **s** - Quick status change menu

### View & Search
- **f** - Filter/search tasks across all columns
- **r** - Refresh board from file
- **h** or **?** - Show help screen

### General
- **q** or **Ctrl+C** - Quit application
- **Esc** - Cancel dialog / Clear filter

## Features

### 1. Column-Based Kanban View

The TUI displays your tasks in four columns:
- **TODO** (Yellow) - Tasks waiting to be started
- **IN PROGRESS** (Blue) - Active tasks
- **DONE** (Green) - Completed tasks
- **BLOCKED** (Red) - Tasks that are blocked

Navigate between columns with arrow keys. The active column is highlighted.

### 2. Task Details on Selection

When you select a task, you'll see:
- ● Priority indicator (color-coded)
- ▶ Selection indicator
- Task title (highlighted)
- Assignee (if set)
- Tags (if set)
- Due date (if set)
- Description preview (truncated if long)

### 3. Add Tasks (Press 'a')

Opens a dialog with fields for:
- **Title** - Task name (required)
- **Description** - Detailed description
- **Priority** - low, medium, high, or critical
- **Assignee** - Person responsible
- **Tags** - Comma-separated tags

Use Tab/Shift+Tab to move between fields. Press Enter on "Save" or use the button.

### 4. Edit Tasks (Press 'e')

Edit any field of the selected task:
- Modify title, description, priority
- Change assignee
- Update tags
- All changes are saved to the markdown file

### 5. Delete Tasks (Press 'd')

Confirmation dialog before deletion:
- Shows task title
- Asks for confirmation (Yes/No)
- Cannot be undone

### 6. Quick Status Change (Press 's')

Fast way to move tasks between columns:
- Opens menu with all statuses
- Use arrow keys to select
- Press Enter to move task

### 7. Real-Time Filtering (Press 'f') 🆕

**NEW FEATURE**: Search across all tasks and columns

Filter searches:
- Task titles
- Descriptions
- Assignees
- Tags

Example filters:
- "backend" - Shows all tasks tagged or mentioning backend
- "alice" - Shows tasks assigned to alice
- "bug" - Shows all bug-related tasks

Press **Esc** to clear the filter.

### 8. Quick Actions Menu (Press 's') 🆕

**NEW FEATURE**: Fast keyboard-driven task status changes

Instead of:
1. Press 'e' to edit
2. Navigate to status field
3. Change status
4. Save

Just:
1. Press 's'
2. Select new status
3. Done!

## Color Coding

### Priority Colors
- 🔴 **Red** - Critical
- 🟡 **Yellow** - High
- 🔵 **Blue** - Medium
- ⚪ **White** - Low

### Column Colors
- 🟡 **Yellow** - TODO
- 🔵 **Blue** - IN PROGRESS
- 🟢 **Green** - DONE
- 🔴 **Red** - BLOCKED

## Tips & Tricks

### 1. Vim-Style Navigation
If you're a Vim user, you'll feel at home:
- `h/j/k/l` for navigation (left/down/up/right)
- Works alongside arrow keys

### 2. Quick Filtering Workflow
```
1. Press 'f'
2. Type "alice"
3. See only Alice's tasks
4. Navigate and edit
5. Press 'f' then Esc to clear
```

### 3. Bulk Status Changes
```
1. Filter for specific tasks (e.g., "frontend")
2. Navigate through filtered results
3. Press 's' to change status quickly
4. Repeat for each task
```

### 4. Multi-File Boards
```bash
# Work board
tasks -f work.md tui

# Personal board
tasks -f personal.md tui

# Project board
cd ~/project
tasks tui  # Uses local tasks.md
```

### 5. Refresh After External Changes
If someone else (or another process) modifies the tasks.md file:
- Press **r** to reload from disk
- Your view updates immediately

## Comparison: CLI vs TUI

| Feature | CLI | TUI |
|---------|-----|-----|
| View all columns | `tasks board` | Always visible |
| Navigate tasks | List commands | Arrow keys |
| Edit task | `tasks update ID` | Press 'e' |
| Filter | `tasks list --filter` | Press 'f' |
| Visual feedback | Text output | Color-coded UI |
| Multi-tasking | New commands | Single interface |
| Speed | Fast for single ops | Fast for browsing |

**When to use CLI:**
- Scripts and automation
- Quick single operations
- CI/CD pipelines
- Remote SSH sessions (if TUI has issues)

**When to use TUI:**
- Planning sessions
- Daily standup review
- Organizing multiple tasks
- Visual task management
- Interactive work sessions

## Accessibility

The TUI is designed to work in:
- ✅ macOS Terminal
- ✅ iTerm2
- ✅ Linux terminals (xterm, gnome-terminal, etc.)
- ✅ Windows Terminal
- ✅ SSH sessions
- ⚠️ tmux/screen (may need 256 color support)

For best experience:
- Use a terminal with 256 color support
- Terminal window at least 80 columns wide
- Terminal window at least 24 rows tall

## Troubleshooting

### Issue: Colors not showing correctly
**Solution**: Check your terminal supports 256 colors
```bash
echo $TERM  # Should show something like "xterm-256color"
```

### Issue: TUI crashes on startup
**Solution**: Ensure tasks.md exists or initialize first
```bash
tasks init --name "My Board"
tasks tui
```

### Issue: Arrow keys not working
**Solution**: Try Vim-style keys (h/j/k/l) instead

### Issue: Can't edit task fields
**Solution**: Use Tab/Shift+Tab to move between fields in dialogs

## Advanced Usage

### Custom File Paths
```bash
# Environment variable
export TASK_BOARD_FILE=~/boards/sprint-5.md
tasks tui

# Command line
tasks -f ~/boards/sprint-5.md tui
```

### Integration with Git
```bash
# Before starting work
git pull
tasks tui

# After making changes
git add tasks.md
git commit -m "Updated task board"
git push
```

### AI Agent Workflow
```bash
# AI checks board visually
tasks tui

# AI creates tasks via CLI (from another terminal)
tasks add "Implement feature X" -a claude

# Refresh TUI to see new task
# (Press 'r' in TUI)
```

## Screenshots (Text Mode)

### Viewing Filtered Tasks
```
Filter: frontend
TODO (1)          IN PROGRESS (1)    DONE (0)         BLOCKED (0)
  ▶ Fix layout      ● Build UI         (empty)          (empty)
    @alice            @bob
    #frontend         #frontend
```

### Editing a Task
```
┌─ Edit Task: T-ABC123 ─────────────────────┐
│                                           │
│ Title:                                    │
│ ┌───────────────────────────────────────┐ │
│ │ Fix responsive layout                 │ │
│ └───────────────────────────────────────┘ │
│                                           │
│ Description:                              │
│ ┌───────────────────────────────────────┐ │
│ │ Mobile view is broken                 │ │
│ │                                       │ │
│ └───────────────────────────────────────┘ │
│                                           │
│ Priority:    Assignee:      Tags:        │
│ ┌─────────┐  ┌─────────┐   ┌──────────┐ │
│ │critical │  │alice    │   │frontend  │ │
│ └─────────┘  └─────────┘   └──────────┘ │
│                                           │
│  [Save]  [Cancel]                        │
└───────────────────────────────────────────┘
```

## Future Enhancements

Potential additions:
- Drag & drop tasks between columns
- Task dependencies visualization
- Time tracking integration
- Customizable color schemes
- Mouse support
- Split screen mode (multiple boards)
- Task history/changelog view
- Bulk operations

---

**Enjoy your new visual task management interface!**

Press 'h' anytime in the TUI for quick help.
