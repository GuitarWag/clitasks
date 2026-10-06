package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/ui/card"
	"github.com/GuitarWag/clitasks/v3/internal/ui/dates"
)

const (
	bodyTop      = 2 // header line and one blank line
	minColWidth  = 24
	colGap       = 2
	rightPaneW   = 44
	bottomPaneH  = 11
	compactBelow = 22 // body heights below this use one-line cards
)

type paneMode int

const (
	paneNone   paneMode = iota // too narrow: enter opens an overlay
	paneHidden                 // the user closed the pane
	paneRight
	paneBottom
)

// paneMode is where the detail pane goes at the current size.
func (m Model) paneMode() paneMode {
	w, h := m.size()
	var mode paneMode
	switch {
	case w >= 140:
		mode = paneRight
	case w >= 100 && h-3 >= 26:
		mode = paneBottom
	case w >= 100:
		mode = paneRight
	default:
		return paneNone
	}
	if !m.detail {
		return paneHidden
	}
	return mode
}

// placed is a card on screen, in body coordinates.
type placed struct {
	id     string
	y0, y1 int // first and last line
}

type colGeom struct {
	index        int // into columnOrder
	x, w         int
	cards        []placed
	above, below int // cards hidden by scrolling
}

type boardGeom struct {
	cols           []colGeom
	colH           int
	compact        bool
	paneX, paneY   int
	paneW, paneH   int
	mode           paneMode
	boardW, boardH int
}

func (m Model) columnTasks(i int) []model.Task {
	var out []model.Task
	for _, t := range m.board.ByStatus()[columnOrder[i]] {
		if m.visible(t) {
			out = append(out, t)
		}
	}
	return out
}

func (m Model) cardOpts(t model.Task, width int, selected, compact bool, conflicts map[string]bool) card.Opts {
	return card.Opts{
		Width: width, Selected: selected, Compact: compact,
		Highlight: m.q.terms(), Waiting: m.waiting(t), Conflict: conflicts[t.ID],
	}
}

// geometry lays out the board for a body of w × h.
func (m Model) geometry(w, h int) boardGeom {
	g := boardGeom{mode: m.paneMode(), boardW: w - 2, boardH: h}
	switch g.mode {
	case paneRight:
		g.paneW, g.paneH = min(rightPaneW, w/4+8), h
		g.boardW = w - g.paneW - 3
		g.paneX = w - g.paneW
	case paneBottom:
		g.paneW, g.paneH = w, bottomPaneH
		g.boardH = h - bottomPaneH - 1
		g.paneY = g.boardH + 1
	}
	g.compact = g.boardH < compactBelow

	n := min(max((g.boardW+colGap)/(minColWidth+colGap), 1), len(columnOrder))
	start := max(0, min(m.col-(n-1), len(columnOrder)-n))
	cw := (g.boardW - colGap*(n-1)) / n
	g.colH = g.boardH
	for i := 0; i < n; i++ {
		ci := start + i
		cg := colGeom{index: ci, x: 1 + i*(cw+colGap), w: cw}
		tasks := m.columnTasks(ci)
		cg.above = min(m.scroll[ci], max(len(tasks)-1, 0))
		y := 3 // header, rule, and a gap line that shows the ↑ marker
		for j := cg.above; j < len(tasks); j++ {
			t := tasks[j]
			hgt := card.Height(t, card.Opts{Width: cw, Compact: g.compact})
			reserve := 0
			if j < len(tasks)-1 {
				reserve = 1 // room for the ↓ marker
			}
			if y+hgt+reserve > g.colH && len(cg.cards) > 0 {
				cg.below = len(tasks) - j
				break
			}
			cg.cards = append(cg.cards, placed{id: t.ID, y0: y, y1: y + hgt - 1})
			y += hgt
			if !g.compact {
				y++
			}
		}
		g.cols = append(g.cols, cg)
	}
	return g
}

func (m Model) renderBoard(w, h int, conflicts map[string]bool) string {
	th := m.look.Theme
	if len(m.board.Info().Tasks) == 0 {
		return m.emptyBoard(w, h)
	}
	g := m.geometry(w, h)
	canvas := make([]string, h)
	base := card.Painter{BG: th.Base}
	for i := range canvas {
		canvas[i] = base.Pad("", w)
	}
	var layers []*lipgloss.Layer
	layers = append(layers, lipgloss.NewLayer(strings.Join(canvas, "\n")))

	for _, cg := range g.cols {
		layers = append(layers, lipgloss.NewLayer(m.renderColumn(cg, g, conflicts)).X(cg.x))
	}
	if g.mode == paneRight || g.mode == paneBottom {
		layers = append(layers, lipgloss.NewLayer(m.renderPane(g.paneW, g.paneH)).X(g.paneX).Y(g.paneY))
	}
	return lipgloss.NewCanvas(w, h).Compose(lipgloss.NewCompositor(layers...)).Render()
}

func (m Model) renderColumn(cg colGeom, g boardGeom, conflicts map[string]bool) string {
	th, ic := m.look.Theme, m.look.Icons
	st := columnOrder[cg.index]
	p := card.Painter{BG: th.Base}
	tasks := m.columnTasks(cg.index)
	focused := cg.index == m.col

	late := 0
	for _, t := range tasks {
		if dates.DueUrgency(t, m.now()) == dates.UrgencyLate || conflicts[t.ID] {
			late++
		}
	}
	head := p.Fg(th.Status(st)).Bold(true).Render(ic.Status(st)+" "+statusName(st)) +
		p.Fg(th.Muted).Render(fmt.Sprintf("  %d", len(tasks)))
	if late > 0 {
		head += p.Fg(th.Error).Render(fmt.Sprintf(" · %d", late))
	}
	rule := p.Fg(th.Overlay0).Render(strings.Repeat("─", cg.w))
	if focused {
		rule = p.Fg(th.Accent).Render(strings.Repeat("━", cg.w))
		if th.Mono {
			rule = p.Style().Bold(true).Render(strings.Repeat("━", cg.w))
		}
	}
	lines := make([]string, g.colH)
	for i := range lines {
		lines[i] = p.Pad("", cg.w)
	}
	lines[0] = p.Pad(head, cg.w)
	lines[1] = rule
	if cg.above > 0 && g.colH > 2 {
		lines[2] = p.Pad(p.Fg(th.Muted).Render(fmt.Sprintf("  ↑ %d more", cg.above)), cg.w)
	}

	if len(tasks) == 0 && g.colH > 4 {
		msg := "Empty"
		if st == model.StatusTodo {
			msg = "Nothing to do · press a"
		}
		lines[3] = p.Pad(p.Fg(th.Muted).Render("  "+msg), cg.w)
	}
	sel, _ := m.selected()
	for _, pc := range cg.cards {
		t, _ := m.board.Get(pc.id)
		c := card.Card(t, m.ctx(), m.cardOpts(t, cg.w, focused && t.ID == sel.ID, g.compact, conflicts))
		for i, l := range strings.Split(c, "\n") {
			if y := pc.y0 + i; y < len(lines) {
				lines[y] = l
			}
		}
	}
	if cg.below > 0 {
		lines[len(lines)-1] = p.Pad(p.Fg(th.Muted).Render(fmt.Sprintf("  ↓ %d more", cg.below)), cg.w)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderPane(w, h int) string {
	th := m.look.Theme
	p := card.Painter{BG: th.Mantle}
	t, ok := m.selected()
	var lines []string
	if !ok {
		lines = []string{p.Pad("", w), p.Pad(p.Fg(th.Muted).Render("  No task selected"), w)}
	} else {
		lines = card.Detail(t, m.related(t), m.ctx(), w)
	}
	return fit(strings.Join(lines, "\n"), w, h, th.Mantle)
}

func (m Model) emptyBoard(w, h int) string {
	th, ic := m.look.Theme, m.look.Icons
	p := card.Painter{BG: th.Base}
	msg := []string{
		p.Fg(th.Accent).Bold(true).Render(ic.Todo + "  No tasks yet"),
		"",
		p.Fg(th.Subtext).Render("Press ") + p.Fg(th.Accent).Bold(true).Render("a") + p.Fg(th.Subtext).Render(" to add your first task,"),
		p.Fg(th.Subtext).Render("or ") + p.Fg(th.Accent).Bold(true).Render("ctrl+k") + p.Fg(th.Subtext).Render(" for every command."),
	}
	var out []string
	for range max((h-len(msg))/2, 0) {
		out = append(out, "")
	}
	for _, l := range msg {
		pad := max((w-ansi.StringWidth(l))/2, 0)
		out = append(out, p.Style().Render(strings.Repeat(" ", pad))+l)
	}
	return fit(strings.Join(out, "\n"), w, h, th.Base)
}

// --- input ---

func (m Model) boardKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	tasks := m.columnTasks(m.col)
	i := max(indexOfID(tasks, m.sel[m.col]), 0)
	switch {
	case key.Matches(k, bk.Up):
		if i > 0 {
			m.sel[m.col] = tasks[i-1].ID
		}
	case key.Matches(k, bk.Down):
		if i < len(tasks)-1 {
			m.sel[m.col] = tasks[i+1].ID
		}
	case key.Matches(k, bk.Top):
		if len(tasks) > 0 {
			m.sel[m.col] = tasks[0].ID
		}
	case key.Matches(k, bk.Bottom):
		if len(tasks) > 0 {
			m.sel[m.col] = tasks[len(tasks)-1].ID
		}
	case key.Matches(k, bk.Left):
		m.col = max(m.col-1, 0)
	case key.Matches(k, bk.Right):
		m.col = min(m.col+1, len(columnOrder)-1)
	case key.Matches(k, bk.Col1), key.Matches(k, bk.Col2), key.Matches(k, bk.Col3), key.Matches(k, bk.Col4):
		m.col = int(k.String()[0] - '1')
	case key.Matches(k, bk.MoveLeft), key.Matches(k, bk.MoveRight):
		t, ok := m.selected()
		if !ok {
			return m, nil
		}
		to := m.col - 1
		if key.Matches(k, bk.MoveRight) {
			to = m.col + 1
		}
		if to < 0 || to >= len(columnOrder) {
			return m, nil
		}
		st := columnOrder[to]
		cmd := m.write("Moved to "+card.StatusLabel(st), func() error {
			_, err := m.board.Move(t.ID, st)
			return err
		})
		if moved, ok := m.board.Get(t.ID); ok && moved.Status == st {
			m.col = to
			m.sel[to] = t.ID
		}
		m.fixScroll()
		return m, cmd
	}
	m.fixScroll()
	return m, nil
}

// fixScroll keeps the selection valid and on screen on both screens.
func (m *Model) fixScroll() {
	for i := range columnOrder {
		tasks := m.columnTasks(i)
		if indexOfID(tasks, m.sel[i]) < 0 {
			m.sel[i] = ""
			if len(tasks) > 0 {
				m.sel[i] = tasks[0].ID
			}
		}
		m.scroll[i] = min(m.scroll[i], max(len(tasks)-1, 0))
	}
	w, h := m.size()
	tasks := m.columnTasks(m.col)
	if si := indexOfID(tasks, m.sel[m.col]); si >= 0 {
		if si < m.scroll[m.col] {
			m.scroll[m.col] = si
		}
		for range len(tasks) {
			g := m.geometry(w, h-3)
			var cg *colGeom
			for i := range g.cols {
				if g.cols[i].index == m.col {
					cg = &g.cols[i]
				}
			}
			if cg == nil || cardIn(cg.cards, m.sel[m.col]) || m.scroll[m.col] >= si {
				break
			}
			m.scroll[m.col]++
		}
	}
	m.fixTimelineScroll()
}

func cardIn(cards []placed, id string) bool {
	for _, c := range cards {
		if c.id == id {
			return true
		}
	}
	return false
}

func (m Model) mouseClick(msg tea.MouseClickMsg) (tea.Model, tea.Cmd) {
	mo := msg.Mouse()
	if mo.Button != tea.MouseLeft {
		return m, nil
	}
	w, h := m.size()
	if mo.Y == 0 {
		return m.clickHeader(mo.X, w)
	}
	if m.screen == screenTimeline {
		return m.timelineClick(mo.X, mo.Y-bodyTop)
	}
	g := m.geometry(w, h-3)
	x, y := mo.X, mo.Y-bodyTop
	for _, cg := range g.cols {
		if x < cg.x || x >= cg.x+cg.w {
			continue
		}
		m.col = cg.index
		for _, pc := range cg.cards {
			if y >= pc.y0 && y <= pc.y1 {
				m.sel[cg.index] = pc.id
				return m.maybeDoubleClick(pc.id)
			}
		}
	}
	m.fixScroll()
	return m, nil
}

// maybeDoubleClick opens the editor on a second click on the same task.
func (m Model) maybeDoubleClick(id string) (tea.Model, tea.Cmd) {
	now := time.Now()
	if m.lastClick.id == id && now.Sub(m.lastClick.at) < doubleClickIn {
		m.lastClick.id = ""
		if t, ok := m.board.Get(id); ok {
			f, cmd := newForm(&m, &t)
			m.ov = f
			return m, cmd
		}
	}
	m.lastClick.id, m.lastClick.at = id, now
	m.fixScroll()
	return m, nil
}

// clickHeader switches screens when a tab is clicked.
func (m Model) clickHeader(x, w int) (tea.Model, tea.Cmd) {
	if x > w-12 {
		m.screen = screenTimeline
	} else if x > w-22 {
		m.screen = screenBoard
	}
	m.fixScroll()
	return m, nil
}

func (m Model) mouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	mo := msg.Mouse()
	if m.screen == screenTimeline {
		return m.timelineWheel(mo)
	}
	w, h := m.size()
	for _, cg := range m.geometry(w, h-3).cols {
		if mo.X >= cg.x && mo.X < cg.x+cg.w {
			m.col = cg.index
		}
	}
	k := bk.Down
	if mo.Button == tea.MouseWheelUp {
		k = bk.Up
	}
	return m.boardKey(tea.KeyPressMsg{Code: []rune(k.Keys()[1])[0], Text: k.Keys()[1]})
}
