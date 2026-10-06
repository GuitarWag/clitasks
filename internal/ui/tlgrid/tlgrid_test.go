package tlgrid

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
	"github.com/GuitarWag/clitasks/internal/ui"
	"github.com/GuitarWag/clitasks/internal/ui/card"
)

var now = time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)

func ctx(theme string) card.Ctx {
	return card.Ctx{Look: ui.Options{Theme: theme, Icons: "unicode"}.Resolve(true), Now: now}
}

func layout() timeline.Layout { return timeline.Build(sampleTasks(), now) }

func plain(out Output) []string {
	var s []string
	for _, l := range out.Lines {
		s = append(s, ansi.Strip(l))
	}
	return s
}

func TestRender_everyLineIsExactlyWidth(t *testing.T) {
	l := layout()
	for _, theme := range []string{"catppuccin-mocha", "mono"} {
		for _, z := range []Zoom{ZoomDay, ZoomWeek, ZoomMonth} {
			for _, w := range []int{60, 100, 181} {
				out := Render(l, BarRows(l), ctx(theme), Opts{Width: w, Zoom: z, Selected: 0})
				for i, ln := range out.Lines {
					assert.Equal(t, w, ansi.StringWidth(ln), "%s %s w=%d line %d", theme, z, w, i)
				}
			}
		}
	}
}

func TestAutoZoom(t *testing.T) {
	l := layout() // Sep 21 – Oct 23: 33 days
	assert.Equal(t, ZoomDay, AutoZoom(l, 160))
	assert.Equal(t, ZoomWeek, AutoZoom(l, 80))
	assert.Equal(t, ZoomMonth, AutoZoom(l, 40))
	assert.Equal(t, ZoomDay, AutoZoom(timeline.Layout{}, 40), "an empty board shows days")
}

func TestRender_dayZoom(t *testing.T) {
	l := layout()
	out := Render(l, BarRows(l), ctx("mono"), Opts{Width: 160, Zoom: ZoomDay, Selected: -1})
	lines := plain(out)
	assert.Equal(t, ZoomDay, out.Zoom)
	assert.Equal(t, "2026-09-20", out.From.Format(time.DateOnly), "one day before the first bar")
	assert.True(t, strings.HasPrefix(strings.TrimSpace(lines[0]), "Sep 2026"))
	assert.Contains(t, lines[0], "Oct ")
	assert.Contains(t, lines[1], "20 21 22")

	lw := out.LabelWidth + 1
	design := lines[2]
	assert.Contains(t, design[:lw*3], "Design schema")
	// Mono draws bars with █. The bar covers Sep 21–28: 8 days × 3 columns.
	chart := []rune(design)[lw:]
	assert.Equal(t, strings.Repeat("█", 24), strings.TrimSpace(strings.ReplaceAll(string(chart), "│", " ")))
}

func TestRender_headerLabelsNeverTouch(t *testing.T) {
	l := timeline.Build([]model.Task{{ID: "x", Start: "2026-01-05", DueDate: "2028-03-01"}}, now)
	for _, z := range []Zoom{ZoomWeek, ZoomMonth} {
		out := Render(l, BarRows(l), ctx("mono"), Opts{Width: 140, Zoom: z})
		for _, ln := range plain(out)[:2] {
			for _, f := range strings.Fields(ln) {
				assert.LessOrEqual(t, len(f), len("2026"), "%s header %q has a run-on label %q", z, ln, f)
			}
		}
	}
}

func TestRender_windowAndClip(t *testing.T) {
	l := layout()
	out := Render(l, BarRows(l), ctx("mono"), Opts{Width: 80, Zoom: ZoomDay, Offset: 3, Height: 2})
	require.Len(t, out.Lines, HeaderLines+2)
	assert.Equal(t, 3, out.First)
	assert.Equal(t, 5, out.Last)
	assert.Contains(t, plain(out)[2], "Deploy to product")
	assert.True(t, out.Clipped, "the day zoom does not fit 33 days in 80 columns")

	out = Render(l, BarRows(l), ctx("mono"), Opts{Width: 80, Zoom: ZoomDay, Offset: 99, Height: 2})
	assert.Equal(t, 3, out.First, "the offset is clamped to the last full window")
}

func TestRender_groupRow(t *testing.T) {
	l := layout()
	rows := append([]Row{{Bar: -1, Group: "in progress", Count: 2}}, BarRows(l)...)
	out := Render(l, rows, ctx("mono"), Opts{Width: 80, Zoom: ZoomWeek})
	g := plain(out)[2]
	assert.Contains(t, g, "IN PROGRESS  2")
	assert.Contains(t, g, "────")
}

func TestAlignAndStep(t *testing.T) {
	thu := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, thu, Align(thu, ZoomDay))
	assert.Equal(t, time.Monday, Align(thu, ZoomWeek).Weekday())
	assert.Equal(t, 1, Step(ZoomDay))
	assert.Equal(t, ZoomDay, ZoomMonth.Next())
}

// A chart that starts on the last day of a month still names the next month.
func TestRender_monthLabelAtMonthEnd(t *testing.T) {
	l := timeline.Build([]model.Task{{ID: "x", Start: "2026-06-01", DueDate: "2026-06-10"}}, now)
	out := Render(l, BarRows(l), ctx("mono"), Opts{Width: 120, Zoom: ZoomDay})
	head := plain(out)[0]
	assert.Contains(t, head, "Jun 2026")
	assert.NotContains(t, head, "May 2026", "one day of May has no room for its long label")
}
