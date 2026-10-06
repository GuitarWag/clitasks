package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/v3/internal/export"
	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/ui/card"
	"github.com/GuitarWag/clitasks/v3/internal/ui/icons"
	"github.com/GuitarWag/clitasks/v3/internal/ui/theme"
	"github.com/GuitarWag/clitasks/v3/internal/ui/tlgrid"
)

func esc(msg tea.Msg) bool {
	k, ok := msg.(tea.KeyPressMsg)
	return ok && (k.String() == "esc" || k.String() == "ctrl+c")
}

func pressed(msg tea.Msg) string {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		return k.String()
	}
	return ""
}

// footer is a line of key hints at the bottom of an overlay.
func (m *Model) footer(pairs ...string) string {
	p := card.Painter{BG: m.look.Theme.Mantle}
	var s string
	for i := 0; i+1 < len(pairs); i += 2 {
		s += hint(p, m.look.Theme, pairs[i], pairs[i+1])
	}
	return s
}

// --- status picker ---

type statusPicker struct {
	task model.Task
	sel  int
}

func newStatusPicker(t model.Task) *statusPicker {
	sp := &statusPicker{task: t}
	for i, s := range columnOrder {
		if s == t.Status {
			sp.sel = i
		}
	}
	return sp
}

func (s *statusPicker) update(m *Model, msg tea.Msg) (bool, tea.Cmd) {
	if esc(msg) {
		return true, nil
	}
	switch k := pressed(msg); k {
	case "up", "k":
		s.sel = max(s.sel-1, 0)
	case "down", "j", "space":
		s.sel = min(s.sel+1, len(columnOrder)-1)
	case "1", "2", "3", "4":
		s.sel = int(k[0] - '1')
		return true, s.apply(m)
	case "enter":
		return true, s.apply(m)
	}
	return false, nil
}

func (s *statusPicker) apply(m *Model) tea.Cmd {
	st := columnOrder[s.sel]
	if st == s.task.Status {
		return nil
	}
	cmd := m.write("Moved to "+card.StatusLabel(st), func() error {
		_, err := m.board.Move(s.task.ID, st)
		return err
	})
	if t, ok := m.board.Get(s.task.ID); ok && t.Status == st {
		m.selectTask(t)
	}
	return cmd
}

func (s *statusPicker) view(m *Model) (string, string) {
	th, ic := m.look.Theme, m.look.Icons
	var lines []string
	p := card.Painter{BG: th.Mantle}
	lines = append(lines, p.Fg(th.Subtext).Render(ansi.Truncate(s.task.Title, 40, "…")), "")
	for i, st := range columnOrder {
		row := p
		if i == s.sel {
			row = card.Painter{BG: th.Surface1}
		}
		cur := "  "
		if st == s.task.Status {
			cur = " " + ic.Bullet
		}
		line := row.Fg(th.Muted).Render(fmt.Sprintf(" %d ", i+1)) +
			row.Fg(th.Status(st)).Bold(true).Render(ic.Status(st)+" "+card.StatusLabel(st)) +
			row.Fg(th.Muted).Render(cur)
		lines = append(lines, row.Pad(line, 32))
	}
	lines = append(lines, "", m.footer("1-4", "pick", "enter", "move", "esc", "cancel"))
	return "Move to", strings.Join(lines, "\n")
}

// --- confirm delete ---

type confirmOverlay struct {
	task       model.Task
	dependents []string
}

func (c *confirmOverlay) update(m *Model, msg tea.Msg) (bool, tea.Cmd) {
	if esc(msg) {
		return true, nil
	}
	switch pressed(msg) {
	case "y", "Y", "enter":
		text := "Deleted " + c.task.ID
		if n := len(c.dependents); n > 0 {
			text += fmt.Sprintf(" · unlinked %d task(s)", n)
		}
		return true, m.write(text, func() error {
			_, err := m.board.Delete(c.task.ID)
			return err
		})
	case "n", "N":
		return true, nil
	}
	return false, nil
}

func (c *confirmOverlay) view(m *Model) (string, string) {
	th, ic := m.look.Theme, m.look.Icons
	p := card.Painter{BG: th.Mantle}
	lines := []string{
		p.Fg(th.Text).Bold(true).Render(ansi.Truncate(c.task.Title, 52, "…")),
		p.Fg(th.Muted).Render(c.task.ID),
	}
	if len(c.dependents) > 0 {
		lines = append(lines, "", p.Fg(th.Warn).Render(ic.Warn+" These tasks wait for it and lose the link:"))
		for _, id := range c.dependents {
			if t, ok := m.board.Get(id); ok {
				lines = append(lines, p.Fg(th.Subtext).Render("  "+ic.After+" "+ansi.Truncate(t.Title, 46, "…")))
			}
		}
	}
	lines = append(lines, "", m.footer("y", "delete", "n/esc", "keep"))
	return "Delete task?", strings.Join(lines, "\n")
}

// --- detail (narrow terminals) ---

type detailOverlay struct {
	id     string
	offset int
}

func (d *detailOverlay) update(m *Model, msg tea.Msg) (bool, tea.Cmd) {
	if esc(msg) {
		return true, nil
	}
	switch pressed(msg) {
	case "enter", "q":
		return true, nil
	case "down", "j":
		d.offset++
	case "up", "k":
		d.offset = max(d.offset-1, 0)
	case "e":
		if t, ok := m.board.Get(d.id); ok {
			f, cmd := newForm(m, &t)
			m.ov = f
			return false, cmd
		}
	}
	return false, nil
}

func (d *detailOverlay) view(m *Model) (string, string) {
	t, ok := m.board.Get(d.id)
	if !ok {
		return "Task", "Task not found"
	}
	w, h := m.size()
	lines := card.Detail(t, m.related(t), m.ctx(), min(w-12, 64))
	room := max(h-10, 4)
	d.offset = min(d.offset, max(len(lines)-room, 0))
	lines = lines[d.offset:min(d.offset+room, len(lines))]
	return "Task", strings.Join(append(lines, "", m.footer("↑↓", "scroll", "e", "edit", "esc", "close")), "\n")
}

// --- help ---

type helpOverlay struct{}

func (h *helpOverlay) update(_ *Model, msg tea.Msg) (bool, tea.Cmd) {
	switch pressed(msg) {
	case "esc", "?", "q", "enter":
		return true, nil
	}
	return false, nil
}

func (h *helpOverlay) view(m *Model) (string, string) {
	th := m.look.Theme
	p := card.Painter{BG: th.Mantle}
	section := func(title string, bs []key.Binding) []string {
		out := []string{p.Fg(th.Subtext).Bold(true).Render(title)}
		for _, b := range bs {
			out = append(out, p.Pad(p.Fg(th.Accent).Bold(true).Render(fmt.Sprintf("%-8s", b.Help().Key))+
				p.Fg(th.Text).Render(b.Help().Desc), 30))
		}
		return append(out, "")
	}
	left := append(section("Everywhere", globalBindings()), section("Board", boardBindings())...)
	right := section("Timeline", timelineBindings())
	var lines []string
	for i := range max(len(left), len(right)) {
		l, r := p.Pad("", 30), ""
		if i < len(left) {
			l = p.Pad(left[i], 30)
		}
		if i < len(right) {
			r = right[i]
		}
		lines = append(lines, l+p.Style().Render("   ")+r)
	}
	lines = append(lines, p.Fg(th.Muted).Render("Search: words  @who  #tag  !high  is:overdue|blocked|waiting  due:<7d"))
	return "Keys", strings.Join(lines, "\n")
}

// --- command palette ---

type action struct {
	name string
	keys string
	run  func(m *Model) tea.Cmd
}

type palette struct {
	in      textinput.Model
	sel     int
	actions []action
}

func newPalette(m *Model) *palette {
	in := textinput.New()
	in.Prompt = ""
	in.Placeholder = "type a command"
	in.Focus()
	return &palette{in: in, actions: paletteActions(m)}
}

func paletteActions(m *Model) []action {
	press := func(k string) func(m *Model) tea.Cmd {
		return func(m *Model) tea.Cmd {
			r := []rune(k)
			mm, cmd := m.Update(tea.KeyPressMsg{Code: r[0], Text: k})
			*m = mm.(Model)
			return cmd
		}
	}
	setZoom := func(z tlgrid.Zoom) func(m *Model) tea.Cmd {
		return func(m *Model) tea.Cmd {
			m.screen, m.tl.zoom, m.tl.from = screenTimeline, z, time.Time{}
			m.fixScroll()
			return nil
		}
	}
	setGroup := func(g groupBy) func(m *Model) tea.Cmd {
		return func(m *Model) tea.Cmd {
			m.screen, m.tl.group, m.tl.offset = screenTimeline, g, 0
			m.fixScroll()
			return nil
		}
	}
	acts := []action{
		{"Add task", "a", press("a")},
		{"Edit selected task", "e", press("e")},
		{"Delete selected task", "d", press("d")},
		{"Change status", "space", func(m *Model) tea.Cmd {
			if t, ok := m.selected(); ok {
				m.ov = newStatusPicker(t)
			}
			return nil
		}},
		{"Search", "/", press("/")},
		{"Toggle details", "enter", func(m *Model) tea.Cmd { m.detail = !m.detail; m.fixScroll(); return nil }},
		{"Show board", "tab", func(m *Model) tea.Cmd { m.screen = screenBoard; m.fixScroll(); return nil }},
		{"Show timeline", "tab", func(m *Model) tea.Cmd { m.screen = screenTimeline; m.fixScroll(); return nil }},
		{"Timeline: day zoom", "z", setZoom(tlgrid.ZoomDay)},
		{"Timeline: week zoom", "z", setZoom(tlgrid.ZoomWeek)},
		{"Timeline: month zoom", "z", setZoom(tlgrid.ZoomMonth)},
		{"Timeline: fit everything", "f", setZoom(tlgrid.ZoomAuto)},
		{"Timeline: no grouping", "b", setGroup(groupNone)},
		{"Timeline: group by status", "b", setGroup(groupStatus)},
		{"Timeline: group by assignee", "b", setGroup(groupAssignee)},
		{"Timeline: group by tag", "b", setGroup(groupTag)},
		{"Copy Mermaid gantt to clipboard", "", func(m *Model) tea.Cmd {
			return tea.Batch(tea.SetClipboard(string(export.ToGantt(m.board.Info(), m.now()))),
				m.showToast("Copied Mermaid gantt to the clipboard", toastOK))
		}},
		{"Reload tasks.md", "r", func(m *Model) tea.Cmd { return m.reload("Reloaded") }},
		{"Keyboard help", "?", func(m *Model) tea.Cmd { m.ov = &helpOverlay{}; return nil }},
	}
	for _, n := range theme.Names() {
		acts = append(acts, action{"Theme: " + n, "", func(m *Model) tea.Cmd {
			m.opts.Theme = n
			m.look = m.opts.Resolve(m.dark)
			return m.showToast("Theme "+n+" · set TASKS_THEME to keep it", toastInfo)
		}})
	}
	for _, n := range []string{icons.NameAuto, icons.NameNerd, icons.NameUnicode, icons.NameASCII} {
		acts = append(acts, action{"Icons: " + n, "", func(m *Model) tea.Cmd {
			m.opts.Icons = n
			m.look = m.opts.Resolve(m.dark)
			return m.showToast("Icons "+n+" · set TASKS_ICONS to keep them", toastInfo)
		}})
	}
	return append(acts, action{"Quit", "q", func(*Model) tea.Cmd { return tea.Quit }})
}

// matches is a fuzzy subsequence match on the action name.
func fuzzy(name, q string) bool {
	name, q = strings.ToLower(name), strings.ToLower(strings.TrimSpace(q))
	i := 0
	for _, r := range name {
		if i < len(q) && rune(q[i]) == r {
			i++
		}
	}
	return i == len(q)
}

func (p *palette) filtered() []action {
	var out []action
	for _, a := range p.actions {
		if fuzzy(a.name, p.in.Value()) {
			out = append(out, a)
		}
	}
	return out
}

func (p *palette) update(m *Model, msg tea.Msg) (bool, tea.Cmd) {
	if esc(msg) {
		return true, nil
	}
	list := p.filtered()
	switch pressed(msg) {
	case "up", "ctrl+p":
		p.sel = max(p.sel-1, 0)
		return false, nil
	case "down", "ctrl+n":
		p.sel = min(p.sel+1, max(len(list)-1, 0))
		return false, nil
	case "enter":
		if len(list) == 0 {
			return true, nil
		}
		a := list[min(p.sel, len(list)-1)]
		m.ov = nil
		cmd := a.run(m)
		// The action may open another overlay; keep it open in that case.
		return m.ov == nil, cmd
	}
	var cmd tea.Cmd
	p.in, cmd = p.in.Update(msg)
	p.sel = min(p.sel, max(len(p.filtered())-1, 0))
	return false, cmd
}

func (p *palette) view(m *Model) (string, string) {
	th := m.look.Theme
	pt := card.Painter{BG: th.Mantle}
	const w = 56
	p.in.SetWidth(w - 4)
	lines := []string{pt.Fg(th.Accent).Render("› ") + p.in.View(), ""}
	list := p.filtered()
	start := max(0, p.sel-11)
	for i := start; i < len(list) && i < start+12; i++ {
		a := list[i]
		row := pt
		if i == p.sel {
			row = card.Painter{BG: th.Surface1}
		}
		name := row.Fg(th.Text).Render(" " + a.name)
		k := row.Fg(th.Muted).Render(a.keys + " ")
		lines = append(lines, spread(row, name, k, w))
	}
	if len(list) == 0 {
		lines = append(lines, pt.Fg(th.Muted).Render(" No command matches"))
	}
	lines = append(lines, "", m.footer("↑↓", "choose", "enter", "run", "esc", "close"))
	return "Commands", strings.Join(lines, "\n")
}
