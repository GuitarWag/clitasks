// Package tlgrid draws a timeline.Layout as a Gantt chart: a label column, a
// two-row date header and one row per bar or group. Each chart cell has a
// semantic kind (bar, today, weekend, open end) that the theme colors.
package tlgrid

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
	"github.com/GuitarWag/clitasks/internal/ui/card"
)

// Zoom is the time scale of the chart.
type Zoom int

const (
	ZoomAuto  Zoom = iota
	ZoomDay        // 3 columns per day, day numbers in the header
	ZoomWeek       // 1 column per day
	ZoomMonth      // 1 column per ISO week
)

func (z Zoom) String() string {
	switch z {
	case ZoomDay:
		return "day"
	case ZoomWeek:
		return "week"
	case ZoomMonth:
		return "month"
	}
	return "auto"
}

// Next cycles day → week → month → day.
func (z Zoom) Next() Zoom {
	if z >= ZoomMonth || z == ZoomAuto {
		return ZoomDay
	}
	return z + 1
}

const day = 24 * time.Hour

// Row is one chart line: a bar (Bar >= 0) or a group header (Bar < 0).
type Row struct {
	Bar        int
	Group      string
	GroupColor color.Color
	Count      int
}

// BarRows lists every bar of l as a row, in layout order.
func BarRows(l timeline.Layout) []Row {
	rows := make([]Row, len(l.Bars))
	for i := range l.Bars {
		rows[i] = Row{Bar: i}
	}
	return rows
}

type Opts struct {
	Width int
	Zoom  Zoom
	// From is the first day shown. Zero starts one unit before the layout.
	From time.Time
	// Selected is the selected row index, or -1.
	Selected int
	// Related marks task IDs linked to the selected task by after.
	Related map[string]bool
	// Conflicts marks task IDs that have an after conflict.
	Conflicts map[string]bool
	Highlight []string
	// Offset and Height select a window of rows. Height <= 0 shows all rows.
	Offset, Height int
}

type Output struct {
	// Lines holds the two header lines and then the visible rows, each
	// exactly Width columns wide.
	Lines      []string
	Zoom       Zoom
	From, To   time.Time
	LabelWidth int
	// First and Last bound the visible rows: rows[First:Last].
	First, Last int
	Clipped     bool
}

// HeaderLines is the number of header lines above the first row.
const HeaderLines = 2

// LabelWidth is the width of the label column for a chart width.
func LabelWidth(width int) int { return min(max(width*3/10, 22), 42) }

func colsPerUnit(z Zoom) int {
	if z == ZoomDay {
		return 3
	}
	return 1
}

func daysPerCol(z Zoom) int {
	if z == ZoomMonth {
		return 7
	}
	return 1
}

// AutoZoom picks the most detailed zoom that fits the layout in width.
func AutoZoom(l timeline.Layout, width int) Zoom {
	cw := width - LabelWidth(width) - 1
	days := int(l.To.Sub(l.From)/day) + 3
	switch {
	case l.Bars == nil, days*3 <= cw:
		return ZoomDay
	case days <= cw:
		return ZoomWeek
	}
	return ZoomMonth
}

// DefaultFrom is the first day shown when the user has not scrolled.
func DefaultFrom(l timeline.Layout, z Zoom, now time.Time) time.Time {
	from := l.From
	if from.IsZero() {
		from = timeline.Day(now)
	}
	from = from.AddDate(0, 0, -daysPerCol(z))
	return Align(from, z)
}

// Align moves d to the start of a column: Monday for week and month zoom.
func Align(d time.Time, z Zoom) time.Time {
	if z == ZoomDay {
		return d
	}
	off := (int(d.Weekday()) + 6) % 7
	return d.AddDate(0, 0, -off)
}

// Step is the number of days that one scroll step moves at zoom z.
func Step(z Zoom) int {
	switch z {
	case ZoomDay:
		return 1
	case ZoomWeek:
		return 7
	}
	return 28
}

type grid struct {
	from    time.Time
	zoom    Zoom
	cols    int
	today   time.Time
	painter map[cellStyle]lipgloss.Style
	c       card.Ctx
}

type cellStyle struct {
	fg, bg color.Color
	bold   bool
}

func (g *grid) style(fg, bg color.Color, bold bool) lipgloss.Style {
	k := cellStyle{fg, bg, bold}
	s, ok := g.painter[k]
	if !ok {
		s = lipgloss.NewStyle().Foreground(fg).Background(bg).Bold(bold)
		g.painter[k] = s
	}
	return s
}

// colDays returns the first and last day that column c covers.
func (g *grid) colDays(c int) (time.Time, time.Time) {
	start := g.from.AddDate(0, 0, c/colsPerUnit(g.zoom)*daysPerCol(g.zoom))
	return start, start.AddDate(0, 0, daysPerCol(g.zoom)-1)
}

func (g *grid) firstColOfUnit(c int) bool { return c%colsPerUnit(g.zoom) == 0 }

func (g *grid) weekend(c int) bool {
	if g.zoom == ZoomMonth {
		return false
	}
	s, _ := g.colDays(c)
	return s.Weekday() == time.Saturday || s.Weekday() == time.Sunday
}

func (g *grid) isToday(c int) bool {
	s, e := g.colDays(c)
	return !g.today.Before(s) && !g.today.After(e) && g.firstColOfUnit(c)
}

// run is a span of cells with one style.
type run struct {
	text  strings.Builder
	style lipgloss.Style
	key   cellStyle
}

type line struct {
	runs []*run
	g    *grid
}

func (l *line) put(s string, fg, bg color.Color, bold bool) {
	k := cellStyle{fg, bg, bold}
	if n := len(l.runs); n > 0 && l.runs[n-1].key == k {
		l.runs[n-1].text.WriteString(s)
		return
	}
	r := &run{style: l.g.style(fg, bg, bold), key: k}
	r.text.WriteString(s)
	l.runs = append(l.runs, r)
}

func (l *line) String() string {
	var b strings.Builder
	for _, r := range l.runs {
		b.WriteString(r.style.Render(r.text.String()))
	}
	return b.String()
}

// Render draws the chart.
func Render(l timeline.Layout, rows []Row, c card.Ctx, o Opts) Output {
	th := c.Look.Theme
	zoom := o.Zoom
	if zoom == ZoomAuto {
		zoom = AutoZoom(l, o.Width)
	}
	from := o.From
	if from.IsZero() {
		from = DefaultFrom(l, zoom, c.Now)
	}
	from = Align(from, zoom)
	lw := LabelWidth(o.Width)
	cw := max(o.Width-lw-1, 1)
	g := &grid{from: from, zoom: zoom, cols: cw, today: timeline.Day(c.Now), painter: map[cellStyle]lipgloss.Style{}, c: c}
	_, to := g.colDays(cw - 1)
	out := Output{Zoom: zoom, From: from, To: to, LabelWidth: lw}

	labelPad := lipgloss.NewStyle().Background(th.Base).Render(strings.Repeat(" ", lw+1))
	out.Lines = append(out.Lines, labelPad+g.monthHeader(), labelPad+g.dayHeader())

	first, last := 0, len(rows)
	if o.Height > 0 && len(rows) > o.Height {
		first = min(max(o.Offset, 0), len(rows)-o.Height)
		last = first + o.Height
	}
	out.First, out.Last = first, last
	for i := first; i < last; i++ {
		r := rows[i]
		if r.Bar < 0 {
			out.Lines = append(out.Lines, g.groupRow(r, lw))
			continue
		}
		b := l.Bars[r.Bar]
		if b.End.After(to) || b.Start.Before(from) {
			out.Clipped = true
		}
		out.Lines = append(out.Lines, g.barRow(b, i == o.Selected, lw, o))
	}
	return out
}

func (g *grid) monthHeader() string {
	th := g.c.Look.Theme
	ln := &line{g: g}
	lastYear, nextFree := 0, 0
	for c := 0; c < g.cols; {
		s, _ := g.colDays(c)
		prev, _ := g.colDays(c - colsPerUnit(g.zoom))
		if (c == 0 || s.Month() != prev.Month()) && c >= nextFree {
			label := s.Format("Jan")
			if s.Year() != lastYear {
				label = s.Format("Jan 2006")
				lastYear = s.Year()
			}
			if c+ansi.StringWidth(label) <= g.cols {
				ln.put(label, th.Subtext, th.Base, true)
				c += ansi.StringWidth(label)
				nextFree = c + 1
				continue
			}
		}
		ln.put(" ", th.Subtext, th.Base, false)
		c++
	}
	return ln.String()
}

func (g *grid) dayHeader() string {
	th := g.c.Look.Theme
	ln := &line{g: g}
	nextFree := 0
	for c := 0; c < g.cols; {
		s, _ := g.colDays(c)
		bg := th.Base
		if g.weekend(c) {
			bg = th.Mantle
		}
		fg := th.Muted
		var label string
		switch g.zoom {
		case ZoomDay:
			label = fmt.Sprintf("%2d ", s.Day())
		case ZoomWeek:
			if s.Weekday() == time.Monday {
				label = fmt.Sprintf("%02d", s.Day())
			}
		case ZoomMonth:
			if c%4 == 0 {
				_, wk := s.ISOWeek()
				label = fmt.Sprintf("W%d", wk)
			}
		}
		if g.isToday(c) {
			fg, bg = th.Base, th.Today
			if label == "" {
				label = "▼"
			}
			if th.Mono {
				fg, bg = th.Text, th.Base
			}
		}
		if label == "" || c < nextFree || c+ansi.StringWidth(label) > g.cols {
			ln.put(" ", fg, bg, false)
			c++
			continue
		}
		ln.put(label, fg, bg, g.isToday(c))
		c += ansi.StringWidth(label)
		if g.zoom != ZoomDay {
			nextFree = c + 1
		}
	}
	return ln.String()
}

func (g *grid) groupRow(r Row, lw int) string {
	th := g.c.Look.Theme
	p := card.Painter{BG: th.Base}
	label := p.Fg(r.GroupColor).Bold(true).Render(" "+strings.ToUpper(r.Group)) +
		p.Fg(th.Muted).Render(fmt.Sprintf("  %d", r.Count))
	ln := &line{g: g}
	ln.put(" "+strings.Repeat("─", g.cols), th.Overlay0, th.Base, false)
	return p.Pad(label, lw) + ln.String()
}

func (g *grid) barRow(b timeline.Bar, selected bool, lw int, o Opts) string {
	th, ic := g.c.Look.Theme, g.c.Look.Icons
	t := b.Task
	rowBg := th.Base
	if selected {
		rowBg = th.Surface0
	}

	// Label: selection edge, status icon, title, and a marker on the right.
	p := card.Painter{BG: rowBg}
	edge := " "
	if selected {
		edge = ic.Selected
	}
	marker, markerFg := "", th.Muted
	switch {
	case o.Conflicts[t.ID]:
		marker, markerFg = ic.Warn, th.Error
	case o.Related[t.ID]:
		marker, markerFg = ic.After, th.Accent
	}
	titleW := lw - 4 - ansi.StringWidth(marker)
	titleStyle := p.Fg(th.Text).Bold(selected)
	if t.Status == model.StatusDone && !selected {
		titleStyle = p.Fg(th.Subtext)
	}
	label := p.Fg(th.Accent).Render(edge) +
		p.Fg(th.Status(t.Status)).Render(ic.Status(t.Status)) + p.Style().Render(" ") +
		card.Mark(ansi.Truncate(t.Title, titleW, "…"), o.Highlight, titleStyle, p.Style().Foreground(th.Base).Background(th.Warn))
	label = p.Pad(label, lw-ansi.StringWidth(marker)-1) + p.Fg(markerFg).Render(marker) + p.Style().Render(" ")

	// Chart cells.
	barColor := th.Status(t.Status)
	ln := &line{g: g}
	ln.put(" ", th.Overlay0, rowBg, false)

	var barCols []int
	for c := 0; c < g.cols; c++ {
		s, e := g.colDays(c)
		if !e.Before(b.Start) && !s.After(b.End) {
			barCols = append(barCols, c)
		}
	}
	inBar := map[int]int{} // column → position in bar
	for i, c := range barCols {
		inBar[c] = i
	}
	dur := fmt.Sprintf("%dd", int(b.End.Sub(b.Start)/day)+1)
	durStart := len(barCols) - len(dur) - 1
	fade := []string{"▓", "▒", "░"}

	for c := 0; c < g.cols; c++ {
		bg := rowBg
		if !selected && g.weekend(c) {
			bg = th.Mantle
		}
		pos, ok := inBar[c]
		switch {
		case ok && b.Open && pos >= len(barCols)-len(fade) && len(barCols) > len(fade):
			ln.put(fade[pos-(len(barCols)-len(fade))], barColor, bg, false)
		case ok && th.Mono:
			ln.put("█", th.Text, bg, false)
		case ok && !b.Open && len(barCols) >= len(dur)+3 && pos >= durStart && pos < durStart+len(dur):
			ln.put(string(dur[pos-durStart]), th.Base, barColor, true)
		case ok && g.isToday(c):
			ln.put("│", th.Base, barColor, true)
		case ok:
			ln.put(" ", th.Base, barColor, false)
		case g.isToday(c):
			ln.put("│", th.Today, bg, false)
		default:
			ln.put(" ", th.Base, bg, false)
		}
	}
	return label + ln.String()
}
