package tlgrid

import (
	"os"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
	"github.com/GuitarWag/clitasks/internal/ui"
	"github.com/GuitarWag/clitasks/internal/ui/card"
)

func sampleTasks() []model.Task {
	at := func(s string) time.Time {
		d, _ := time.ParseInLocation(time.DateOnly, s, time.Local)
		return d.Add(12 * time.Hour)
	}
	return []model.Task{
		{ID: "T-1", Title: "Design schema", Status: model.StatusDone, Start: "2026-09-21", DueDate: "2026-09-28", CreatedAt: at("2026-09-20"), UpdatedAt: at("2026-09-28")},
		{ID: "T-2", Title: "Write the Markdown parser", Status: model.StatusInProgress, Start: "2026-09-26", DueDate: "2026-10-09", After: []string{"T-1"}, CreatedAt: at("2026-09-20")},
		{ID: "T-3", Title: "Ship the TUI view", Status: model.StatusInProgress, Start: "2026-10-04", After: []string{"T-2"}, CreatedAt: at("2026-09-22")},
		{ID: "T-4", Title: "Deploy to production", Status: model.StatusBlocked, Start: "2026-10-12", DueDate: "2026-10-16", After: []string{"T-3"}, CreatedAt: at("2026-09-22")},
		{ID: "T-5", Title: "Write the docs", Status: model.StatusTodo, Start: "2026-10-13", DueDate: "2026-10-23", CreatedAt: at("2026-09-22")},
	}
}

// TestPreview writes ANSI previews when UI_PREVIEW_DIR is set.
func TestPreview(t *testing.T) {
	dir := os.Getenv("UI_PREVIEW_DIR")
	if dir == "" {
		t.Skip()
	}
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)
	for _, th := range []string{"catppuccin-mocha", "catppuccin-latte"} {
		look := ui.Options{Theme: th, Icons: "unicode"}.Resolve(true)
		c := card.Ctx{Look: look, Now: now}
		l := timeline.Build(sampleTasks(), now)
		var parts []string
		for _, z := range []Zoom{ZoomDay, ZoomWeek, ZoomMonth} {
			out := Render(l, BarRows(l), c, Opts{Width: 120, Zoom: z, Selected: 1,
				Related: map[string]bool{"T-1": true, "T-3": true}, Conflicts: map[string]bool{"T-2": true}})
			parts = append(parts, strings.Join(out.Lines, "\n"), "")
		}
		s := lipgloss.NewStyle().Background(look.Theme.Base).Padding(1, 1).Render(strings.Join(parts, "\n"))
		_ = os.WriteFile(dir+"/tl-"+th+".ansi", []byte(s), 0o644)
	}
}
