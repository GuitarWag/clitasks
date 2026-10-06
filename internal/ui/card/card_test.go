package card

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/ui"
)

var now = time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)

func ctx(theme string) Ctx {
	return Ctx{Look: ui.Options{Theme: theme, Icons: "unicode"}.Resolve(true), Now: now}
}

// plain and bracket are styles that show the marks as text in tests.
var (
	plain   = lipgloss.NewStyle()
	bracket = lipgloss.NewStyle().Transform(func(s string) string { return "[" + s + "]" })
)

var long = model.Task{
	ID: "T-1", Title: "Write the Markdown parser for legacy boards and the new format too",
	Status: model.StatusInProgress, Priority: model.PriorityHigh, Assignee: "a-very-long-assignee",
	Tags: []string{"backend"}, DueDate: "2026-10-06", Start: "2026-09-26", After: []string{"T-2"},
	Description: "Parse both formats.",
}

// Every line must be exactly the requested width, or columns and panels
// drift apart in the layout.
func TestCard_exactWidth(t *testing.T) {
	for _, theme := range []string{"nord", "mono"} {
		for _, o := range []Opts{{Width: 32}, {Width: 32, Selected: true}, {Width: 24, Compact: true}, {Width: 18, Waiting: true}} {
			for i, l := range strings.Split(Card(long, ctx(theme), o), "\n") {
				assert.Equal(t, o.Width, ansi.StringWidth(l), "%s %+v line %d", theme, o, i)
			}
		}
	}
	for i, l := range Detail(long, Related{Conflicts: []string{"x"}}, ctx("nord"), 40) {
		assert.Equal(t, 40, ansi.StringWidth(l), "detail line %d", i)
	}
}

func TestCard_content(t *testing.T) {
	out := ansi.Strip(Card(long, ctx("nord"), Opts{Width: 34, Waiting: true}))
	lines := strings.Split(out, "\n")
	assert.Len(t, lines, 4, "two title lines, meta, tags")
	assert.Contains(t, lines[1], "…", "a long title is cut on the second line")
	assert.Contains(t, lines[2], "▲ high")
	assert.Contains(t, lines[2], "⊙ 2d late")
	assert.Contains(t, lines[2], "↪", "the waiting chip survives")
	assert.NotContains(t, lines[2], "@a-very", "the assignee is dropped first")
	assert.Contains(t, lines[3], "#backend")

	compact := ansi.Strip(Card(long, ctx("nord"), Opts{Width: 40, Compact: true}))
	assert.NotContains(t, compact, "\n")
	assert.Contains(t, compact, "2d late")
}

func TestMark(t *testing.T) {
	got := Mark("Parser for the PARSER", []string{"parser"}, plain, bracket)
	assert.Equal(t, "[Parser] for the [PARSER]", got)
	assert.Equal(t, "no match", Mark("no match", []string{"zzz"}, plain, bracket))
}

func TestDetail_content(t *testing.T) {
	dep := model.Task{ID: "T-2", Title: "Design", Status: model.StatusDone}
	out := ansi.Strip(strings.Join(Detail(long, Related{After: []model.Task{dep}, Conflicts: []string{"T-1 starts before T-2 ends"}}, ctx("nord"), 60), "\n"))
	for _, want := range []string{"T-1", "In progress", "Sat Sep 26, 2026", "12 days ago", "Tue Oct 6, 2026", "2 days ago", "Waits for", "● Design", "⚠ T-1 starts before T-2 ends", "Parse both formats."} {
		assert.Contains(t, out, want)
	}
}

func TestCard_showID(t *testing.T) {
	out := ansi.Strip(Card(long, ctx("nord"), Opts{Width: 34, ShowID: true}))
	lines := strings.Split(out, "\n")
	assert.Contains(t, lines[len(lines)-1], "T-1")
}

// Height must match what Card renders, or the board scrolls wrongly.
func TestHeight_matchesCard(t *testing.T) {
	short := model.Task{ID: "T-2", Title: "Short", Priority: model.PriorityLow}
	for _, tk := range []model.Task{long, short} {
		for _, o := range []Opts{{Width: 34}, {Width: 20}, {Width: 34, ShowID: true}, {Width: 30, Compact: true}} {
			assert.Equal(t, len(strings.Split(Card(tk, ctx("nord"), o), "\n")), Height(tk, o), "%s %+v", tk.ID, o)
		}
	}
}
