package export

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleBoard() model.Board {
	now, _ := time.Parse(time.RFC3339, "2026-05-15T12:00:00Z")
	return model.Board{
		Name:        "Sample",
		Description: "test",
		CreatedAt:   now,
		UpdatedAt:   now,
		Tasks: []model.Task{
			{
				ID: "T-1", Title: `with, "quotes"`, Status: model.StatusTodo,
				Priority: model.PriorityHigh, Assignee: "alice",
				Tags: []string{"a", "b"}, DueDate: "2026-06-01",
				CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: "T-2", Title: "second", Status: model.StatusDone,
				Priority: model.PriorityLow, CreatedAt: now, UpdatedAt: now,
			},
		},
	}
}

func TestToJSON_shape(t *testing.T) {
	out, err := ToJSON(sampleBoard())
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(out, &got))

	assert.Equal(t, "Sample", got["name"])
	tasks := got["tasks"].([]any)
	require.Len(t, tasks, 2)
	first := tasks[0].(map[string]any)
	assert.Equal(t, "T-1", first["id"])
	assert.Equal(t, "alice", first["assignee"])
	assert.ElementsMatch(t, []any{"a", "b"}, first["tags"])
	assert.Equal(t, "2026-06-01", first["dueDate"])
}

func TestToJSON_emptyTasksIsArray(t *testing.T) {
	now, _ := time.Parse(time.RFC3339, "2026-05-15T12:00:00Z")
	b := model.Board{Name: "Empty", CreatedAt: now, UpdatedAt: now}
	out, err := ToJSON(b)
	require.NoError(t, err)
	assert.Contains(t, string(out), `"tasks": []`)
}

func TestToCSV_headerAndRows(t *testing.T) {
	out, err := ToCSV(sampleBoard())
	require.NoError(t, err)

	r := csv.NewReader(strings.NewReader(string(out)))
	rows, err := r.ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 3) // header + 2

	assert.Equal(t, []string{"ID", "Title", "Status", "Priority", "Assignee", "Tags", "Due Date", "Created", "Updated"}, rows[0])
	assert.Equal(t, "T-1", rows[1][0])
	assert.Equal(t, `with, "quotes"`, rows[1][1], "CSV decode restores embedded quotes/commas")
	assert.Equal(t, "a;b", rows[1][5])
	assert.Equal(t, "2026-06-01", rows[1][6])
}

func TestToSummary_contents(t *testing.T) {
	out, err := ToSummary(sampleBoard())
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "Board: Sample")
	assert.Contains(t, s, "Description: test")
	assert.Contains(t, s, "Total Tasks: 2")
	assert.Contains(t, s, "TODO: 1")
	assert.Contains(t, s, "DONE: 1")
	assert.Contains(t, s, "IN PROGRESS: 0")
	assert.Contains(t, s, "BLOCKED: 0")
}

func TestRender_dispatch(t *testing.T) {
	b := sampleBoard()
	j, err := Render(b, FormatJSON)
	require.NoError(t, err)
	assert.Contains(t, string(j), `"name"`)

	c, err := Render(b, FormatCSV)
	require.NoError(t, err)
	assert.Contains(t, string(c), "ID,Title")

	s, err := Render(b, FormatSummary)
	require.NoError(t, err)
	assert.Contains(t, string(s), "Board:")

	_, err = Render(b, "xml")
	assert.Error(t, err)
}
