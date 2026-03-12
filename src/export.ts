import { TaskBoard } from './board';

export class Exporter {
  constructor(private board: TaskBoard) {}

  /**
   * Export board to JSON
   */
  toJSON(): string {
    const info = this.board.getBoardInfo();
    return JSON.stringify(info, null, 2);
  }

  /**
   * Export tasks to CSV
   */
  toCSV(): string {
    const tasks = this.board.listTasks();
    const headers = ['ID', 'Title', 'Status', 'Priority', 'Assignee', 'Tags', 'Due Date', 'Created', 'Updated'];
    const rows = tasks.map(task => [
      task.id,
      `"${task.title.replace(/"/g, '""')}"`,
      task.status,
      task.priority,
      task.assignee || '',
      task.tags ? task.tags.join(';') : '',
      task.dueDate || '',
      task.createdAt,
      task.updatedAt,
    ]);

    return [headers.join(','), ...rows.map(row => row.join(','))].join('\n');
  }

  /**
   * Export summary
   */
  toSummary(): string {
    const info = this.board.getBoardInfo();
    const tasksByStatus = this.board.getTasksByStatus();

    const lines = [
      `Board: ${info.name}`,
      info.description ? `Description: ${info.description}` : '',
      `Total Tasks: ${info.tasks.length}`,
      '',
      'Status Breakdown:',
      `  TODO: ${tasksByStatus['todo'].length}`,
      `  IN PROGRESS: ${tasksByStatus['in-progress'].length}`,
      `  DONE: ${tasksByStatus['done'].length}`,
      `  BLOCKED: ${tasksByStatus['blocked'].length}`,
    ];

    return lines.filter(l => l !== '').join('\n');
  }
}
