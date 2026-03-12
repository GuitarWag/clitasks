#!/usr/bin/env node

import * as blessed from 'blessed';
import { TaskBoard } from './board';
import { Task, TaskStatus, TaskPriority } from './types';

interface ColumnData {
  status: TaskStatus;
  title: string;
  color: string;
}

const COLUMNS: ColumnData[] = [
  { status: 'todo', title: 'TODO', color: 'yellow' },
  { status: 'in-progress', title: 'IN PROGRESS', color: 'blue' },
  { status: 'done', title: 'DONE', color: 'green' },
  { status: 'blocked', title: 'BLOCKED', color: 'red' },
];

export class TaskTUI {
  private screen: blessed.Widgets.Screen;
  private board: TaskBoard;
  private currentColumn: number = 0;
  private currentTaskIndex: number = 0;
  private filterText: string = '';
  private showingHelp: boolean = false;

  constructor(filePath?: string) {
    this.board = new TaskBoard(filePath);
    this.screen = blessed.screen({
      smartCSR: true,
      title: 'Task Board - TUI',
    });

    this.setupUI();
    this.setupKeyBindings();
  }

  private setupUI(): void {
    // Header
    const header = blessed.box({
      top: 0,
      left: 0,
      width: '100%',
      height: 3,
      content: '',
      tags: true,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'cyan',
        },
      },
    });
    this.screen.append(header);

    // Filter input box
    const filterBox = blessed.box({
      top: 3,
      left: 0,
      width: '100%',
      height: 3,
      content: '',
      tags: true,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'magenta',
        },
      },
    });
    this.screen.append(filterBox);

    // Main content area for columns
    const mainArea = blessed.box({
      top: 6,
      left: 0,
      width: '100%',
      height: '100%-9',
      tags: true,
    });
    this.screen.append(mainArea);

    // Footer with help
    const footer = blessed.box({
      bottom: 0,
      left: 0,
      width: '100%',
      height: 3,
      content: '{center}{bold}↑/↓{/bold} Navigate | {bold}←/→{/bold} Switch Column | {bold}e{/bold} Edit | {bold}a{/bold} Add | {bold}d{/bold} Delete | {bold}s{/bold} Status | {bold}f{/bold} Filter | {bold}h{/bold} Help | {bold}q{/bold} Quit{/center}',
      tags: true,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'cyan',
        },
      },
    });
    this.screen.append(footer);

    this.render();
  }

  private setupKeyBindings(): void {
    // Quit
    this.screen.key(['q', 'C-c'], () => {
      return process.exit(0);
    });

    // Navigate up/down
    this.screen.key(['up', 'k'], () => {
      if (this.currentTaskIndex > 0) {
        this.currentTaskIndex--;
        this.render();
      }
    });

    this.screen.key(['down', 'j'], () => {
      const tasks = this.getFilteredTasksForColumn(COLUMNS[this.currentColumn].status);
      if (this.currentTaskIndex < tasks.length - 1) {
        this.currentTaskIndex++;
        this.render();
      }
    });

    // Navigate left/right between columns
    this.screen.key(['left', 'h'], () => {
      if (this.currentColumn > 0) {
        this.currentColumn--;
        this.currentTaskIndex = 0;
        this.render();
      }
    });

    this.screen.key(['right', 'l'], () => {
      if (this.currentColumn < COLUMNS.length - 1) {
        this.currentColumn++;
        this.currentTaskIndex = 0;
        this.render();
      }
    });

    // Add task
    this.screen.key(['a'], () => {
      this.showAddTaskDialog();
    });

    // Edit task
    this.screen.key(['e'], () => {
      this.showEditTaskDialog();
    });

    // Delete task
    this.screen.key(['d'], () => {
      this.showDeleteConfirmation();
    });

    // Change status (quick actions)
    this.screen.key(['s'], () => {
      this.showStatusMenu();
    });

    // Filter
    this.screen.key(['f'], () => {
      this.showFilterDialog();
    });

    // Help
    this.screen.key(['h', '?'], () => {
      this.showHelp();
    });

    // Refresh
    this.screen.key(['r'], () => {
      this.board = new TaskBoard(this.board.getFilePath());
      this.render();
    });
  }

  private getFilteredTasksForColumn(status: TaskStatus): Task[] {
    let tasks = this.board.listTasks({ status });

    if (this.filterText) {
      const filter = this.filterText.toLowerCase();
      tasks = tasks.filter(task =>
        task.title.toLowerCase().includes(filter) ||
        task.description?.toLowerCase().includes(filter) ||
        task.assignee?.toLowerCase().includes(filter) ||
        task.tags?.some(tag => tag.toLowerCase().includes(filter))
      );
    }

    return tasks;
  }

  private render(): void {
    const info = this.board.getBoardInfo();

    // Update header
    const header = this.screen.children[0] as blessed.Widgets.BoxElement;
    header.setContent(`{center}{bold}{cyan-fg}${info.name}{/cyan-fg}{/bold}\n` +
      `{center}${info.description || ''} | Tasks: ${info.tasks.length} | File: ${this.board.getFilePath()}{/center}`);

    // Update filter box
    const filterBox = this.screen.children[1] as blessed.Widgets.BoxElement;
    const filterDisplay = this.filterText
      ? `{bold}Filter:{/bold} ${this.filterText} (Press Esc to clear)`
      : `Press 'f' to filter tasks`;
    filterBox.setContent(`{center}${filterDisplay}{/center}`);

    // Clear main area
    const mainArea = this.screen.children[2] as blessed.Widgets.BoxElement;
    mainArea.children.forEach(child => child.detach());

    // Calculate column width
    const columnWidth = Math.floor(100 / COLUMNS.length);

    // Render columns
    COLUMNS.forEach((column, colIndex) => {
      const tasks = this.getFilteredTasksForColumn(column.status);
      const isActive = colIndex === this.currentColumn;

      const columnBox = blessed.box({
        parent: mainArea,
        left: `${colIndex * columnWidth}%`,
        top: 0,
        width: `${columnWidth}%`,
        height: '100%',
        border: {
          type: 'line',
        },
        style: {
          border: {
            fg: isActive ? column.color : 'white',
            bold: isActive,
          },
        },
        tags: true,
      });

      // Column header
      let content = `{center}{bold}{${column.color}-fg}${column.title} (${tasks.length}){/${column.color}-fg}{/bold}{/center}\n\n`;

      // Render tasks
      tasks.forEach((task, taskIndex) => {
        const isSelected = isActive && taskIndex === this.currentTaskIndex;
        const priorityColor = this.getPriorityColor(task.priority);
        const prefix = isSelected ? '▶ ' : '  ';

        let taskLine = `${prefix}{${priorityColor}-fg}●{/${priorityColor}-fg} `;

        if (isSelected) {
          taskLine += `{inverse}${task.title}{/inverse}`;
        } else {
          taskLine += task.title;
        }

        content += taskLine + '\n';

        if (isSelected) {
          // Show more details for selected task
          if (task.assignee) {
            content += `  @${task.assignee}\n`;
          }
          if (task.tags && task.tags.length > 0) {
            content += `  #${task.tags.join(' #')}\n`;
          }
          if (task.dueDate) {
            content += `  Due: ${task.dueDate}\n`;
          }
          if (task.description) {
            const desc = task.description.length > 50
              ? task.description.substring(0, 47) + '...'
              : task.description;
            content += `  ${desc}\n`;
          }
          content += '\n';
        }
      });

      if (tasks.length === 0) {
        content += '\n{center}No tasks{/center}';
      }

      columnBox.setContent(content);
    });

    this.screen.render();
  }

  private getPriorityColor(priority: TaskPriority): string {
    const colors = {
      critical: 'red',
      high: 'yellow',
      medium: 'blue',
      low: 'white',
    };
    return colors[priority];
  }

  private getCurrentTask(): Task | null {
    const tasks = this.getFilteredTasksForColumn(COLUMNS[this.currentColumn].status);
    return tasks[this.currentTaskIndex] || null;
  }

  private showAddTaskDialog(): void {
    // Step 1: Get title
    const titlePrompt = blessed.prompt({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 60,
      height: 9,
      border: 'line',
      style: {
        border: { fg: 'green' },
      },
      label: ' {bold}Add New Task - Step 1/5{/bold} ',
      tags: true,
    });

    titlePrompt.input('Task Title (required):', '', (err, title) => {
      if (err || !title || !title.trim()) {
        titlePrompt.destroy();
        this.render();
        return;
      }

      const taskData: any = { title: title.trim() };

      // Step 2: Get description
      const descPrompt = blessed.prompt({
        parent: this.screen,
        top: 'center',
        left: 'center',
        width: 60,
        height: 9,
        border: 'line',
        style: {
          border: { fg: 'green' },
        },
        label: ' {bold}Add New Task - Step 2/5{/bold} ',
        tags: true,
      });

      descPrompt.input('Description (optional, press Enter to skip):', '', (err, desc) => {
        taskData.description = desc && desc.trim() ? desc.trim() : undefined;

        // Step 3: Get priority
        const priorityPrompt = blessed.prompt({
          parent: this.screen,
          top: 'center',
          left: 'center',
          width: 60,
          height: 9,
          border: 'line',
          style: {
            border: { fg: 'green' },
          },
          label: ' {bold}Add New Task - Step 3/5{/bold} ',
          tags: true,
        });

        priorityPrompt.input('Priority (low/medium/high/critical, default: medium):', 'medium', (err, priority) => {
          const validPriorities = ['low', 'medium', 'high', 'critical'];
          taskData.priority = validPriorities.includes(priority?.trim()) ? priority.trim() : 'medium';

          // Step 4: Get assignee
          const assigneePrompt = blessed.prompt({
            parent: this.screen,
            top: 'center',
            left: 'center',
            width: 60,
            height: 9,
            border: 'line',
            style: {
              border: { fg: 'green' },
            },
            label: ' {bold}Add New Task - Step 4/5{/bold} ',
            tags: true,
          });

          assigneePrompt.input('Assignee (optional, press Enter to skip):', '', (err, assignee) => {
            taskData.assignee = assignee && assignee.trim() ? assignee.trim() : undefined;

            // Step 5: Get tags
            const tagsPrompt = blessed.prompt({
              parent: this.screen,
              top: 'center',
              left: 'center',
              width: 60,
              height: 9,
              border: 'line',
              style: {
                border: { fg: 'green' },
              },
              label: ' {bold}Add New Task - Step 5/5{/bold} ',
              tags: true,
            });

            tagsPrompt.input('Tags (comma-separated, optional):', '', (err, tags) => {
              if (tags && tags.trim()) {
                taskData.tags = tags.trim().split(',').map(t => t.trim()).filter(t => t);
              }

              // Create the task
              this.board.addTask(taskData.title, {
                description: taskData.description,
                priority: taskData.priority,
                assignee: taskData.assignee,
                tags: taskData.tags,
              });

              this.render();
            });
          });
        });
      });
    });
  }

  private showEditTaskDialog(): void {
    const task = this.getCurrentTask();
    if (!task) return;

    const updates: any = {};

    // Step 1: Edit title
    const titlePrompt = blessed.prompt({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 60,
      height: 9,
      border: 'line',
      style: {
        border: { fg: 'yellow' },
      },
      label: ` {bold}Edit Task: ${task.id} - Step 1/5{/bold} `,
      tags: true,
    });

    titlePrompt.input('Task Title:', task.title, (err, title) => {
      if (err || title === null) {
        titlePrompt.destroy();
        this.render();
        return;
      }
      if (title.trim() && title.trim() !== task.title) {
        updates.title = title.trim();
      }

      // Step 2: Edit description
      const descPrompt = blessed.prompt({
        parent: this.screen,
        top: 'center',
        left: 'center',
        width: 60,
        height: 9,
        border: 'line',
        style: {
          border: { fg: 'yellow' },
        },
        label: ` {bold}Edit Task: ${task.id} - Step 2/5{/bold} `,
        tags: true,
      });

      descPrompt.input('Description:', task.description || '', (err, desc) => {
        const newDesc = desc && desc.trim() ? desc.trim() : undefined;
        if (newDesc !== task.description) {
          updates.description = newDesc;
        }

        // Step 3: Edit priority
        const priorityPrompt = blessed.prompt({
          parent: this.screen,
          top: 'center',
          left: 'center',
          width: 60,
          height: 9,
          border: 'line',
          style: {
            border: { fg: 'yellow' },
          },
          label: ` {bold}Edit Task: ${task.id} - Step 3/5{/bold} `,
          tags: true,
        });

        priorityPrompt.input('Priority (low/medium/high/critical):', task.priority, (err, priority) => {
          const validPriorities = ['low', 'medium', 'high', 'critical'];
          const newPriority = priority && validPriorities.includes(priority.trim()) ? priority.trim() : task.priority;
          if (newPriority !== task.priority) {
            updates.priority = newPriority;
          }

          // Step 4: Edit assignee
          const assigneePrompt = blessed.prompt({
            parent: this.screen,
            top: 'center',
            left: 'center',
            width: 60,
            height: 9,
            border: 'line',
            style: {
              border: { fg: 'yellow' },
            },
            label: ` {bold}Edit Task: ${task.id} - Step 4/5{/bold} `,
            tags: true,
          });

          assigneePrompt.input('Assignee:', task.assignee || '', (err, assignee) => {
            const newAssignee = assignee && assignee.trim() ? assignee.trim() : undefined;
            if (newAssignee !== task.assignee) {
              updates.assignee = newAssignee;
            }

            // Step 5: Edit tags
            const tagsPrompt = blessed.prompt({
              parent: this.screen,
              top: 'center',
              left: 'center',
              width: 60,
              height: 9,
              border: 'line',
              style: {
                border: { fg: 'yellow' },
              },
              label: ` {bold}Edit Task: ${task.id} - Step 5/5{/bold} `,
              tags: true,
            });

            const currentTags = task.tags ? task.tags.join(', ') : '';
            tagsPrompt.input('Tags (comma-separated):', currentTags, (err, tags) => {
              const newTags = tags && tags.trim() ? tags.trim().split(',').map(t => t.trim()).filter(t => t) : undefined;
              if (JSON.stringify(newTags) !== JSON.stringify(task.tags)) {
                updates.tags = newTags;
              }

              // Apply updates
              if (Object.keys(updates).length > 0) {
                this.board.updateTask(task.id, updates);
              }

              this.render();
            });
          });
        });
      });
    });
  }

  private showDeleteConfirmation(): void {
    const task = this.getCurrentTask();
    if (!task) return;

    const confirm = blessed.question({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 50,
      height: 7,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'red',
        },
      },
    });

    confirm.ask(`Delete task "${task.title}"?\n\nThis cannot be undone.`, (err, value) => {
      if (value) {
        this.board.deleteTask(task.id);
        if (this.currentTaskIndex > 0) {
          this.currentTaskIndex--;
        }
      }
      confirm.destroy();
      this.render();
    });

    this.screen.render();
  }

  private showStatusMenu(): void {
    const task = this.getCurrentTask();
    if (!task) return;

    const menu = blessed.list({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 40,
      height: 10,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'cyan',
        },
        selected: {
          bg: 'blue',
        },
      },
      keys: true,
      vi: true,
      items: [
        'TODO',
        'IN PROGRESS',
        'DONE',
        'BLOCKED',
      ],
    });

    blessed.text({
      parent: menu,
      top: -2,
      left: 2,
      content: '{bold}Move to:{/bold}',
      tags: true,
    });

    menu.on('select', (item, index) => {
      const statuses: TaskStatus[] = ['todo', 'in-progress', 'done', 'blocked'];
      this.board.moveTask(task.id, statuses[index]);
      menu.destroy();
      this.render();
    });

    menu.key(['escape'], () => {
      menu.destroy();
      this.render();
    });

    menu.focus();
    this.screen.render();
  }

  private showFilterDialog(): void {
    const input = blessed.textbox({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 50,
      height: 3,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'magenta',
        },
      },
      inputOnFocus: true,
      value: this.filterText,
    });

    blessed.text({
      parent: input,
      top: -2,
      left: 2,
      content: '{bold}Filter tasks (title, description, assignee, tags):{/bold}',
      tags: true,
    });

    input.on('submit', (value) => {
      this.filterText = value.trim();
      this.currentTaskIndex = 0;
      input.destroy();
      this.render();
    });

    input.key(['escape'], () => {
      this.filterText = '';
      this.currentTaskIndex = 0;
      input.destroy();
      this.render();
    });

    input.focus();
    this.screen.render();
  }

  private showHelp(): void {
    const help = blessed.box({
      parent: this.screen,
      top: 'center',
      left: 'center',
      width: 80,
      height: 24,
      border: {
        type: 'line',
      },
      style: {
        border: {
          fg: 'cyan',
        },
      },
      tags: true,
      scrollable: true,
      alwaysScroll: true,
      keys: true,
      vi: true,
      content: `{center}{bold}Task Board - TUI Help{/bold}{/center}

{bold}Navigation:{/bold}
  ↑/k             Move up
  ↓/j             Move down
  ←/h             Previous column
  →/l             Next column

{bold}Task Management:{/bold}
  a               Add new task
  e               Edit selected task
  d               Delete selected task
  s               Change task status (quick menu)

{bold}View Options:{/bold}
  f               Filter tasks (search)
  r               Refresh board
  h/?             Show this help

{bold}General:{/bold}
  q/Ctrl+C        Quit application
  Esc             Cancel/Clear filter

{bold}New Features:{/bold}
  • Real-time filtering across all columns
  • Quick status change menu with keyboard shortcuts
  • Visual column-based kanban view
  • In-TUI task editing with full metadata support

{center}Press any key to close{/center}`,
    });

    help.key(['escape', 'q', 'h', '?', 'enter', 'space'], () => {
      help.destroy();
      this.render();
    });

    help.focus();
    this.screen.render();
  }

  public run(): void {
    this.render();
    this.screen.render();
  }
}

// Run if executed directly
if (require.main === module) {
  const filePath = process.argv[2] || process.env.TASK_BOARD_FILE || 'tasks.md';
  const tui = new TaskTUI(filePath);
  tui.run();
}
