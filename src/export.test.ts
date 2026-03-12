import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { TaskBoard } from './board';
import { Exporter } from './export';

function tmpFile(): string {
  return path.join(os.tmpdir(), `test-board-${Date.now()}-${Math.random().toString(36).slice(2)}.md`);
}

describe('Exporter', () => {
  let filePath: string;
  let board: TaskBoard;
  let exporter: Exporter;

  beforeEach(() => {
    filePath = tmpFile();
    board = new TaskBoard(filePath);

    board.addTask('Auth feature', {
      priority: 'high',
      assignee: 'alice',
      tags: ['backend', 'security'],
      dueDate: '2026-06-01',
      description: 'Add JWT auth',
    });
    board.addTask('Write docs', { priority: 'low', assignee: 'bob' });

    const tasks = board.listTasks();
    board.moveTask(tasks[1].id, 'done');

    exporter = new Exporter(board);
  });

  afterEach(() => {
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
  });

  describe('toJSON', () => {
    it('returns valid JSON', () => {
      const json = exporter.toJSON();
      const parsed = JSON.parse(json);

      assert.equal(typeof parsed.name, 'string');
      assert.ok(Array.isArray(parsed.tasks));
      assert.equal(parsed.tasks.length, 2);
    });

    it('includes all task fields', () => {
      const parsed = JSON.parse(exporter.toJSON());
      const task = parsed.tasks.find((t: any) => t.title === 'Auth feature');

      assert.equal(task.priority, 'high');
      assert.equal(task.assignee, 'alice');
      assert.deepEqual(task.tags, ['backend', 'security']);
      assert.equal(task.dueDate, '2026-06-01');
    });
  });

  describe('toCSV', () => {
    it('includes header row', () => {
      const csv = exporter.toCSV();
      const lines = csv.split('\n');

      assert.equal(lines[0], 'ID,Title,Status,Priority,Assignee,Tags,Due Date,Created,Updated');
    });

    it('includes all tasks as rows', () => {
      const csv = exporter.toCSV();
      const lines = csv.split('\n');

      // header + 2 tasks
      assert.equal(lines.length, 3);
    });

    it('escapes double quotes in titles', () => {
      board.addTask('Task with "quotes"', { priority: 'medium' });
      exporter = new Exporter(board);

      const csv = exporter.toCSV();
      assert.ok(csv.includes('""quotes""'));
    });

    it('joins tags with semicolons', () => {
      const csv = exporter.toCSV();
      assert.ok(csv.includes('backend;security'));
    });

    it('handles commas in title', () => {
      board.addTask('Task with, commas, inside', { priority: 'medium' });
      exporter = new Exporter(board);

      const csv = exporter.toCSV();
      // Title should be wrapped in quotes
      assert.ok(csv.includes('"Task with, commas, inside"'));
    });

    it('outputs empty fields for missing optional values', () => {
      // "Write docs" task has no tags, no due date
      const csv = exporter.toCSV();
      const lines = csv.split('\n');
      // Find the "Write docs" row
      const docsRow = lines.find(l => l.includes('Write docs'));
      assert.ok(docsRow);
      // Should have empty tag and due date fields (consecutive commas)
      const fields = docsRow!.split(',');
      // Tags field (index 5) and Due Date field (index 6) should be empty
      assert.equal(fields[5], ''); // tags
      assert.equal(fields[6], ''); // due date
    });
  });

  describe('toSummary', () => {
    it('includes board name', () => {
      const summary = exporter.toSummary();
      assert.ok(summary.includes('Board: My Board'));
    });

    it('includes total task count', () => {
      const summary = exporter.toSummary();
      assert.ok(summary.includes('Total Tasks: 2'));
    });

    it('includes status breakdown', () => {
      const summary = exporter.toSummary();
      assert.ok(summary.includes('TODO: 1'));
      assert.ok(summary.includes('DONE: 1'));
      assert.ok(summary.includes('IN PROGRESS: 0'));
      assert.ok(summary.includes('BLOCKED: 0'));
    });
  });

  describe('empty board', () => {
    it('toJSON returns valid JSON with empty tasks array', () => {
      const emptyPath = tmpFile();
      const emptyBoard = new TaskBoard(emptyPath);
      const emptyExporter = new Exporter(emptyBoard);

      const parsed = JSON.parse(emptyExporter.toJSON());
      assert.ok(Array.isArray(parsed.tasks));
      assert.equal(parsed.tasks.length, 0);

      if (fs.existsSync(emptyPath)) fs.unlinkSync(emptyPath);
    });

    it('toCSV returns only the header row', () => {
      const emptyPath = tmpFile();
      const emptyBoard = new TaskBoard(emptyPath);
      const emptyExporter = new Exporter(emptyBoard);

      const csv = emptyExporter.toCSV();
      const lines = csv.split('\n');
      assert.equal(lines.length, 1);
      assert.ok(lines[0].startsWith('ID,'));

      if (fs.existsSync(emptyPath)) fs.unlinkSync(emptyPath);
    });

    it('toSummary shows zero total tasks', () => {
      const emptyPath = tmpFile();
      const emptyBoard = new TaskBoard(emptyPath);
      const emptyExporter = new Exporter(emptyBoard);

      const summary = emptyExporter.toSummary();
      assert.ok(summary.includes('Total Tasks: 0'));

      if (fs.existsSync(emptyPath)) fs.unlinkSync(emptyPath);
    });
  });
});
