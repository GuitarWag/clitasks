package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tmpFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "board.md")
}

func fixedClock(s string) func() time.Time {
	t, _ := time.Parse(time.RFC3339Nano, s)
	return func() time.Time { return t }
}

func TestRead_defaultsWhenMissing(t *testing.T) {
	s := NewMarkdown(tmpFile(t)).WithClock(fixedClock("2026-01-01T00:00:00Z"))
	b, err := s.Read()
	require.NoError(t, err)
	assert.Equal(t, "My Board", b.Name)
	assert.Equal(t, "Task management board", b.Description)
	assert.Empty(t, b.Tasks)
}

func TestRead_parsesBoardHeader(t *testing.T) {
	p := tmpFile(t)
	md := strings.Join([]string{
		"# Board: Project Alpha",
		"> Description: Main dev board",
		"> Created: 2026-01-01T00:00:00Z | Updated: 2026-01-02T00:00:00Z",
		"",
		"## TODO",
		"",
		"_No tasks_",
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(p, []byte(md), 0o644))

	b, err := NewMarkdown(p).Read()
	require.NoError(t, err)
	assert.Equal(t, "Project Alpha", b.Name)
	assert.Equal(t, "Main dev board", b.Description)
	assert.Equal(t, "2026-01-01T00:00:00Z", b.CreatedAt.UTC().Format(time.RFC3339))
	assert.Equal(t, "2026-01-02T00:00:00Z", b.UpdatedAt.UTC().Format(time.RFC3339))
}

func TestRead_parsesTasksWithAllFields(t *testing.T) {
	p := tmpFile(t)
	md := strings.Join([]string{
		"# Board: Test",
		"> Created: 2026-01-01T00:00:00Z | Updated: 2026-01-01T00:00:00Z",
		"",
		"## TODO",
		"",
		"- [ ] [T-001] **Fix login bug** `priority:high` `assignee:alice` `tags:backend,auth` `due:2026-03-01`",
		"  > JWT tokens expire too quickly",
		"  > Created: 2026-01-01T00:00:00Z | Updated: 2026-01-02T00:00:00Z",
		"",
		"## IN PROGRESS",
		"",
		"- [>] [T-002] **Add tests** `priority:medium` `assignee:bob`",
		"  > Created: 2026-01-01T00:00:00Z | Updated: 2026-01-01T00:00:00Z",
		"",
		"## DONE",
		"",
		"- [x] [T-003] **Setup CI** `priority:low`",
		"  > Created: 2026-01-01T00:00:00Z | Updated: 2026-01-01T00:00:00Z",
		"",
		"## BLOCKED",
		"",
		"- [!] [T-004] **Deploy** `priority:critical` `tags:devops`",
		"  > Waiting for approval",
		"  > Created: 2026-01-01T00:00:00Z | Updated: 2026-01-01T00:00:00Z",
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(p, []byte(md), 0o644))

	b, err := NewMarkdown(p).Read()
	require.NoError(t, err)
	require.Len(t, b.Tasks, 4)

	todo := b.Tasks[0]
	assert.Equal(t, "T-001", todo.ID)
	assert.Equal(t, "Fix login bug", todo.Title)
	assert.Equal(t, model.StatusTodo, todo.Status)
	assert.Equal(t, model.PriorityHigh, todo.Priority)
	assert.Equal(t, "alice", todo.Assignee)
	assert.Equal(t, []string{"backend", "auth"}, todo.Tags)
	assert.Equal(t, "2026-03-01", todo.DueDate)
	assert.Equal(t, "JWT tokens expire too quickly", todo.Description)

	assert.Equal(t, "T-002", b.Tasks[1].ID)
	assert.Equal(t, model.StatusInProgress, b.Tasks[1].Status)
	assert.Equal(t, "bob", b.Tasks[1].Assignee)

	assert.Equal(t, "T-003", b.Tasks[2].ID)
	assert.Equal(t, model.StatusDone, b.Tasks[2].Status)

	assert.Equal(t, "T-004", b.Tasks[3].ID)
	assert.Equal(t, model.StatusBlocked, b.Tasks[3].Status)
	assert.Equal(t, model.PriorityCritical, b.Tasks[3].Priority)
	assert.Equal(t, []string{"devops"}, b.Tasks[3].Tags)
	assert.Equal(t, "Waiting for approval", b.Tasks[3].Description)
}

func TestRead_parsesMinimalTask(t *testing.T) {
	p := tmpFile(t)
	md := strings.Join([]string{
		"# Board: Test",
		"",
		"## TODO",
		"",
		"- [ ] [T-100] **Quick task** `priority:medium`",
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(p, []byte(md), 0o644))

	b, err := NewMarkdown(p).Read()
	require.NoError(t, err)
	require.Len(t, b.Tasks, 1)
	assert.Equal(t, "T-100", b.Tasks[0].ID)
	assert.Empty(t, b.Tasks[0].Description)
	assert.Empty(t, b.Tasks[0].Assignee)
	assert.Empty(t, b.Tasks[0].Tags)
}

func TestRead_handlesEmptyFile(t *testing.T) {
	p := tmpFile(t)
	require.NoError(t, os.WriteFile(p, []byte(""), 0o644))

	b, err := NewMarkdown(p).Read()
	require.NoError(t, err)
	assert.Equal(t, "My Board", b.Name)
	assert.Empty(t, b.Tasks)
}

func TestRoundtrip_writeThenRead(t *testing.T) {
	p := tmpFile(t)
	now, _ := time.Parse(time.RFC3339, "2026-05-15T12:00:00Z")

	src := &model.Board{
		Name:        "Roundtrip",
		Description: "test",
		CreatedAt:   now,
		UpdatedAt:   now,
		Tasks: []model.Task{
			{
				ID: "T-A", Title: "A", Status: model.StatusTodo,
				Priority: model.PriorityHigh, Assignee: "x",
				Tags: []string{"one", "two"}, DueDate: "2026-06-01",
				Description: "hello", CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "T-B", Title: "B", Status: model.StatusInProgress,
				Priority: model.PriorityMedium, CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "T-C", Title: "C", Status: model.StatusDone,
				Priority: model.PriorityLow, CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "T-D", Title: "D", Status: model.StatusBlocked,
				Priority: model.PriorityCritical, CreatedAt: now, UpdatedAt: now,
			},
		},
	}

	s := NewMarkdown(p)
	require.NoError(t, s.Write(src))

	got, err := s.Read()
	require.NoError(t, err)
	assert.Equal(t, src.Name, got.Name)
	assert.Equal(t, src.Description, got.Description)
	require.Len(t, got.Tasks, 4)
	assert.Equal(t, "T-A", got.Tasks[0].ID)
	assert.Equal(t, []string{"one", "two"}, got.Tasks[0].Tags)
	assert.Equal(t, "hello", got.Tasks[0].Description)
	assert.Equal(t, "2026-06-01", got.Tasks[0].DueDate)
	assert.Equal(t, model.StatusInProgress, got.Tasks[1].Status)
	assert.Equal(t, model.StatusDone, got.Tasks[2].Status)
	assert.Equal(t, model.StatusBlocked, got.Tasks[3].Status)
}

func TestRoundtrip_emptyTasks(t *testing.T) {
	p := tmpFile(t)
	now, _ := time.Parse(time.RFC3339, "2026-05-15T12:00:00Z")
	src := &model.Board{Name: "Empty", CreatedAt: now, UpdatedAt: now}

	s := NewMarkdown(p)
	require.NoError(t, s.Write(src))
	got, err := s.Read()
	require.NoError(t, err)
	assert.Equal(t, "Empty", got.Name)
	assert.Empty(t, got.Tasks)

	data, _ := os.ReadFile(p)
	str := string(data)
	assert.Contains(t, str, "## TODO")
	assert.Contains(t, str, "_No tasks_")
}

func TestAtomicWrite_doesNotLeaveTempFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "board.md")
	s := NewMarkdown(p)
	now, _ := time.Parse(time.RFC3339, "2026-05-15T12:00:00Z")
	require.NoError(t, s.Write(&model.Board{Name: "X", CreatedAt: now, UpdatedAt: now}))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tasks-", "leftover temp file: %s", e.Name())
	}
}

func TestPath_returnsConfigured(t *testing.T) {
	s := NewMarkdown("/foo/bar.md")
	assert.Equal(t, "/foo/bar.md", s.Path())

	s2 := NewMarkdown("")
	assert.Equal(t, DefaultFile, s2.Path())
}
