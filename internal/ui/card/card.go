// Package card renders a task as a board card and as a detail panel. The
// TUI and the CLI output share these renderers.
package card

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/ui"
	"github.com/GuitarWag/clitasks/v3/internal/ui/dates"
	"github.com/GuitarWag/clitasks/v3/internal/ui/theme"
)

// Ctx is what every renderer needs: the look and the current time.
type Ctx struct {
	Look ui.Look
	Now  time.Time
}

type Opts struct {
	Width    int
	Selected bool
	// Compact renders one line per card, for short terminals.
	Compact bool
	// Highlight holds lower-case search terms to mark in the title.
	Highlight []string
	// Waiting is true when an after task is not done yet.
	Waiting bool
	// Conflict is true when the timeline has an after conflict for the task.
	Conflict bool
	// ShowID adds the task ID as a last line, for output people copy IDs from.
	ShowID bool
}

// Painter builds styled segments on one background color, so a line has no
// gaps where the terminal background shows through.
type Painter struct {
	BG   color.Color
	Mono bool
}

func (p Painter) Style() lipgloss.Style {
	return lipgloss.NewStyle().Background(p.BG)
}

func (p Painter) Fg(c color.Color) lipgloss.Style { return p.Style().Foreground(c) }

// Pad fills s with background spaces up to width w, cutting it if longer.
func (p Painter) Pad(s string, w int) string {
	if sw := ansi.StringWidth(s); sw > w {
		s = ansi.Truncate(s, w, "…")
	} else if sw < w {
		s += p.Style().Render(strings.Repeat(" ", w-sw))
	}
	return s
}

// Card renders t as a card of exactly o.Width columns.
func Card(t model.Task, c Ctx, o Opts) string {
	th, ic := c.Look.Theme, c.Look.Icons
	bg, stripe := th.Surface0, th.Status(t.Status)
	if o.Selected {
		bg, stripe = th.Surface1, th.Accent
	}
	p := Painter{BG: bg, Mono: th.Mono}
	inner := max(o.Width-3, 4) // stripe, space, content, space

	edge := ic.Selected
	if th.Mono && !o.Selected {
		edge = " "
	}
	left := p.Fg(stripe).Render(edge) + p.Style().Render(" ")
	right := p.Style().Render(" ")
	line := func(s string) string { return left + p.Pad(s, inner) + right }

	title := p.Fg(th.Text).Bold(true)
	if th.Mono && o.Selected {
		title = title.Reverse(true)
	}

	if o.Compact {
		meta := chips(t, c, p, true, o)
		mw := ansi.StringWidth(meta)
		tw := max(inner-mw-1, 4)
		head := highlight(ansi.Truncate(t.Title, tw, "…"), o.Highlight, title, p, th)
		gap := max(inner-ansi.StringWidth(head)-mw, 1)
		return line(head + p.Style().Render(strings.Repeat(" ", gap)) + meta)
	}

	var lines []string
	wrapped := strings.Split(ansi.Wordwrap(t.Title, inner, " -"), "\n")
	for i, l := range wrapped {
		if i == 1 && len(wrapped) > 2 {
			l = ansi.Truncate(strings.Join(wrapped[1:], " "), inner, "…")
		}
		lines = append(lines, line(highlight(ansi.Truncate(l, inner, "…"), o.Highlight, title, p, th)))
		if i == 1 {
			break
		}
	}
	meta := chips(t, c, p, false, o)
	if ansi.StringWidth(meta) > inner && t.Assignee != "" {
		// Drop the assignee before the waiting or conflict chip gets cut.
		noAssignee := t
		noAssignee.Assignee = ""
		meta = chips(noAssignee, c, p, false, o)
	}
	lines = append(lines, line(meta))
	if len(t.Tags) > 0 {
		var tags []string
		for _, tg := range t.Tags {
			tags = append(tags, ic.Tag+tg)
		}
		lines = append(lines, line(p.Fg(th.Muted).Render(strings.Join(tags, " "))))
	}
	if o.ShowID {
		lines = append(lines, line(p.Fg(th.Overlay1).Render(t.ID)))
	}
	return strings.Join(lines, "\n")
}

// chips is the meta line: priority, due, assignee and waiting or conflict.
func chips(t model.Task, c Ctx, p Painter, compact bool, o Opts) string {
	th, ic := c.Look.Theme, c.Look.Icons
	sep := p.Fg(th.Overlay0).Render(" " + ic.Bullet + " ")
	var parts []string

	prio := p.Fg(th.Priority(t.Priority)).Render(ic.Priority(t.Priority))
	if !compact {
		prio += p.Fg(th.Subtext).Render(" " + string(t.Priority))
	}
	parts = append(parts, prio)

	if t.DueDate != "" {
		fg := th.Subtext
		switch dates.DueUrgency(t, c.Now) {
		case dates.UrgencyLate:
			fg = th.Error
		case dates.UrgencyToday:
			fg = th.Today
		case dates.UrgencySoon:
			fg = th.Warn
		}
		parts = append(parts, p.Fg(fg).Render(ic.Due+" "+dates.DueLabel(t, c.Now)))
	}
	if t.Assignee != "" && !compact {
		parts = append(parts, p.Fg(th.Subtext).Render(ic.Assignee+t.Assignee))
	}
	if o.Conflict {
		parts = append(parts, p.Fg(th.Error).Render(ic.Warn))
	} else if o.Waiting {
		parts = append(parts, p.Fg(th.Warn).Render(ic.After))
	}
	return strings.Join(parts, sep)
}

// highlight renders plain text s in base and marks every term with the
// search highlight. Terms are lower case.
func highlight(s string, terms []string, base lipgloss.Style, p Painter, th theme.Theme) string {
	mark := p.Style().Foreground(th.Base).Background(th.Warn).Bold(true)
	if th.Mono {
		mark = p.Style().Reverse(true)
	}
	return Mark(s, terms, base, mark)
}

// Mark renders s in base with each case-insensitive match of terms in mark.
func Mark(s string, terms []string, base, mark lipgloss.Style) string {
	lower := strings.ToLower(s)
	if len(terms) == 0 || len(lower) != len(s) {
		return base.Render(s)
	}
	hit := make([]bool, len(s))
	for _, t := range terms {
		if t == "" {
			continue
		}
		for i := 0; ; {
			j := strings.Index(lower[i:], t)
			if j < 0 {
				break
			}
			for k := i + j; k < i+j+len(t); k++ {
				hit[k] = true
			}
			i += j + len(t)
		}
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && hit[j] == hit[i] {
			j++
		}
		if hit[i] {
			b.WriteString(mark.Render(s[i:j]))
		} else {
			b.WriteString(base.Render(s[i:j]))
		}
		i = j
	}
	return b.String()
}

// Height is the number of lines Card renders for t, without rendering it.
func Height(t model.Task, o Opts) int {
	if o.Compact {
		return 1
	}
	inner := max(o.Width-3, 4)
	h := min(strings.Count(ansi.Wordwrap(t.Title, inner, " -"), "\n")+1, 2) + 1
	if len(t.Tags) > 0 {
		h++
	}
	if o.ShowID {
		h++
	}
	return h
}
