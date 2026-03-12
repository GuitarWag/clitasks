export type TaskStatus = 'todo' | 'in-progress' | 'done' | 'blocked';
export type TaskPriority = 'low' | 'medium' | 'high' | 'critical';

export interface Task {
  id: string;
  title: string;
  description?: string;
  status: TaskStatus;
  priority: TaskPriority;
  assignee?: string;
  tags?: string[];
  createdAt: string;
  updatedAt: string;
  dueDate?: string;
}

export interface Board {
  name: string;
  description?: string;
  tasks: Task[];
  createdAt: string;
  updatedAt: string;
}
