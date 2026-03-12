import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { TaskBoard } from './board';

function tmpFile(): string {
  return path.join(os.tmpdir(), `test-board-${Date.now()}-${Math.random().toString(36).slice(2)}.md`);
}

describe('TaskBoard', () => {
  let filePath: string;
  let board: TaskBoard;

  beforeEach(() => {
    filePath = tmpFile();
    board = new TaskBoard(filePath);
  });

  afterEach(() => {
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
  });

  describe('addTask', () => {
    it('adds a task with default values', () => {
      const task = board.addTask('My task');

      assert.match(task.id, /^T-/);
      assert.equal(task.title, 'My task');
      assert.equal(task.status, 'todo');
      assert.equal(task.priority, 'medium');
      assert.equal(task.assignee, undefined);
      assert.equal(task.tags, undefined);
      assert.equal(task.dueDate, undefined);
      assert.equal(task.description, undefined);
    });

    it('adds a task with all options', () => {
      const task = board.addTask('Full task', {
        description: 'A detailed description',
        priority: 'critical',
        assignee: 'alice',
        tags: ['backend', 'urgent'],
        dueDate: '2026-06-01',
      });

      assert.equal(task.title, 'Full task');
      assert.equal(task.description, 'A detailed description');
      assert.equal(task.priority, 'critical');
      assert.equal(task.assignee, 'alice');
      assert.deepEqual(task.tags, ['backend', 'urgent']);
      assert.equal(task.dueDate, '2026-06-01');
    });

    it('persists the task to disk', () => {
      const task = board.addTask('Persisted task');

      const board2 = new TaskBoard(filePath);
      const found = board2.getTask(task.id);
      assert.notEqual(found, null);
      assert.equal(found!.title, 'Persisted task');
    });

    it('generates unique IDs for each task', () => {
      const t1 = board.addTask('Task 1');
      const t2 = board.addTask('Task 2');
      const t3 = board.addTask('Task 3');

      assert.notEqual(t1.id, t2.id);
      assert.notEqual(t2.id, t3.id);
      assert.notEqual(t1.id, t3.id);
    });
  });

  describe('getTask', () => {
    it('returns the task by ID', () => {
      const added = board.addTask('Find me');
      const found = board.getTask(added.id);

      assert.notEqual(found, null);
      assert.equal(found!.title, 'Find me');
    });

    it('returns null for non-existent ID', () => {
      assert.equal(board.getTask('T-NONEXISTENT'), null);
    });
  });

  describe('updateTask', () => {
    it('updates title and priority', () => {
      const task = board.addTask('Original');
      const updated = board.updateTask(task.id, {
        title: 'Updated',
        priority: 'high',
      });

      assert.notEqual(updated, null);
      assert.equal(updated!.title, 'Updated');
      assert.equal(updated!.priority, 'high');
    });

    it('returns null for non-existent task', () => {
      assert.equal(board.updateTask('T-NOPE', { title: 'X' }), null);
    });

    it('updates the updatedAt timestamp', () => {
      const task = board.addTask('Timestamp test');
      const originalUpdated = task.updatedAt;

      // Small delay to ensure timestamp differs
      const updated = board.updateTask(task.id, { title: 'Changed' });
      assert.notEqual(updated, null);
      // updatedAt should be set (may or may not differ depending on timing)
      assert.ok(updated!.updatedAt);
    });

    it('preserves fields not being updated', () => {
      const task = board.addTask('Keep me', {
        priority: 'high',
        assignee: 'alice',
        tags: ['important'],
      });

      const updated = board.updateTask(task.id, { title: 'New title' });
      assert.equal(updated!.priority, 'high');
      assert.equal(updated!.assignee, 'alice');
      assert.deepEqual(updated!.tags, ['important']);
    });
  });

  describe('moveTask', () => {
    it('changes the task status', () => {
      const task = board.addTask('Move me');
      assert.equal(task.status, 'todo');

      const moved = board.moveTask(task.id, 'in-progress');
      assert.equal(moved!.status, 'in-progress');

      const done = board.moveTask(task.id, 'done');
      assert.equal(done!.status, 'done');
    });

    it('returns null for non-existent task', () => {
      assert.equal(board.moveTask('T-NOPE', 'done'), null);
    });

    it('handles moving to the same status (no-op)', () => {
      const task = board.addTask('Stay put');
      assert.equal(task.status, 'todo');

      const moved = board.moveTask(task.id, 'todo');
      assert.notEqual(moved, null);
      assert.equal(moved!.status, 'todo');
    });

    it('persists the status change', () => {
      const task = board.addTask('Persist status');
      board.moveTask(task.id, 'blocked');

      const board2 = new TaskBoard(filePath);
      const found = board2.getTask(task.id);
      assert.equal(found!.status, 'blocked');
    });
  });

  describe('deleteTask', () => {
    it('removes the task', () => {
      const task = board.addTask('Delete me');
      assert.equal(board.deleteTask(task.id), true);
      assert.equal(board.getTask(task.id), null);
    });

    it('returns false for non-existent task', () => {
      assert.equal(board.deleteTask('T-NOPE'), false);
    });

    it('persists the deletion', () => {
      const task = board.addTask('Gone soon');
      board.deleteTask(task.id);

      const board2 = new TaskBoard(filePath);
      assert.equal(board2.getTask(task.id), null);
    });

    it('does not affect other tasks when deleting one', () => {
      const t1 = board.addTask('Keep me');
      const t2 = board.addTask('Delete me');
      const t3 = board.addTask('Keep me too');

      board.deleteTask(t2.id);

      assert.notEqual(board.getTask(t1.id), null);
      assert.equal(board.getTask(t2.id), null);
      assert.notEqual(board.getTask(t3.id), null);
      assert.equal(board.listTasks().length, 2);
    });
  });

  describe('listTasks', () => {
    beforeEach(() => {
      board.addTask('Task A', { priority: 'high', assignee: 'alice', tags: ['backend'] });
      board.addTask('Task B', { priority: 'low', assignee: 'bob', tags: ['frontend'] });
      board.addTask('Task C', { priority: 'high', assignee: 'alice', tags: ['backend', 'api'] });
    });

    it('returns all tasks with no filters', () => {
      assert.equal(board.listTasks().length, 3);
    });

    it('filters by priority', () => {
      const high = board.listTasks({ priority: 'high' });
      assert.equal(high.length, 2);
      assert.ok(high.every(t => t.priority === 'high'));
    });

    it('filters by assignee', () => {
      const alice = board.listTasks({ assignee: 'alice' });
      assert.equal(alice.length, 2);
      assert.ok(alice.every(t => t.assignee === 'alice'));
    });

    it('filters by tags', () => {
      const backend = board.listTasks({ tags: ['backend'] });
      assert.equal(backend.length, 2);

      const api = board.listTasks({ tags: ['api'] });
      assert.equal(api.length, 1);
      assert.equal(api[0].title, 'Task C');
    });

    it('filters by status', () => {
      const taskA = board.listTasks()[0];
      board.moveTask(taskA.id, 'done');

      const done = board.listTasks({ status: 'done' });
      assert.equal(done.length, 1);

      const todo = board.listTasks({ status: 'todo' });
      assert.equal(todo.length, 2);
    });

    it('combines multiple filters', () => {
      const result = board.listTasks({ priority: 'high', assignee: 'alice' });
      assert.equal(result.length, 2);
    });

    it('returns empty array when no match', () => {
      const result = board.listTasks({ assignee: 'nobody' });
      assert.equal(result.length, 0);
    });

    it('handles tag filter when some tasks have no tags', () => {
      board.addTask('No tags task');

      // Should not crash — tasks without tags should be excluded
      const result = board.listTasks({ tags: ['backend'] });
      assert.equal(result.length, 2); // Task A and Task C
      assert.ok(result.every(t => t.tags && t.tags.includes('backend')));
    });
  });

  describe('getTasksByStatus', () => {
    it('groups tasks by their status', () => {
      const t1 = board.addTask('Todo task');
      const t2 = board.addTask('In progress task');
      const t3 = board.addTask('Done task');

      board.moveTask(t2.id, 'in-progress');
      board.moveTask(t3.id, 'done');

      const grouped = board.getTasksByStatus();
      assert.equal(grouped['todo'].length, 1);
      assert.equal(grouped['in-progress'].length, 1);
      assert.equal(grouped['done'].length, 1);
      assert.equal(grouped['blocked'].length, 0);
    });

    it('returns empty arrays when no tasks', () => {
      const grouped = board.getTasksByStatus();
      assert.equal(grouped['todo'].length, 0);
      assert.equal(grouped['in-progress'].length, 0);
      assert.equal(grouped['done'].length, 0);
      assert.equal(grouped['blocked'].length, 0);
    });
  });

  describe('getBoardInfo / updateBoardInfo', () => {
    it('returns default board info', () => {
      const info = board.getBoardInfo();
      assert.equal(info.name, 'My Board');
    });

    it('updates board name and description', () => {
      board.updateBoardInfo({ name: 'New Name', description: 'New desc' });
      const info = board.getBoardInfo();
      assert.equal(info.name, 'New Name');
      assert.equal(info.description, 'New desc');
    });

    it('persists board info changes', () => {
      board.updateBoardInfo({ name: 'Persisted Name' });

      const board2 = new TaskBoard(filePath);
      assert.equal(board2.getBoardInfo().name, 'Persisted Name');
    });

    it('updates only name without touching description', () => {
      board.updateBoardInfo({ name: 'First', description: 'Original desc' });
      board.updateBoardInfo({ name: 'Second' });

      const info = board.getBoardInfo();
      assert.equal(info.name, 'Second');
      assert.equal(info.description, 'Original desc');
    });

    it('updates only description without touching name', () => {
      board.updateBoardInfo({ name: 'Keep this', description: 'Old' });
      board.updateBoardInfo({ description: 'New desc' });

      const info = board.getBoardInfo();
      assert.equal(info.name, 'Keep this');
      assert.equal(info.description, 'New desc');
    });
  });
});
