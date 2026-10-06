package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/v3/internal/board"
	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/storage"
	"github.com/GuitarWag/clitasks/v3/internal/ui"
)

// demoBoard writes a realistic board and returns a model over it, at
// Thursday 2026-10-08 noon.
func demoBoard(t *testing.T, theme string, w, h int) Model {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tasks.md")
	at := func(s string) time.Time {
		d, _ := time.ParseInLocation(time.DateOnly, s, time.Local)
		return d.Add(12 * time.Hour)
	}
	tasks := []model.Task{
		{ID: "T-MV1A2B-D01", Title: "Design the storage schema", Status: model.StatusDone, Priority: model.PriorityHigh, Assignee: "alice", Tags: []string{"backend"}, Start: "2026-09-21", DueDate: "2026-09-28", Description: "Markdown sections per status, one task per list item.", CreatedAt: at("2026-09-20"), UpdatedAt: at("2026-09-28")},
		{ID: "T-MV1A2B-P02", Title: "Write the Markdown parser for legacy boards", Status: model.StatusInProgress, Priority: model.PriorityHigh, Assignee: "bob", Tags: []string{"backend", "parser"}, Start: "2026-09-26", DueDate: "2026-10-09", After: []string{"T-MV1A2B-D01"}, Description: "Parse both the two-line and the one-line timestamp formats. Keep unknown fields.", CreatedAt: at("2026-09-20"), UpdatedAt: at("2026-10-06")},
		{ID: "T-MV1A2B-T03", Title: "Ship the TUI timeline view", Status: model.StatusInProgress, Priority: model.PriorityMedium, Assignee: "wagner", Tags: []string{"tui"}, Start: "2026-10-04", After: []string{"T-MV1A2B-P02"}, CreatedAt: at("2026-09-22"), UpdatedAt: at("2026-10-07")},
		{ID: "T-MV1A2B-X04", Title: "Deploy to production", Status: model.StatusBlocked, Priority: model.PriorityCritical, Assignee: "alice", Tags: []string{"ops"}, Start: "2026-10-12", DueDate: "2026-10-16", After: []string{"T-MV1A2B-T03"}, CreatedAt: at("2026-09-22"), UpdatedAt: at("2026-10-01")},
		{ID: "T-MV1A2B-W05", Title: "Write the user docs", Status: model.StatusTodo, Priority: model.PriorityLow, Tags: []string{"docs"}, Start: "2026-10-13", DueDate: "2026-10-23", CreatedAt: at("2026-09-22"), UpdatedAt: at("2026-09-22")},
		{ID: "T-MV1A2B-R06", Title: "Fix the flaky CSV export test", Status: model.StatusTodo, Priority: model.PriorityHigh, Assignee: "bob", Tags: []string{"tests"}, DueDate: "2026-10-06", CreatedAt: at("2026-09-30"), UpdatedAt: at("2026-09-30")},
		{ID: "T-MV1A2B-M07", Title: "Mermaid export for GitHub READMEs", Status: model.StatusTodo, Priority: model.PriorityMedium, Assignee: "wagner", DueDate: "2026-10-10", CreatedAt: at("2026-10-01"), UpdatedAt: at("2026-10-01")},
		{ID: "T-MV1A2B-C08", Title: "Command palette", Status: model.StatusDone, Priority: model.PriorityMedium, Assignee: "wagner", Start: "2026-09-29", DueDate: "2026-10-02", CreatedAt: at("2026-09-29"), UpdatedAt: at("2026-10-02")},
		{ID: "T-MV1A2B-L09", Title: "Live reload when an agent edits tasks.md", Status: model.StatusTodo, Priority: model.PriorityMedium, Tags: []string{"tui"}, DueDate: "next sprint", CreatedAt: at("2026-10-02"), UpdatedAt: at("2026-10-02")},
	}
	require.NoError(t, storage.NewMarkdown(p).Write(&model.Board{Name: "Sprint 12", Description: "clitasks v3", Tasks: tasks, CreatedAt: at("2026-09-20"), UpdatedAt: at("2026-10-07")}))
	b, err := board.Open(storage.NewMarkdown(p), board.WithClock(func() time.Time { return at("2026-10-08") }))
	require.NoError(t, err)
	m := newModel(b, p, ui.Options{Theme: theme, Icons: "unicode"})
	m.now = func() time.Time { return at("2026-10-08") }
	mm, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return mm.(Model)
}

func TestPreview(t *testing.T) {
	dir := os.Getenv("UI_PREVIEW_DIR")
	if dir == "" {
		t.Skip()
	}
	shot := func(name string, m Model) {
		bg := m.look.Theme.Base
		out := lipgloss.NewStyle().Background(bg).Render(m.render())
		require.NoError(t, os.WriteFile(filepath.Join(dir, name+".ansi"), []byte(out), 0o644))
	}
	for _, th := range []string{"catppuccin-mocha", "catppuccin-latte", "tokyo-night", "nord"} {
		for _, sz := range [][2]int{{160, 44}, {120, 36}, {90, 28}} {
			m := demoBoard(t, th, sz[0], sz[1])
			m = send(m, "j")
			shot(fmt.Sprintf("board-%s-%dx%d", th, sz[0], sz[1]), m)
			tl := send(m, "tab")
			shot(fmt.Sprintf("timeline-%s-%dx%d", th, sz[0], sz[1]), tl)
		}
	}
	m := demoBoard(t, "catppuccin-mocha", 140, 42)
	m = send(m, "l")
	shot("form-edit", send(m, "e"))
	shot("palette", send(m, "ctrl+k"))
	shot("help", send(m, "?"))
	shot("status", send(m, "space"))
	shot("confirm", send(m, "d"))
	s := send(m, "/")
	for _, r := range "parser" {
		s = send(s, string(r))
	}
	shot("search", s)
}
