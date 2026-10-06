package tui

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/v3/internal/board"
	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/timeline"
	"github.com/GuitarWag/clitasks/v3/internal/ui/card"
	"github.com/GuitarWag/clitasks/v3/internal/ui/tlgrid"
)

type groupBy int

const (
	groupNone groupBy = iota
	groupStatus
	groupAssignee
	groupTag
)

func (g groupBy) String() string {
	return [...]string{"none", "status", "assignee", "tag"}[g]
}

type tlState struct {
	sel             string // selected task ID
	offset          int    // first visible row
	zoom            tlgrid.Zoom
	from            time.Time // zero: start at the layout
	group           groupBy
	showUnscheduled bool
}

const (
	tlPaneH        = 10
	tlPaneMinBodyH = 24
)

// tlRows builds the chart rows: bars, with group headers when grouped.
func (m Model) tlRows() ([]tlgrid.Row, timeline.Layout) {
	l := m.layout()
	if m.tl.group == groupNone {
		return tlgrid.BarRows(l), l
	}
	th := m.look.Theme
	type grp struct {
		name  string
		order int
		bars  []int
	}
	groups := map[string]*grp{}
	for i, b := range l.Bars {
		t := b.Task
		var name string
		order := 0
		switch m.tl.group {
		case groupStatus:
			name = card.StatusLabel(t.Status)
			order = slices.Index(columnOrder, t.Status)
		case groupAssignee:
			name = t.Assignee
		case groupTag:
			if len(t.Tags) > 0 {
				name = t.Tags[0]
			}
		}
		if name == "" {
			name, order = "none", 1<<20
		}
		g, ok := groups[name]
		if !ok {
			g = &grp{name: name, order: order}
			groups[name] = g
		}
		g.bars = append(g.bars, i)
	}
	var list []*grp
	for _, g := range groups {
		list = append(list, g)
	}
	slices.SortFunc(list, func(a, b *grp) int {
		if a.order != b.order {
			return a.order - b.order
		}
		return strings.Compare(a.name, b.name)
	})
	var rows []tlgrid.Row
	for _, g := range list {
		color := th.Accent
		if m.tl.group == groupStatus && len(g.bars) > 0 {
			color = th.Status(l.Bars[g.bars[0]].Task.Status)
		}
		rows = append(rows, tlgrid.Row{Bar: -1, Group: g.name, GroupColor: color, Count: len(g.bars)})
		for _, i := range g.bars {
			rows = append(rows, tlgrid.Row{Bar: i})
		}
	}
	return rows, l
}

// tlGeom is the vertical split of the timeline body.
type tlGeom struct {
	chartRows int // visible rows under the two header lines
	infoY     int
	unschedY  int
	unschedH  int
	paneY     int
	paneH     int
}

func (m Model) tlGeometry(h int, l timeline.Layout) tlGeom {
	g := tlGeom{}
	used := 1 // info line
	if n := len(l.Unscheduled); n > 0 {
		g.unschedH = 1
		if m.tl.showUnscheduled {
			g.unschedH = 1 + min(n, 6)
		}
		used += g.unschedH
	}
	if m.detail && h >= tlPaneMinBodyH {
		g.paneH = tlPaneH
		used += tlPaneH + 1
	}
	g.chartRows = max(h-used-tlgrid.HeaderLines-1, 1)
	g.infoY = tlgrid.HeaderLines + g.chartRows + 1
	g.unschedY = g.infoY + 1
	g.paneY = h - g.paneH
	return g
}

func (m Model) tlZoom(l timeline.Layout, w int) tlgrid.Zoom {
	if m.tl.zoom == tlgrid.ZoomAuto {
		return tlgrid.AutoZoom(l, w)
	}
	return m.tl.zoom
}

func (m Model) tlFrom(l timeline.Layout, z tlgrid.Zoom) time.Time {
	if m.tl.from.IsZero() {
		return tlgrid.DefaultFrom(l, z, m.now())
	}
	return tlgrid.Align(m.tl.from, z)
}

func (m Model) renderTimeline(w, h int, conflicts map[string]bool) string {
	th, ic := m.look.Theme, m.look.Icons
	p := card.Painter{BG: th.Base}
	rows, l := m.tlRows()
	if len(m.board.Info().Tasks) == 0 {
		return m.emptyBoard(w, h)
	}
	if len(l.Bars) == 0 && len(l.Unscheduled) == 0 {
		msg := p.Fg(th.Muted).Render("  No task matches the search · esc clears it")
		return fit("\n"+msg, w, h, th.Base)
	}
	g := m.tlGeometry(h, l)
	zoom := m.tlZoom(l, w)
	selIdx := -1
	for i, r := range rows {
		if r.Bar >= 0 && l.Bars[r.Bar].Task.ID == m.tl.sel {
			selIdx = i
		}
	}
	related := map[string]bool{}
	if t, ok := m.board.Get(m.tl.sel); ok {
		for _, id := range t.After {
			related[id] = true
		}
		for _, id := range m.board.Dependents(t.ID) {
			related[id] = true
		}
	}
	out := tlgrid.Render(l, rows, m.ctx(), tlgrid.Opts{
		Width: w, Zoom: zoom, From: m.tlFrom(l, zoom), Selected: selIdx,
		Related: related, Conflicts: conflicts, Highlight: m.q.terms(),
		Offset: m.tl.offset, Height: g.chartRows,
	})

	lines := make([]string, 0, h)
	lines = append(lines, out.Lines...)
	lines = append(lines, "")
	info := p.Fg(th.Muted).Render(fmt.Sprintf(" %s zoom · %s – %s", strings.ToUpper(zoom.String()[:1])+zoom.String()[1:],
		out.From.Format("Jan 2"), out.To.Format("Jan 2, 2006")))
	if len(rows) > g.chartRows {
		info += p.Fg(th.Muted).Render(fmt.Sprintf(" · rows %d–%d of %d", out.First+1, out.Last, len(rows)))
	}
	right := p.Fg(th.Muted).Render(fmt.Sprintf("group: %s ", m.tl.group))
	lines = append(lines, spread(p, info, right, w))

	if n := len(l.Unscheduled); n > 0 {
		head := p.Fg(th.Warn).Render(fmt.Sprintf(" %s %d unscheduled", ic.Warn, n)) +
			p.Fg(th.Muted).Render(" · due or start is not a date · u to "+map[bool]string{true: "hide", false: "show"}[m.tl.showUnscheduled])
		lines = append(lines, head)
		if m.tl.showUnscheduled {
			for i, t := range l.Unscheduled {
				if i >= g.unschedH-1 {
					break
				}
				lines = append(lines, p.Fg(th.Status(t.Status)).Render("   "+ic.Status(t.Status)+" ")+
					p.Fg(th.Text).Render(ansi.Truncate(t.Title, w/2, "…"))+
					p.Fg(th.Error).Render("  "+badField(t)))
			}
		}
	}
	body := fit(strings.Join(lines, "\n"), w, h-g.paneH, th.Base)
	if g.paneH > 0 {
		body += "\n" + m.renderPane(w, g.paneH)
	}
	return body
}

func badField(t model.Task) string {
	var bad []string
	if _, err := model.ParseDate(t.DueDate); t.DueDate != "" && err != nil {
		bad = append(bad, "due:"+t.DueDate)
	}
	if _, err := model.ParseDate(t.Start); t.Start != "" && err != nil {
		bad = append(bad, "start:"+t.Start)
	}
	return strings.Join(bad, " ")
}

// --- input ---

func (m Model) timelineKey(k tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	w, _ := m.size()
	rows, l := m.tlRows()
	zoom := m.tlZoom(l, w)
	from := m.tlFrom(l, zoom)
	shift := func(days int) {
		m.tl.zoom = zoom
		m.tl.from = from.AddDate(0, 0, days)
	}
	visibleDays := func() int {
		out := tlgrid.Render(l, nil, m.ctx(), tlgrid.Opts{Width: w, Zoom: zoom, From: from})
		return int(out.To.Sub(out.From)/(24*time.Hour)) + 1
	}

	switch {
	case key.Matches(k, tk.Up), key.Matches(k, tk.Down):
		d := 1
		if key.Matches(k, tk.Up) {
			d = -1
		}
		m.moveTimelineSel(rows, l, d)
		m.revealBar(l, zoom, w)
	case key.Matches(k, tk.ScrollLeft):
		shift(-tlgrid.Step(zoom))
	case key.Matches(k, tk.ScrollRight):
		shift(tlgrid.Step(zoom))
	case key.Matches(k, tk.PageLeft):
		shift(-visibleDays() * 3 / 4)
	case key.Matches(k, tk.PageRight):
		shift(visibleDays() * 3 / 4)
	case key.Matches(k, tk.Zoom), key.Matches(k, tk.ZoomIn), key.Matches(k, tk.ZoomOut):
		switch {
		case key.Matches(k, tk.Zoom):
			m.tl.zoom = zoom.Next()
		case key.Matches(k, tk.ZoomIn):
			m.tl.zoom = max(zoom-1, tlgrid.ZoomDay)
		default:
			m.tl.zoom = min(zoom+1, tlgrid.ZoomMonth)
		}
		m.tl.from = time.Time{}
		m.revealBar(l, m.tl.zoom, w)
	case key.Matches(k, tk.Today):
		m.tl.zoom = zoom
		m.tl.from = timeline.Day(m.now()).AddDate(0, 0, -visibleDays()/4)
	case key.Matches(k, tk.Fit):
		m.tl.zoom, m.tl.from = tlgrid.ZoomAuto, time.Time{}
	case key.Matches(k, tk.Group):
		m.tl.group = (m.tl.group + 1) % 4
		m.tl.offset = 0
	case key.Matches(k, tk.Unscheduled):
		m.tl.showUnscheduled = !m.tl.showUnscheduled
	case key.Matches(k, tk.Earlier), key.Matches(k, tk.Later),
		key.Matches(k, tk.DueEarlier), key.Matches(k, tk.DueLater),
		key.Matches(k, tk.StartToday), key.Matches(k, tk.DueToday):
		return m, m.editDates(k, l)
	}
	m.fixTimelineScroll()
	return m, nil
}

func (m *Model) moveTimelineSel(rows []tlgrid.Row, l timeline.Layout, d int) {
	cur := -1
	for i, r := range rows {
		if r.Bar >= 0 && l.Bars[r.Bar].Task.ID == m.tl.sel {
			cur = i
		}
	}
	for i := cur + d; i >= 0 && i < len(rows); i += d {
		if rows[i].Bar >= 0 {
			m.tl.sel = l.Bars[rows[i].Bar].Task.ID
			return
		}
	}
}

// revealBar scrolls horizontally when the selected bar is off screen.
func (m *Model) revealBar(l timeline.Layout, zoom tlgrid.Zoom, w int) {
	for _, b := range l.Bars {
		if b.Task.ID != m.tl.sel {
			continue
		}
		from := m.tlFrom(l, zoom)
		out := tlgrid.Render(l, nil, m.ctx(), tlgrid.Opts{Width: w, Zoom: zoom, From: from})
		if b.End.Before(out.From) || b.Start.After(out.To) {
			m.tl.zoom = zoom
			m.tl.from = b.Start.AddDate(0, 0, -tlgrid.Step(zoom))
		}
	}
}

// editDates moves or sets the dates of the selected bar through the board,
// so the normal validation applies.
func (m *Model) editDates(k tea.KeyPressMsg, l timeline.Layout) tea.Cmd {
	var bar *timeline.Bar
	for i := range l.Bars {
		if l.Bars[i].Task.ID == m.tl.sel {
			bar = &l.Bars[i]
		}
	}
	if bar == nil {
		return m.showToast("This task has no bar to move", toastWarn)
	}
	t := bar.Task
	day := func(d time.Time) string { return d.Format(time.DateOnly) }
	// Shift the stored due date. bar.End can differ from it: the layout
	// clamps a due date before the start up to the start.
	due := bar.End
	if d, err := model.ParseDate(t.DueDate); err == nil {
		due = d
	}
	today := day(timeline.Day(m.now()))
	in := board.UpdateInput{}
	var msg string
	switch {
	case key.Matches(k, tk.Earlier), key.Matches(k, tk.Later):
		d := 1
		if key.Matches(k, tk.Earlier) {
			d = -1
		}
		s := day(bar.Start.AddDate(0, 0, d))
		in.Start = &s
		if t.DueDate != "" {
			e := day(due.AddDate(0, 0, d))
			in.DueDate = &e
		}
		msg = fmt.Sprintf("Moved %s to start %s", t.ID, s)
	case key.Matches(k, tk.DueEarlier), key.Matches(k, tk.DueLater):
		d := 1
		if key.Matches(k, tk.DueEarlier) {
			d = -1
		}
		e := day(due.AddDate(0, 0, d))
		in.DueDate = &e
		msg = "Due " + e
	case key.Matches(k, tk.StartToday):
		in.Start = &today
		msg = "Start set to today"
	case key.Matches(k, tk.DueToday):
		in.DueDate = &today
		msg = "Due set to today"
	}
	return m.write(msg, func() error {
		_, err := m.board.Update(t.ID, in)
		return err
	})
}

func (m *Model) fixTimelineScroll() {
	rows, l := m.tlRows()
	found := false
	for _, r := range rows {
		if r.Bar >= 0 && l.Bars[r.Bar].Task.ID == m.tl.sel {
			found = true
		}
	}
	if !found {
		m.tl.sel = ""
		for _, r := range rows {
			if r.Bar >= 0 {
				m.tl.sel = l.Bars[r.Bar].Task.ID
				break
			}
		}
	}
	_, h := m.size()
	g := m.tlGeometry(h-3, l)
	idx := 0
	for i, r := range rows {
		if r.Bar >= 0 && l.Bars[r.Bar].Task.ID == m.tl.sel {
			idx = i
		}
	}
	if idx < m.tl.offset {
		m.tl.offset = idx
		if idx > 0 && rows[idx-1].Bar < 0 {
			m.tl.offset = idx - 1 // keep the group header above
		}
	}
	if idx >= m.tl.offset+g.chartRows {
		m.tl.offset = idx - g.chartRows + 1
	}
	m.tl.offset = max(min(m.tl.offset, len(rows)-g.chartRows), 0)
}

func (m Model) timelineClick(x, y int) (tea.Model, tea.Cmd) {
	rows, l := m.tlRows()
	_, h := m.size()
	g := m.tlGeometry(h-3, l)
	r := y - tlgrid.HeaderLines
	if r < 0 || r >= g.chartRows {
		return m, nil
	}
	i := m.tl.offset + r
	if i >= len(rows) || rows[i].Bar < 0 {
		return m, nil
	}
	m.tl.sel = l.Bars[rows[i].Bar].Task.ID
	return m.maybeDoubleClick(m.tl.sel)
}

func (m Model) timelineWheel(mo tea.Mouse) (tea.Model, tea.Cmd) {
	rows, l := m.tlRows()
	w, _ := m.size()
	zoom := m.tlZoom(l, w)
	horizontal := mo.Mod.Contains(tea.ModShift) || mo.Button == tea.MouseWheelLeft || mo.Button == tea.MouseWheelRight
	back := mo.Button == tea.MouseWheelUp || mo.Button == tea.MouseWheelLeft
	if horizontal {
		d := tlgrid.Step(zoom)
		if back {
			d = -d
		}
		m.tl.zoom = zoom
		m.tl.from = m.tlFrom(l, zoom).AddDate(0, 0, d)
		return m, nil
	}
	d := 1
	if back {
		d = -1
	}
	m.moveTimelineSel(rows, l, d)
	m.fixTimelineScroll()
	return m, nil
}
