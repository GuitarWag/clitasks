import * as fs from 'fs';
import * as path from 'path';
import { Board, Task, TaskStatus, TaskPriority } from './types';

const DEFAULT_BOARD_FILE = 'tasks.md';

/**
 * Markdown storage format:
 *
 * # Board: [Board Name]
 * > Description: [Board Description]
 * > Created: [ISO Date]
 * > Updated: [ISO Date]
 *
 * ## TODO
 * - [ ] [ID] **[Title]** `priority:high` `assignee:name` `tags:tag1,tag2` `due:YYYY-MM-DD`
 *   > [Description]
 *   > Created: [ISO Date] | Updated: [ISO Date]
 *
 * ## IN PROGRESS
 * - [>] [ID] **[Title]** ...
 *
 * ## DONE
 * - [x] [ID] **[Title]** ...
 *
 * ## BLOCKED
 * - [!] [ID] **[Title]** ...
 */

export class MarkdownStorage {
  private filePath: string;

  constructor(filePath: string = DEFAULT_BOARD_FILE) {
    this.filePath = filePath;
  }

  /**
   * Read the board from the markdown file
   */
  readBoard(): Board {
    if (!fs.existsSync(this.filePath)) {
      return this.createDefaultBoard();
    }

    const content = fs.readFileSync(this.filePath, 'utf-8');
    return this.parseMarkdown(content);
  }

  /**
   * Write the board to the markdown file
   */
  writeBoard(board: Board): void {
    const markdown = this.boardToMarkdown(board);
    fs.writeFileSync(this.filePath, markdown, 'utf-8');
  }

  /**
   * Create a default empty board
   */
  private createDefaultBoard(): Board {
    const now = new Date().toISOString();
    return {
      name: 'My Board',
      description: 'Task management board',
      tasks: [],
      createdAt: now,
      updatedAt: now,
    };
  }

  /**
   * Parse markdown content into a Board object
   */
  private parseMarkdown(content: string): Board {
    const lines = content.split('\n');
    const board: Board = {
      name: 'My Board',
      description: undefined,
      tasks: [],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    let currentStatus: TaskStatus | null = null;
    let currentTask: Partial<Task> | null = null;

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i]; // Don't trim - we need leading spaces to identify description lines

      // Parse board title
      if (line.startsWith('# Board:')) {
        board.name = line.replace('# Board:', '').trim();
        continue;
      }

      // Parse board metadata
      if (line.startsWith('> Description:')) {
        board.description = line.replace('> Description:', '').trim();
        continue;
      }
      if (line.startsWith('> Created:')) {
        const metadata = line.replace('> Created:', '').trim();
        // Handle both formats: "date | Updated: date" or just "date"
        if (metadata.includes('|')) {
          const parts = metadata.split('|');
          board.createdAt = parts[0].trim();
          if (parts[1] && parts[1].includes('Updated:')) {
            board.updatedAt = parts[1].replace('Updated:', '').trim();
          }
        } else {
          board.createdAt = metadata;
        }
        continue;
      }
      if (line.startsWith('> Updated:')) {
        // Handle standalone Updated line (for backwards compatibility)
        board.updatedAt = line.replace('> Updated:', '').trim();
        continue;
      }

      // Parse status sections
      if (line.startsWith('## TODO')) {
        // Save current task before changing section
        if (currentTask && currentTask.id) {
          board.tasks.push(currentTask as Task);
          currentTask = null;
        }
        currentStatus = 'todo';
        continue;
      }
      if (line.startsWith('## IN PROGRESS')) {
        // Save current task before changing section
        if (currentTask && currentTask.id) {
          board.tasks.push(currentTask as Task);
          currentTask = null;
        }
        currentStatus = 'in-progress';
        continue;
      }
      if (line.startsWith('## DONE')) {
        // Save current task before changing section
        if (currentTask && currentTask.id) {
          board.tasks.push(currentTask as Task);
          currentTask = null;
        }
        currentStatus = 'done';
        continue;
      }
      if (line.startsWith('## BLOCKED')) {
        // Save current task before changing section
        if (currentTask && currentTask.id) {
          board.tasks.push(currentTask as Task);
          currentTask = null;
        }
        currentStatus = 'blocked';
        continue;
      }

      // Parse tasks
      if (currentStatus && (line.startsWith('- [ ]') || line.startsWith('- [x]') ||
                            line.startsWith('- [>]') || line.startsWith('- [!]'))) {
        // Save previous task if exists
        if (currentTask && currentTask.id) {
          board.tasks.push(currentTask as Task);
        }

        // Parse task line
        const taskMatch = line.match(/^- \[.\] \[(.+?)\] \*\*(.+?)\*\*(.*)$/);
        if (taskMatch) {
          const [, id, title, metadata] = taskMatch;
          currentTask = {
            id,
            title,
            status: currentStatus,
            priority: 'medium',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString(),
          };

          // Parse metadata
          const priorityMatch = metadata.match(/`priority:(\w+)`/);
          if (priorityMatch) {
            currentTask.priority = priorityMatch[1] as TaskPriority;
          }

          const assigneeMatch = metadata.match(/`assignee:([^`]+)`/);
          if (assigneeMatch) {
            currentTask.assignee = assigneeMatch[1];
          }

          const tagsMatch = metadata.match(/`tags:([^`]+)`/);
          if (tagsMatch) {
            currentTask.tags = tagsMatch[1].split(',').map(t => t.trim());
          }

          const dueMatch = metadata.match(/`due:([^`]+)`/);
          if (dueMatch) {
            currentTask.dueDate = dueMatch[1];
          }
        }
        continue;
      }

      // Parse task description
      if (currentTask && line.startsWith('  >') && !line.includes('Created:') && !line.includes('Updated:')) {
        currentTask.description = line.replace('  >', '').trim();
        continue;
      }

      // Parse task timestamps
      if (currentTask && line.includes('Created:') && line.includes('Updated:')) {
        const createdMatch = line.match(/Created: ([^ ]+)/);
        const updatedMatch = line.match(/Updated: ([^ ]+)/);
        if (createdMatch) currentTask.createdAt = createdMatch[1];
        if (updatedMatch) currentTask.updatedAt = updatedMatch[1];
        continue;
      }
    }

    // Save last task
    if (currentTask && currentTask.id) {
      board.tasks.push(currentTask as Task);
    }

    return board;
  }

  /**
   * Convert Board object to markdown format
   */
  private boardToMarkdown(board: Board): string {
    const lines: string[] = [];

    // Board header
    lines.push(`# Board: ${board.name}`);
    if (board.description) {
      lines.push(`> Description: ${board.description}`);
    }
    lines.push(`> Created: ${board.createdAt} | Updated: ${board.updatedAt}`);
    lines.push('');

    // Group tasks by status
    const todoTasks = board.tasks.filter(t => t.status === 'todo');
    const inProgressTasks = board.tasks.filter(t => t.status === 'in-progress');
    const doneTasks = board.tasks.filter(t => t.status === 'done');
    const blockedTasks = board.tasks.filter(t => t.status === 'blocked');

    // TODO section
    lines.push('## TODO');
    lines.push('');
    todoTasks.forEach(task => {
      lines.push(...this.taskToMarkdown(task, '[ ]'));
    });
    if (todoTasks.length === 0) {
      lines.push('_No tasks_');
    }
    lines.push('');

    // IN PROGRESS section
    lines.push('## IN PROGRESS');
    lines.push('');
    inProgressTasks.forEach(task => {
      lines.push(...this.taskToMarkdown(task, '[>]'));
    });
    if (inProgressTasks.length === 0) {
      lines.push('_No tasks_');
    }
    lines.push('');

    // DONE section
    lines.push('## DONE');
    lines.push('');
    doneTasks.forEach(task => {
      lines.push(...this.taskToMarkdown(task, '[x]'));
    });
    if (doneTasks.length === 0) {
      lines.push('_No tasks_');
    }
    lines.push('');

    // BLOCKED section
    lines.push('## BLOCKED');
    lines.push('');
    blockedTasks.forEach(task => {
      lines.push(...this.taskToMarkdown(task, '[!]'));
    });
    if (blockedTasks.length === 0) {
      lines.push('_No tasks_');
    }
    lines.push('');

    return lines.join('\n');
  }

  /**
   * Convert a single task to markdown lines
   */
  private taskToMarkdown(task: Task, checkbox: string): string[] {
    const lines: string[] = [];

    // Build metadata string
    const metadata: string[] = [];
    metadata.push(`\`priority:${task.priority}\``);
    if (task.assignee) metadata.push(`\`assignee:${task.assignee}\``);
    if (task.tags && task.tags.length > 0) metadata.push(`\`tags:${task.tags.join(',')}\``);
    if (task.dueDate) metadata.push(`\`due:${task.dueDate}\``);

    // Task line
    lines.push(`- ${checkbox} [${task.id}] **${task.title}** ${metadata.join(' ')}`);

    // Description
    if (task.description) {
      lines.push(`  > ${task.description}`);
    }

    // Timestamps
    lines.push(`  > Created: ${task.createdAt} | Updated: ${task.updatedAt}`);

    return lines;
  }

  /**
   * Get the file path
   */
  getFilePath(): string {
    return this.filePath;
  }
}
