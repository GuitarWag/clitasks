package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/GuitarWag/clitasks/internal/theme"
	"github.com/GuitarWag/clitasks/internal/timeline"
)

const (
	defaultTimelineWidth = 100
	// timelineChrome is the number of lines around the chart rows: board
	// header, chart axis, summary and footer.
	timelineChrome = 9
)

func (m Model) timelineLayout() timeline.Layout {
	return timeline.Build(applyFilter(m.board.Info().Tasks, m.filter), m.now())
}

func (m Model) timelineASCII(l timeline.Layout) timeline.ASCII {
	w := m.width
	if w <= 0 {
		w = defaultTimelineWidth
	}
	return timeline.RenderASCII(l, timeline.ASCIIOptions{
		Width: w, Today: m.now(), From: m.tlFrom, Scale: m.tlScale,
	})
}

func (m Model) updateTimeline(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, m.keys.Timeline), key.Matches(km, m.keys.Esc):
		m.showTimeline = false
		m.clampSelection()
	case key.Matches(km, m.keys.Up):
		if m.tlRow > 0 {
			m.tlRow--
		}
	case key.Matches(km, m.keys.Down):
		if m.tlRow < len(m.timelineLayout().Bars)-1 {
			m.tlRow++
		}
	case key.Matches(km, m.keys.Left), key.Matches(km, m.keys.Right):
		out := m.timelineASCII(m.timelineLayout())
		step := -7 * 24 * time.Hour
		if key.Matches(km, m.keys.Right) {
			step = -step
		}
		// Keep the scale from the first render, so the columns do not change
		// size while the user scrolls.
		m.tlScale = out.Scale
		m.tlFrom = out.From.Add(step)
	default:
		// a, e, d, s, f, r, ?, q work the same as on the board. They act on
		// selectedTask, which reads the timeline row.
		return m.updateBoard(msg)
	}
	return m, nil
}

func (m Model) viewTimeline() string {
	var b strings.Builder
	m.writeHeader(&b)

	l := m.timelineLayout()
	if len(l.Bars) == 0 {
		fmt.Fprintln(&b, m.styles.taskDim.Render("No scheduled tasks"))
	} else {
		out := m.timelineASCII(l)
		fmt.Fprintln(&b, m.styles.headerDim.Render(out.Header))
		first, last := visibleRows(len(out.Rows), m.tlRow, m.height-timelineChrome)
		for i := first; i < last; i++ {
			st := theme.StatusStyle(l.Bars[i].Task.Status)
			if i == m.tlRow {
				st = m.styles.taskSel
			}
			fmt.Fprintln(&b, st.Render(out.Rows[i]))
		}
		scale := "day"
		if out.Scale == timeline.ScaleWeek {
			scale = "week"
		}
		fmt.Fprintln(&b, m.styles.headerDim.Render(fmt.Sprintf("1 column = 1 %s · %s to %s · rows %d-%d of %d",
			scale, out.From.Format(time.DateOnly), out.To.Format(time.DateOnly), first+1, last, len(out.Rows))))
	}
	if n := len(l.Warnings); n > 0 {
		fmt.Fprintln(&b, m.styles.error.Render(fmt.Sprintf("%d after conflict(s), run `tasks timeline` to list them", n)))
	}
	if n := len(l.Unscheduled); n > 0 {
		fmt.Fprintln(&b, m.styles.headerDim.Render(fmt.Sprintf("%d unscheduled task(s) not shown", n)))
	}
	b.WriteString(m.styles.footer.Render("↑/k ↓/j select · ←/h →/l scroll a week · e edit · s status · d delete · a add · f filter · t/esc board · ? help · q quit"))
	return b.String()
}

// visibleRows returns the [first, last) window of n rows that has room for
// space rows and keeps sel inside it. space <= 0 means no height limit.
func visibleRows(n, sel, space int) (int, int) {
	if space <= 0 || n <= space {
		return 0, n
	}
	first := max(0, sel-space+1)
	return first, first + space
}
