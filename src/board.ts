import { Board, Task, TaskStatus, TaskPriority } from './types';
import { MarkdownStorage } from './storage';

export class TaskBoard {
  private storage: MarkdownStorage;
  private board: Board;

  constructor(filePath?: string) {
    this.storage = new MarkdownStorage(filePath);
    this.board = this.storage.readBoard();
  }

  /**
   * Add a new task
   */
  addTask(
    title: string,
    options: {
      description?: string;
      priority?: TaskPriority;
      assignee?: string;
      tags?: string[];
      dueDate?: string;
    } = {}
  ): Task {
    const now = new Date().toISOString();
    const id = this.generateTaskId();

    const task: Task = {
      id,
      title,
      description: options.description,
      status: 'todo',
      priority: options.priority || 'medium',
      assignee: options.assignee,
      tags: options.tags,
      createdAt: now,
      updatedAt: now,
      dueDate: options.dueDate,
    };

    this.board.tasks.push(task);
    this.board.updatedAt = now;
    this.save();

    return task;
  }

  /**
   * Update a task
   */
  updateTask(
    id: string,
    updates: Partial<Omit<Task, 'id' | 'createdAt' | 'updatedAt'>>
  ): Task | null {
    const task = this.board.tasks.find(t => t.id === id);
    if (!task) return null;

    Object.assign(task, updates, { updatedAt: new Date().toISOString() });
    this.board.updatedAt = new Date().toISOString();
    this.save();

    return task;
  }

  /**
   * Move task to different status
   */
  moveTask(id: string, status: TaskStatus): Task | null {
    return this.updateTask(id, { status });
  }

  /**
   * Delete a task
   */
  deleteTask(id: string): boolean {
    const index = this.board.tasks.findIndex(t => t.id === id);
    if (index === -1) return false;

    this.board.tasks.splice(index, 1);
    this.board.updatedAt = new Date().toISOString();
    this.save();

    return true;
  }

  /**
   * Get a task by ID
   */
  getTask(id: string): Task | null {
    return this.board.tasks.find(t => t.id === id) || null;
  }

  /**
   * List tasks with optional filters
   */
  listTasks(filters: {
    status?: TaskStatus;
    priority?: TaskPriority;
    assignee?: string;
    tags?: string[];
  } = {}): Task[] {
    let tasks = [...this.board.tasks];

    if (filters.status) {
      tasks = tasks.filter(t => t.status === filters.status);
    }
    if (filters.priority) {
      tasks = tasks.filter(t => t.priority === filters.priority);
    }
    if (filters.assignee) {
      tasks = tasks.filter(t => t.assignee === filters.assignee);
    }
    if (filters.tags && filters.tags.length > 0) {
      tasks = tasks.filter(t =>
        t.tags && filters.tags!.some(tag => t.tags!.includes(tag))
      );
    }

    return tasks;
  }

  /**
   * Get all tasks grouped by status
   */
  getTasksByStatus(): Record<TaskStatus, Task[]> {
    return {
      'todo': this.board.tasks.filter(t => t.status === 'todo'),
      'in-progress': this.board.tasks.filter(t => t.status === 'in-progress'),
      'done': this.board.tasks.filter(t => t.status === 'done'),
      'blocked': this.board.tasks.filter(t => t.status === 'blocked'),
    };
  }

  /**
   * Get board info
   */
  getBoardInfo(): Board {
    return { ...this.board };
  }

  /**
   * Update board info
   */
  updateBoardInfo(updates: { name?: string; description?: string }): void {
    if (updates.name) this.board.name = updates.name;
    if (updates.description) this.board.description = updates.description;
    this.board.updatedAt = new Date().toISOString();
    this.save();
  }

  /**
   * Get file path
   */
  getFilePath(): string {
    return this.storage.getFilePath();
  }

  /**
   * Save the board to storage
   */
  private save(): void {
    this.storage.writeBoard(this.board);
  }

  /**
   * Generate a unique task ID
   */
  private generateTaskId(): string {
    const prefix = 'T';
    const timestamp = Date.now().toString(36).toUpperCase();
    const random = Math.random().toString(36).substring(2, 5).toUpperCase();
    return `${prefix}-${timestamp}-${random}`;
  }
}
