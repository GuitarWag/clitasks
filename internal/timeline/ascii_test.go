package timeline

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/internal/model"
)

// 2026-06-01 is a Monday.
func dayScaleLayout() Layout {
	return Build([]model.Task{
		task("T-1", func(t *model.Task) { t.Title = "Design"; t.Start = "2026-06-01"; t.DueDate = "2026-06-05" }),
		task("T-2", func(t *model.Task) { t.Title = "Build the parser"; t.Start = "2026-06-08"; t.DueDate = "2026-06-12" }),
		task("T-3", func(t *model.Task) { t.Title = "Open"; t.Start = "2026-06-10" }),
	}, today)
}

func TestRenderASCII_dayScale(t *testing.T) {
	out := RenderASCII(dayScaleLayout(), ASCIIOptions{Width: 50, Today: today})
	assert.Equal(t, ScaleDay, out.Scale)
	assert.False(t, out.Clipped)
	got := strings.Join(append([]string{out.Header}, out.Rows...), "\n")
	want := "" +
		"                 06-01  06-08  06-15  06-22  06-29\n" +
		"T-1 Design       █████         |\n" +
		"T-2 Build the p… .......█████  |\n" +
		"T-3 Open                  █████>"
	want = strings.ReplaceAll(want, ".", " ")
	assert.Equal(t, want, got)
}

func TestRenderASCII_weekScaleWhenRangeIsWide(t *testing.T) {
	l := Build([]model.Task{
		task("T-1", func(t *model.Task) { t.Start = "2026-06-03"; t.DueDate = "2026-08-20" }),
	}, today)
	out := RenderASCII(l, ASCIIOptions{Width: 40, Today: today})
	require.Equal(t, ScaleWeek, out.Scale)
	assert.Equal(t, date("2026-06-01"), out.From, "week scale starts on Monday")
	// June 1 to Aug 20 covers 12 ISO weeks.
	assert.Equal(t, strings.Repeat("█", 12), strings.TrimSpace(out.Rows[0][14:]))
	assert.False(t, out.Clipped)
}

func TestRenderASCII_clipsAndScrolls(t *testing.T) {
	l := Build([]model.Task{
		task("T-1", func(t *model.Task) { t.Start = "2026-01-05"; t.DueDate = "2026-12-28" }),
	}, today)
	out := RenderASCII(l, ASCIIOptions{Width: 30, Today: today})
	assert.Equal(t, ScaleWeek, out.Scale)
	assert.True(t, out.Clipped)

	scrolled := RenderASCII(l, ASCIIOptions{Width: 30, Today: today, Scale: ScaleDay, From: noon("2026-12-20")})
	assert.Equal(t, ScaleDay, scrolled.Scale, "a given scale is kept")
	assert.Equal(t, date("2026-12-20"), scrolled.From)
	assert.Contains(t, scrolled.Rows[0], "█████████")
}

func TestRenderASCII_empty(t *testing.T) {
	out := RenderASCII(Layout{}, ASCIIOptions{Width: 40, Today: today})
	assert.Empty(t, out.Rows)
	assert.Equal(t, date("2026-06-15"), out.From)
}

// The TUI passes a UTC-midnight date as From. West of UTC that must not
// become the previous day (run with TZ=America/New_York).
func TestRenderASCII_fromIsUsedAsADate(t *testing.T) {
	l := dayScaleLayout()
	out := RenderASCII(l, ASCIIOptions{Width: 50, Today: today, Scale: ScaleWeek, From: date("2026-06-08")})
	assert.Equal(t, date("2026-06-08"), out.From)
	out = RenderASCII(l, ASCIIOptions{Width: 50, Today: today, Scale: ScaleDay, From: date("2026-06-03")})
	assert.Equal(t, date("2026-06-03"), out.From)
}
