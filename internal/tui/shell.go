package tui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/ui/card"
	"github.com/GuitarWag/clitasks/internal/ui/dates"
	"github.com/GuitarWag/clitasks/internal/ui/theme"
)

// renderHeader is the top bar: brand, board name, tabs and alerts.
func (m Model) renderHeader(w int) string {
	th, ic := m.look.Theme, m.look.Icons
	p := card.Painter{BG: th.Crust}
	info := m.board.Info()

	brand := card.Pill(ic.Done+" tasks", th.Accent, th.Base, th.Mono)
	left := brand + p.Style().Render(" ") +
		p.Fg(th.Text).Bold(true).Render(info.Name) +
		p.Fg(th.Muted).Render(fmt.Sprintf("  %d tasks", len(info.Tasks)))

	tab := func(icon, name string, active bool) string {
		label := " " + strings.TrimSpace(icon+" "+name) + " "
		if active {
			st := lipgloss.NewStyle().Background(th.Surface1).Foreground(th.Text).Bold(true)
			if th.Mono {
				st = lipgloss.NewStyle().Reverse(true).Bold(true)
			}
			return st.Render(label)
		}
		return p.Fg(th.Subtext).Render(label)
	}
	tabs := tab(ic.Board, "Board", m.screen == screenBoard) + p.Style().Render(" ") +
		tab(ic.Timeline, "Timeline", m.screen == screenTimeline)

	var alerts []string
	late := 0
	for _, t := range info.Tasks {
		if dates.DueUrgency(t, m.now()) == dates.UrgencyLate {
			late++
		}
	}
	if late > 0 {
		alerts = append(alerts, p.Fg(th.Error).Render(fmt.Sprintf("%s %d late", ic.Due, late)))
	}
	if n := len(m.conflicts()); n > 0 {
		alerts = append(alerts, p.Fg(th.Warn).Render(fmt.Sprintf("%s %d %s", ic.Warn, n, plural(n, "conflict"))))
	}
	right := tabs
	if len(alerts) > 0 {
		right = strings.Join(alerts, p.Style().Render("  ")) + p.Style().Render("   ") + tabs
	}
	return spread(p, left, right, w)
}

// spread puts left and right on one line of width w, on p's background.
func spread(p card.Painter, left, right string, w int) string {
	lw, rw := ansi.StringWidth(left), ansi.StringWidth(right)
	if lw+rw+1 > w {
		left = ansi.Truncate(left, max(w-rw-1, 0), "…")
		lw = ansi.StringWidth(left)
	}
	gap := max(w-lw-rw, 0)
	return p.Pad(left+p.Style().Render(strings.Repeat(" ", gap))+right, w)
}

// renderStatus is the bottom bar: mode, file, search and key hints.
func (m Model) renderStatus(w int) string {
	th, ic := m.look.Theme, m.look.Icons
	p := card.Painter{BG: th.Crust}

	mode, modeBg := "NORMAL", th.Accent
	switch {
	case m.searching:
		mode, modeBg = "SEARCH", th.Info
	case m.ov != nil:
		mode, modeBg = "EDIT", th.Warn
	}
	left := card.Pill(mode, modeBg, th.Base, th.Mono) + p.Style().Render(" ")
	if m.searching {
		in := m.searchIn
		in.SetWidth(max(w/2, 20))
		left += p.Fg(th.Info).Render(ic.Search+" ") + in.View()
		return spread(p, left, hint(p, th, "enter", "apply")+hint(p, th, "esc", "clear"), w)
	}
	left += p.Fg(th.Subtext).Render(shortPath(m.path))
	if !m.q.empty() {
		left += p.Style().Render("  ") + p.Fg(th.Warn).Bold(true).Render(ic.Search+" "+m.q.raw) +
			p.Fg(th.Muted).Render(fmt.Sprintf(" %d match", m.matchCount()))
	}
	var hints string
	for _, b := range m.hints() {
		h := hint(p, th, b.Help().Key, b.Help().Desc)
		if ansi.StringWidth(left)+ansi.StringWidth(hints)+ansi.StringWidth(h)+2 > w {
			break
		}
		hints += h
	}
	return spread(p, left, hints, w)
}

func hint(p card.Painter, th theme.Theme, k, desc string) string {
	return p.Fg(th.Accent).Bold(true).Render(k) + p.Fg(th.Muted).Render(" "+desc+"  ")
}

func shortPath(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func (m Model) matchCount() int {
	n := 0
	for _, t := range m.board.Info().Tasks {
		if m.visible(t) {
			n++
		}
	}
	return n
}

func (m Model) renderToast() string {
	th, ic := m.look.Theme, m.look.Icons
	bg, icon := th.Info, "i"
	switch m.toast.kind {
	case toastOK:
		bg, icon = th.Success, ic.Done
	case toastWarn:
		bg, icon = th.Warn, ic.Warn
	case toastErr:
		bg, icon = th.Error, ic.Warn
	}
	text := ansi.Truncate(m.toast.text, max(m.width-10, 20), "…")
	return card.Pill(icon+" "+text, bg, th.Base, th.Mono)
}

// overlayBox frames the open overlay: a rounded border, a title pill and the
// body on the Mantle background.
func (m Model) overlayBox(w, h int) string {
	th := m.look.Theme
	title, body := m.ov.view(&m)
	bw := 0
	for _, l := range strings.Split(body, "\n") {
		bw = max(bw, ansi.StringWidth(l))
	}
	bw = min(max(bw, ansi.StringWidth(title)+4), w-6)
	p := card.Painter{BG: th.Mantle}
	lines := []string{card.Pill(title, th.Accent, th.Base, th.Mono), ""}
	lines = append(lines, strings.Split(body, "\n")...)
	if limit := h - 4; len(lines) > limit {
		lines = lines[:limit]
	}
	for i, l := range lines {
		lines[i] = p.Pad(l, bw)
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).BorderForeground(th.Accent).BorderBackground(th.Base).
		Background(th.Mantle).Padding(1, 2)
	out := box.Render(strings.Join(lines, "\n"))
	if th.Mono {
		return out
	}
	return fillBG(out, th.Mantle)
}

// fillBG gives every cell of s that has no background the color bg.
// Components such as huh leave spacing cells unstyled, and those would
// show the terminal background through the panel.
func fillBG(s string, bg color.Color) string {
	w, h := lipgloss.Width(s), lipgloss.Height(s)
	c := lipgloss.NewCanvas(w, h)
	c.Compose(lipgloss.NewLayer(s))
	for y := range h {
		for x := range w {
			if cell := c.CellAt(x, y); cell != nil && cell.Style.Bg == nil {
				cell.Style.Bg = bg
			}
		}
	}
	return c.Render()
}

// statusName is the column title of a status.
func statusName(s model.TaskStatus) string {
	switch s {
	case model.StatusInProgress:
		return "IN PROGRESS"
	case model.StatusBlocked:
		return "BLOCKED"
	case model.StatusDone:
		return "DONE"
	}
	return "TODO"
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
