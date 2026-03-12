#!/usr/bin/env node

import { Command } from 'commander';
import chalk from 'chalk';
import { TaskBoard } from './board';
import { TaskStatus, TaskPriority } from './types';
import { Exporter } from './export';
import * as fs from 'fs';

const program = new Command();

// Default file path
const getFilePath = (options: any): string => {
  return options.file || process.env.TASK_BOARD_FILE || 'tasks.md';
};

// Initialize board
const getBoard = (options: any): TaskBoard => {
  return new TaskBoard(getFilePath(options));
};

// Format priority with color
const formatPriority = (priority: TaskPriority): string => {
  const colors = {
    critical: chalk.red,
    high: chalk.yellow,
    medium: chalk.blue,
    low: chalk.gray,
  };
  return colors[priority](priority);
};

// Format status with icon
const formatStatus = (status: TaskStatus): string => {
  const icons = {
    'todo': '☐',
    'in-progress': '▶',
    'done': '✓',
    'blocked': '!',
  };
  return icons[status] + ' ' + status;
};

// Display task details
const displayTask = (task: any, detailed: boolean = false): void => {
  console.log(`\n${chalk.cyan(task.id)} ${chalk.bold(task.title)}`);
  console.log(`  Status: ${formatStatus(task.status)} | Priority: ${formatPriority(task.priority)}`);

  if (task.assignee) {
    console.log(`  Assignee: ${chalk.green(task.assignee)}`);
  }

  if (task.tags && task.tags.length > 0) {
    console.log(`  Tags: ${task.tags.map((t: string) => chalk.magenta(`#${t}`)).join(' ')}`);
  }

  if (task.dueDate) {
    console.log(`  Due: ${chalk.yellow(task.dueDate)}`);
  }

  if (detailed && task.description) {
    console.log(`  ${chalk.dim(task.description)}`);
  }

  if (detailed) {
    console.log(`  ${chalk.dim(`Created: ${task.createdAt} | Updated: ${task.updatedAt}`)}`);
  }
};

// Main CLI setup
program
  .name('tasks')
  .description('CLI task management with Markdown storage')
  .version('2.1.0')
  .option('-f, --file <path>', 'Path to the markdown file (default: tasks.md)');

// Init command
program
  .command('init')
  .description('Initialize a new task board')
  .option('-n, --name <name>', 'Board name', 'My Board')
  .option('-d, --description <desc>', 'Board description')
  .action((options) => {
    const board = getBoard(program.opts());
    board.updateBoardInfo({
      name: options.name,
      description: options.description,
    });
    console.log(chalk.green(`✓ Board initialized at ${board.getFilePath()}`));
  });

// Add command
program
  .command('add <title>')
  .description('Add a new task')
  .option('-d, --description <desc>', 'Task description')
  .option('-p, --priority <priority>', 'Priority (low|medium|high|critical)', 'medium')
  .option('-a, --assignee <name>', 'Assignee name')
  .option('-t, --tags <tags>', 'Comma-separated tags')
  .option('--due <date>', 'Due date (YYYY-MM-DD)')
  .action((title, options) => {
    const board = getBoard(program.opts());
    const task = board.addTask(title, {
      description: options.description,
      priority: options.priority as TaskPriority,
      assignee: options.assignee,
      tags: options.tags ? options.tags.split(',').map((t: string) => t.trim()) : undefined,
      dueDate: options.due,
    });
    console.log(chalk.green(`✓ Task added: ${task.id}`));
    displayTask(task);
  });

// List command
program
  .command('list')
  .description('List tasks')
  .option('-s, --status <status>', 'Filter by status (todo|in-progress|done|blocked)')
  .option('-p, --priority <priority>', 'Filter by priority (low|medium|high|critical)')
  .option('-a, --assignee <name>', 'Filter by assignee')
  .option('-t, --tags <tags>', 'Filter by tags (comma-separated)')
  .option('--detailed', 'Show detailed information')
  .action((options) => {
    const board = getBoard(program.opts());
    const tasks = board.listTasks({
      status: options.status as TaskStatus,
      priority: options.priority as TaskPriority,
      assignee: options.assignee,
      tags: options.tags ? options.tags.split(',').map((t: string) => t.trim()) : undefined,
    });

    if (tasks.length === 0) {
      console.log(chalk.yellow('No tasks found'));
      return;
    }

    console.log(chalk.bold(`\nFound ${tasks.length} task(s):`));
    tasks.forEach(task => displayTask(task, options.detailed));
  });

// Board command (view kanban board)
program
  .command('board')
  .description('Display kanban board view')
  .action(() => {
    const board = getBoard(program.opts());
    const info = board.getBoardInfo();
    const tasksByStatus = board.getTasksByStatus();

    console.log(chalk.bold.cyan(`\n# ${info.name}`));
    if (info.description) {
      console.log(chalk.dim(info.description));
    }
    console.log(chalk.dim(`Last updated: ${info.updatedAt}\n`));

    const statuses: TaskStatus[] = ['todo', 'in-progress', 'blocked', 'done'];
    const statusLabels = {
      'todo': 'TODO',
      'in-progress': 'IN PROGRESS',
      'done': 'DONE',
      'blocked': 'BLOCKED',
    };

    statuses.forEach(status => {
      const tasks = tasksByStatus[status];
      console.log(chalk.bold(`\n## ${statusLabels[status]} (${tasks.length})`));

      if (tasks.length === 0) {
        console.log(chalk.dim('  No tasks'));
      } else {
        tasks.forEach(task => {
          const priority = formatPriority(task.priority);
          const assignee = task.assignee ? chalk.green(`@${task.assignee}`) : '';
          const tags = task.tags ? task.tags.map(t => chalk.magenta(`#${t}`)).join(' ') : '';
          console.log(`  ${chalk.cyan(task.id)} ${task.title} ${priority} ${assignee} ${tags}`);
          if (task.description) {
            console.log(chalk.dim(`    ${task.description}`));
          }
        });
      }
    });
    console.log('');
  });

// Show command (view single task)
program
  .command('show <id>')
  .description('Show task details')
  .action((id) => {
    const board = getBoard(program.opts());
    const task = board.getTask(id);

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    displayTask(task, true);
  });

// Update command
program
  .command('update <id>')
  .description('Update a task')
  .option('-t, --title <title>', 'New title')
  .option('-d, --description <desc>', 'New description')
  .option('-p, --priority <priority>', 'New priority')
  .option('-a, --assignee <name>', 'New assignee')
  .option('--tags <tags>', 'New tags (comma-separated)')
  .option('--due <date>', 'New due date')
  .action((id, options) => {
    const board = getBoard(program.opts());
    const updates: any = {};

    if (options.title) updates.title = options.title;
    if (options.description) updates.description = options.description;
    if (options.priority) updates.priority = options.priority;
    if (options.assignee) updates.assignee = options.assignee;
    if (options.tags) updates.tags = options.tags.split(',').map((t: string) => t.trim());
    if (options.due) updates.dueDate = options.due;

    const task = board.updateTask(id, updates);

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.green(`✓ Task updated: ${id}`));
    displayTask(task, true);
  });

// Move command (change status)
program
  .command('move <id> <status>')
  .description('Move task to different status (todo|in-progress|done|blocked)')
  .action((id, status) => {
    const validStatuses = ['todo', 'in-progress', 'done', 'blocked'];
    if (!validStatuses.includes(status)) {
      console.log(chalk.red(`✗ Invalid status. Must be one of: ${validStatuses.join(', ')}`));
      return;
    }

    const board = getBoard(program.opts());
    const task = board.moveTask(id, status as TaskStatus);

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.green(`✓ Task moved to ${status}: ${id}`));
    displayTask(task);
  });

// Start command (shortcut for moving to in-progress)
program
  .command('start <id>')
  .description('Start working on a task (move to in-progress)')
  .action((id) => {
    const board = getBoard(program.opts());
    const task = board.moveTask(id, 'in-progress');

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.green(`✓ Started task: ${id}`));
    displayTask(task);
  });

// Complete command (shortcut for moving to done)
program
  .command('complete <id>')
  .description('Mark task as complete')
  .action((id) => {
    const board = getBoard(program.opts());
    const task = board.moveTask(id, 'done');

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.green(`✓ Task completed: ${id}`));
    displayTask(task);
  });

// Block command (shortcut for moving to blocked)
program
  .command('block <id>')
  .description('Mark task as blocked')
  .action((id) => {
    const board = getBoard(program.opts());
    const task = board.moveTask(id, 'blocked');

    if (!task) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.yellow(`! Task blocked: ${id}`));
    displayTask(task);
  });

// Delete command
program
  .command('delete <id>')
  .description('Delete a task')
  .action((id) => {
    const board = getBoard(program.opts());
    const success = board.deleteTask(id);

    if (!success) {
      console.log(chalk.red(`✗ Task not found: ${id}`));
      return;
    }

    console.log(chalk.green(`✓ Task deleted: ${id}`));
  });

// Info command
program
  .command('info')
  .description('Show board information')
  .action(() => {
    const board = getBoard(program.opts());
    const info = board.getBoardInfo();

    console.log(chalk.bold.cyan(`\n${info.name}`));
    if (info.description) {
      console.log(chalk.dim(info.description));
    }
    console.log(`\nFile: ${chalk.yellow(board.getFilePath())}`);
    console.log(`Total tasks: ${chalk.cyan(info.tasks.length)}`);
    console.log(`Created: ${chalk.dim(info.createdAt)}`);
    console.log(`Updated: ${chalk.dim(info.updatedAt)}\n`);
  });

// Stats command
program
  .command('stats')
  .description('Show board statistics')
  .action(() => {
    const board = getBoard(program.opts());
    const info = board.getBoardInfo();
    const tasksByStatus = board.getTasksByStatus();

    console.log(chalk.bold.cyan(`\n📊 ${info.name} - Statistics\n`));

    // Status breakdown
    console.log(chalk.bold('Status Breakdown:'));
    console.log(`  TODO:        ${chalk.yellow(tasksByStatus['todo'].length)}`);
    console.log(`  IN PROGRESS: ${chalk.blue(tasksByStatus['in-progress'].length)}`);
    console.log(`  DONE:        ${chalk.green(tasksByStatus['done'].length)}`);
    console.log(`  BLOCKED:     ${chalk.red(tasksByStatus['blocked'].length)}`);
    console.log(`  ${chalk.dim('──────────────')}`);
    console.log(`  Total:       ${chalk.cyan(info.tasks.length)}\n`);

    // Priority breakdown
    const priorityCounts = {
      critical: info.tasks.filter(t => t.priority === 'critical').length,
      high: info.tasks.filter(t => t.priority === 'high').length,
      medium: info.tasks.filter(t => t.priority === 'medium').length,
      low: info.tasks.filter(t => t.priority === 'low').length,
    };

    console.log(chalk.bold('Priority Breakdown:'));
    console.log(`  Critical: ${chalk.red(priorityCounts.critical)}`);
    console.log(`  High:     ${chalk.yellow(priorityCounts.high)}`);
    console.log(`  Medium:   ${chalk.blue(priorityCounts.medium)}`);
    console.log(`  Low:      ${chalk.gray(priorityCounts.low)}\n`);

    // Assignee breakdown
    const assignees = new Map<string, number>();
    info.tasks.forEach(task => {
      if (task.assignee) {
        assignees.set(task.assignee, (assignees.get(task.assignee) || 0) + 1);
      }
    });

    if (assignees.size > 0) {
      console.log(chalk.bold('Assignee Breakdown:'));
      Array.from(assignees.entries())
        .sort((a, b) => b[1] - a[1])
        .forEach(([assignee, count]) => {
          console.log(`  ${chalk.green(assignee)}: ${count} task${count > 1 ? 's' : ''}`);
        });
      console.log('');
    }

    // Completion rate
    const completionRate = info.tasks.length > 0
      ? ((tasksByStatus['done'].length / info.tasks.length) * 100).toFixed(1)
      : 0;
    console.log(chalk.bold(`Completion Rate: ${chalk.cyan(completionRate + '%')}\n`));
  });

// Export command
program
  .command('export')
  .description('Export board data')
  .option('-f, --format <format>', 'Export format (json|csv|summary)', 'json')
  .option('-o, --output <file>', 'Output file (defaults to stdout)')
  .action((options) => {
    const board = getBoard(program.opts());
    const exporter = new Exporter(board);

    let output: string;

    switch (options.format) {
      case 'json':
        output = exporter.toJSON();
        break;
      case 'csv':
        output = exporter.toCSV();
        break;
      case 'summary':
        output = exporter.toSummary();
        break;
      default:
        console.log(chalk.red(`✗ Invalid format: ${options.format}. Use json, csv, or summary.`));
        return;
    }

    if (options.output) {
      fs.writeFileSync(options.output, output, 'utf-8');
      console.log(chalk.green(`✓ Exported to ${options.output}`));
    } else {
      console.log(output);
    }
  });

// TUI command
program
  .command('tui')
  .description('Launch interactive Terminal UI')
  .action(() => {
    // Dynamic import to avoid loading blessed unless needed
    const { TaskTUI } = require('./tui');
    const filePath = getFilePath(program.opts());
    const tui = new TaskTUI(filePath);
    tui.run();
  });

program.parse();
