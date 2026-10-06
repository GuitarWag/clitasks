package card

import (
	"os"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/ui"
)

// TestPreview writes ANSI previews for each preset when UI_PREVIEW_DIR is set.
// Render them with charmbracelet/freeze to review the design.
func TestPreview(t *testing.T) {
	dir := os.Getenv("UI_PREVIEW_DIR")
	if dir == "" {
		t.Skip()
	}
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)
	for _, th := range []string{"catppuccin-mocha", "catppuccin-latte", "tokyo-night", "nord"} {
		look := ui.Options{Theme: th, Icons: os.Getenv("ICONS")}.Resolve(true)
		c := Ctx{Look: look, Now: now}
		tasks := []model.Task{
			{ID: "T-1", Title: "Write the Markdown parser for legacy boards", Status: model.StatusInProgress, Priority: model.PriorityHigh, Assignee: "bob", Tags: []string{"backend", "parser"}, DueDate: "2026-10-09", Start: "2026-09-26", After: []string{"T-2"}, Description: "Parse both the two-line and the one-line timestamp formats. Keep unknown fields."},
			{ID: "T-2", Title: "Design schema", Status: model.StatusDone, Priority: model.PriorityMedium, DueDate: "2026-10-01"},
			{ID: "T-3", Title: "Deploy to production", Status: model.StatusBlocked, Priority: model.PriorityCritical, Assignee: "alice", DueDate: "2026-10-05"},
		}
		var col []string
		for i, tk := range tasks {
			col = append(col, Card(tk, c, Opts{Width: 32, Selected: i == 0, Highlight: []string{"parser"}, Waiting: i == 0}), lipgloss.NewStyle().Background(look.Theme.Base).Width(32).Render(""))
		}
		col = append(col, Card(tasks[2], c, Opts{Width: 32, Compact: true}))
		left := lipgloss.NewStyle().Background(look.Theme.Base).Render(strings.Join(col, "\n"))
		det := strings.Join(Detail(tasks[0], Related{After: tasks[1:2], Dependents: tasks[2:3], Conflicts: []string{"T-1 starts Sep 26, before T-2 ends Oct 1"}}, c, 46), "\n")
		out := lipgloss.NewStyle().Background(look.Theme.Base).Padding(1, 2).Render(lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", det))
		_ = os.WriteFile(dir+"/card-"+th+".ansi", []byte(out), 0o644)
	}
}
