package timeline

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

const (
	cellBar   = '█'
	cellOpen  = '>'
	cellToday = '|'
	cellEmpty = ' '

	maxLabelWidth = 32
	minChartWidth = 10
	day           = 24 * time.Hour
)

// Scale is the number of days in one chart column.
type Scale int

const (
	ScaleDay  Scale = 1
	ScaleWeek Scale = 7
)

// ASCIIOptions controls RenderASCII. From moves the left edge of the chart,
// for scrolling; RenderASCII uses its calendar date in its own location. Zero From starts at the layout's first day. Zero Scale picks
// the scale from the range; a scrolling caller passes the first result's
// Scale so the scale does not change as the edge moves.
type ASCIIOptions struct {
	Width int
	Today time.Time
	From  time.Time
	Scale Scale
}

// ASCII is a rendered chart. Rows[i] is the line for Layout.Bars[i]. The
// output has no color, so callers can style each row.
type ASCII struct {
	Header string
	Rows   []string
	Scale  Scale
	// From and To are the first and last day that the chart columns show.
	From, To time.Time
	// Clipped is true when at least one bar does not fit in the columns.
	Clipped bool
}

// RenderASCII draws one line for each bar: a label (ID and title) and a chart.
// It uses one column per day if the whole range fits, else one column per
// ISO week (Monday start).
func RenderASCII(l Layout, o ASCIIOptions) ASCII {
	labelW := min(maxLabelWidth, max(o.Width/3, 12))
	chartW := max(o.Width-labelW-1, minChartWidth)
	today := Day(o.Today)

	from, to := l.From, l.To
	if from.IsZero() {
		from, to = today, today
	}
	if !o.From.IsZero() {
		// From is already a date. Day() would shift it a day west of UTC.
		y, m, d := o.From.Date()
		from = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	}
	scale := o.Scale
	if scale == 0 {
		scale = ScaleDay
		if days(from, to)+1 > chartW {
			scale = ScaleWeek
		}
	}
	if scale == ScaleWeek {
		from = monday(from)
	}
	out := ASCII{Scale: scale, From: from}
	out.To = from.Add(time.Duration(chartW*int(scale))*day - day)

	out.Header = runewidth.FillRight("", labelW) + " " + axis(from, chartW, scale)
	for _, b := range l.Bars {
		label := runewidth.FillRight(runewidth.Truncate(b.Task.ID+" "+b.Task.Title, labelW, "…"), labelW)
		var sb strings.Builder
		for c := range chartW {
			cs := from.Add(time.Duration(c*int(scale)) * day)
			ce := cs.Add(time.Duration(scale)*day - day)
			sb.WriteRune(cell(b, cs, ce, today))
		}
		if b.End.After(out.To) || b.Start.Before(from) {
			out.Clipped = true
		}
		out.Rows = append(out.Rows, label+" "+strings.TrimRight(sb.String(), " "))
	}
	return out
}

func cell(b Bar, cs, ce, today time.Time) rune {
	if !b.Start.After(ce) && !b.End.Before(cs) {
		if b.Open && !b.End.Before(cs) && !b.End.After(ce) {
			return cellOpen
		}
		return cellBar
	}
	if !today.Before(cs) && !today.After(ce) {
		return cellToday
	}
	return cellEmpty
}

// axis labels the chart columns with MM-DD dates: each Monday on the day
// scale and every fourth week on the week scale, when the label fits.
func axis(from time.Time, cols int, scale Scale) string {
	line := []rune(strings.Repeat(" ", cols))
	next := 0
	for c := range cols {
		d := from.Add(time.Duration(c*int(scale)) * day)
		mark := (scale == ScaleDay && d.Weekday() == time.Monday) || (scale == ScaleWeek && c%4 == 0)
		label := d.Format("01-02")
		if !mark || c < next || c+len(label) > cols {
			continue
		}
		copy(line[c:], []rune(label))
		next = c + len(label) + 1
	}
	return strings.TrimRight(string(line), " ")
}

func days(a, b time.Time) int { return int(b.Sub(a) / day) }

func monday(t time.Time) time.Time {
	off := (int(t.Weekday()) + 6) % 7
	return t.Add(-time.Duration(off) * day)
}
