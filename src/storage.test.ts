import { describe, it, beforeEach, afterEach } from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';
import { MarkdownStorage } from './storage';

function tmpFile(): string {
  return path.join(os.tmpdir(), `test-board-${Date.now()}-${Math.random().toString(36).slice(2)}.md`);
}

describe('MarkdownStorage', () => {
  let filePath: string;

  beforeEach(() => {
    filePath = tmpFile();
  });

  afterEach(() => {
    if (fs.existsSync(filePath)) fs.unlinkSync(filePath);
  });

  describe('readBoard', () => {
    it('returns a default board when file does not exist', () => {
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.name, 'My Board');
      assert.equal(board.description, 'Task management board');
      assert.deepEqual(board.tasks, []);
    });

    it('parses board header and metadata', () => {
      const md = [
        '# Board: Project Alpha',
        '> Description: Main dev board',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-02T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '_No tasks_',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.name, 'Project Alpha');
      assert.equal(board.description, 'Main dev board');
      assert.equal(board.createdAt, '2026-01-01T00:00:00.000Z');
      assert.equal(board.updatedAt, '2026-01-02T00:00:00.000Z');
    });

    it('parses tasks with all metadata fields', () => {
      const md = [
        '# Board: Test',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '- [ ] [T-001] **Fix login bug** `priority:high` `assignee:alice` `tags:backend,auth` `due:2026-03-01`',
        '  > JWT tokens expire too quickly',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-02T00:00:00.000Z',
        '',
        '## IN PROGRESS',
        '',
        '- [>] [T-002] **Add tests** `priority:medium` `assignee:bob`',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## DONE',
        '',
        '- [x] [T-003] **Setup CI** `priority:low`',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## BLOCKED',
        '',
        '- [!] [T-004] **Deploy** `priority:critical` `tags:devops`',
        '  > Waiting for approval',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.tasks.length, 4);

      const todo = board.tasks[0];
      assert.equal(todo.id, 'T-001');
      assert.equal(todo.title, 'Fix login bug');
      assert.equal(todo.status, 'todo');
      assert.equal(todo.priority, 'high');
      assert.equal(todo.assignee, 'alice');
      assert.deepEqual(todo.tags, ['backend', 'auth']);
      assert.equal(todo.dueDate, '2026-03-01');
      assert.equal(todo.description, 'JWT tokens expire too quickly');

      const inProgress = board.tasks[1];
      assert.equal(inProgress.id, 'T-002');
      assert.equal(inProgress.status, 'in-progress');
      assert.equal(inProgress.assignee, 'bob');

      const done = board.tasks[2];
      assert.equal(done.id, 'T-003');
      assert.equal(done.status, 'done');
      assert.equal(done.priority, 'low');

      const blocked = board.tasks[3];
      assert.equal(blocked.id, 'T-004');
      assert.equal(blocked.status, 'blocked');
      assert.equal(blocked.priority, 'critical');
      assert.equal(blocked.description, 'Waiting for approval');
    });

    it('handles an empty file gracefully', () => {
      fs.writeFileSync(filePath, '');
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.name, 'My Board');
      assert.deepEqual(board.tasks, []);
    });

    it('handles a file with only whitespace', () => {
      fs.writeFileSync(filePath, '   \n\n  \n');
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.deepEqual(board.tasks, []);
    });

    it('parses standalone Updated line for backwards compatibility', () => {
      const md = [
        '# Board: Legacy',
        '> Created: 2026-01-01T00:00:00.000Z',
        '> Updated: 2026-02-15T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '_No tasks_',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.createdAt, '2026-01-01T00:00:00.000Z');
      assert.equal(board.updatedAt, '2026-02-15T00:00:00.000Z');
    });

    it('parses a task with no metadata (just ID and title)', () => {
      const md = [
        '# Board: Minimal',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '- [ ] [T-BARE] **Bare task**',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.tasks.length, 1);
      assert.equal(board.tasks[0].id, 'T-BARE');
      assert.equal(board.tasks[0].title, 'Bare task');
      assert.equal(board.tasks[0].priority, 'medium'); // default
      assert.equal(board.tasks[0].assignee, undefined);
      assert.equal(board.tasks[0].tags, undefined);
      assert.equal(board.tasks[0].dueDate, undefined);
    });

    it('parses description with special characters', () => {
      const md = [
        '# Board: Special',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '- [ ] [T-SP] **Special chars** `priority:high`',
        '  > Description with `backticks` and pipes | and > arrows',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.tasks[0].description, 'Description with `backticks` and pipes | and > arrows');
    });

    it('parses board with no description line', () => {
      const md = [
        '# Board: No Desc',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '_No tasks_',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.name, 'No Desc');
      assert.equal(board.description, undefined);
    });

    it('parses multiple tasks in the same section', () => {
      const md = [
        '# Board: Test',
        '> Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## TODO',
        '',
        '- [ ] [T-A] **Task A** `priority:high`',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '- [ ] [T-B] **Task B** `priority:low`',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '- [ ] [T-C] **Task C** `priority:medium`',
        '  > Created: 2026-01-01T00:00:00.000Z | Updated: 2026-01-01T00:00:00.000Z',
        '',
        '## IN PROGRESS',
        '',
        '_No tasks_',
        '',
        '## DONE',
        '',
        '_No tasks_',
        '',
        '## BLOCKED',
        '',
        '_No tasks_',
        '',
      ].join('\n');

      fs.writeFileSync(filePath, md);
      const storage = new MarkdownStorage(filePath);
      const board = storage.readBoard();

      assert.equal(board.tasks.length, 3);
      assert.equal(board.tasks[0].id, 'T-A');
      assert.equal(board.tasks[1].id, 'T-B');
      assert.equal(board.tasks[2].id, 'T-C');
    });
  });

  describe('writeBoard', () => {
    it('writes a board and reads it back identically', () => {
      const storage = new MarkdownStorage(filePath);
      const original = {
        name: 'Roundtrip Test',
        description: 'Testing write/read roundtrip',
        tasks: [
          {
            id: 'T-RT1',
            title: 'First task',
            description: 'With a description',
            status: 'todo' as const,
            priority: 'high' as const,
            assignee: 'alice',
            tags: ['backend', 'api'],
            dueDate: '2026-06-01',
            createdAt: '2026-01-01T00:00:00.000Z',
            updatedAt: '2026-01-02T00:00:00.000Z',
          },
          {
            id: 'T-RT2',
            title: 'Second task',
            status: 'done' as const,
            priority: 'low' as const,
            createdAt: '2026-01-01T00:00:00.000Z',
            updatedAt: '2026-01-03T00:00:00.000Z',
          },
        ],
        createdAt: '2026-01-01T00:00:00.000Z',
        updatedAt: '2026-01-03T00:00:00.000Z',
      };

      storage.writeBoard(original);
      const parsed = storage.readBoard();

      assert.equal(parsed.name, original.name);
      assert.equal(parsed.description, original.description);
      assert.equal(parsed.tasks.length, 2);

      assert.equal(parsed.tasks[0].id, 'T-RT1');
      assert.equal(parsed.tasks[0].title, 'First task');
      assert.equal(parsed.tasks[0].description, 'With a description');
      assert.equal(parsed.tasks[0].status, 'todo');
      assert.equal(parsed.tasks[0].priority, 'high');
      assert.equal(parsed.tasks[0].assignee, 'alice');
      assert.deepEqual(parsed.tasks[0].tags, ['backend', 'api']);
      assert.equal(parsed.tasks[0].dueDate, '2026-06-01');

      assert.equal(parsed.tasks[1].id, 'T-RT2');
      assert.equal(parsed.tasks[1].status, 'done');
      assert.equal(parsed.tasks[1].priority, 'low');
      assert.equal(parsed.tasks[1].description, undefined);
    });

    it('writes empty sections with _No tasks_ placeholder', () => {
      const storage = new MarkdownStorage(filePath);
      storage.writeBoard({
        name: 'Empty Board',
        tasks: [],
        createdAt: '2026-01-01T00:00:00.000Z',
        updatedAt: '2026-01-01T00:00:00.000Z',
      });

      const content = fs.readFileSync(filePath, 'utf-8');
      assert.equal((content.match(/_No tasks_/g) || []).length, 4);
    });
  });

  describe('getFilePath', () => {
    it('returns the configured file path', () => {
      const storage = new MarkdownStorage(filePath);
      assert.equal(storage.getFilePath(), filePath);
    });

    it('defaults to tasks.md', () => {
      const storage = new MarkdownStorage();
      assert.equal(storage.getFilePath(), 'tasks.md');
    });
  });
});
