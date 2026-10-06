package tui

import (
	"errors"
	"image/color"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
	"github.com/GuitarWag/clitasks/internal/timeline"
	"github.com/GuitarWag/clitasks/internal/ui"
	"github.com/GuitarWag/clitasks/internal/ui/card"
	"github.com/GuitarWag/clitasks/internal/ui/tlgrid"
)

type screen int

const (
	screenBoard screen = iota
	screenTimeline
)

var columnOrder = []model.TaskStatus{
	model.StatusTodo, model.StatusInProgress, model.StatusBlocked, model.StatusDone,
}

const (
	reloadEvery   = 2 * time.Second
	toastFor      = 3 * time.Second
	doubleClickIn = 400 * time.Millisecond
)

type toastKind int

const (
	toastInfo toastKind = iota
	toastOK
	toastWarn
	toastErr
)

type toast struct {
	text string
	kind toastKind
	seq  int
}

type (
	reloadTickMsg time.Time
	toastDoneMsg  int
)

// overlay is a floating panel over the dimmed screen.
type overlay interface {
	// update handles a message. done closes the overlay.
	update(m *Model, msg tea.Msg) (done bool, cmd tea.Cmd)
	// view returns the title and body of the panel.
	view(m *Model) (title, body string)
}

type Model struct {
	board *board.Board
	path  string
	opts  ui.Options
	look  ui.Look
	dark  bool // terminal background, for themes picked at runtime
	now   func() time.Time

	width, height int
	unfocused     bool

	screen screen
	ov     overlay
	toast  *toast
	seq    int

	searching bool
	searchIn  textinput.Model
	q         query

	// Board: the column with focus, the selected task ID in each column and
	// the first visible card in each column.
	col    int
	sel    [4]string
	scroll [4]int
	detail bool

	tl tlState

	mtime     time.Time
	reloadDue bool
	lastClick struct {
		at time.Time
		id string
	}
}

func newModel(b *board.Board, path string, opts ui.Options) Model {
	si := textinput.New()
	si.Prompt = ""
	si.Placeholder = "words  @who  #tag  !high  is:overdue  due:<7d"
	m := Model{
		board:    b,
		path:     path,
		opts:     opts,
		look:     opts.Resolve(true),
		dark:     true,
		now:      time.Now,
		searchIn: si,
		detail:   true,
		tl:       tlState{zoom: tlgrid.ZoomAuto},
	}
	m.mtime = m.fileMtime()
	return m
}

func (m Model) ctx() card.Ctx { return card.Ctx{Look: m.look, Now: m.now()} }

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.reloadTick())
}

func (m Model) reloadTick() tea.Cmd {
	return tea.Tick(reloadEvery, func(t time.Time) tea.Msg { return reloadTickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.fixScroll()
		return m, nil
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.look = m.opts.Resolve(m.dark)
		return m, nil
	case tea.FocusMsg:
		m.unfocused = false
		return m, nil
	case tea.BlurMsg:
		m.unfocused = true
		return m, nil
	case reloadTickMsg:
		return m, tea.Batch(m.checkReload(), m.reloadTick())
	case toastDoneMsg:
		if m.toast != nil && m.toast.seq == int(msg) {
			m.toast = nil
		}
		return m, nil
	}

	if m.ov != nil {
		done, cmd := m.ov.update(&m, msg)
		if done {
			m.ov = nil
			if m.reloadDue {
				m.reloadDue = false
				cmd = tea.Batch(cmd, m.reload("Reloaded · tasks.md changed while you edited"))
			}
			m.fixScroll()
		}
		return m, cmd
	}

	if m.searching {
		return m.updateSearch(msg)
	}

	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		return m.mouseClick(msg)
	case tea.MouseWheelMsg:
		return m.mouseWheel(msg)
	case tea.KeyPressMsg:
		if cmd, ok := m.globalKey(msg); ok {
			return m, cmd
		}
		if m.screen == screenTimeline {
			return m.timelineKey(msg)
		}
		return m.boardKey(msg)
	}
	return m, nil
}

// globalKey handles the keys that work the same on every screen.
func (m *Model) globalKey(k tea.KeyPressMsg) (tea.Cmd, bool) {
	switch {
	case key.Matches(k, gk.Quit):
		return tea.Quit, true
	case key.Matches(k, gk.Screen):
		// Read the selection before switching: selected() depends on the screen.
		t, ok := m.selected()
		if m.screen == screenBoard {
			m.screen = screenTimeline
		} else {
			m.screen = screenBoard
		}
		if ok && m.visible(t) {
			m.selectTask(t)
		}
		m.fixScroll()
		return nil, true
	case key.Matches(k, gk.Esc):
		if !m.q.empty() {
			m.q = query{}
			m.searchIn.SetValue("")
			m.fixScroll()
		}
		return nil, true
	case key.Matches(k, gk.Search):
		m.searching = true
		m.searchIn.SetValue(m.q.raw)
		m.searchIn.CursorEnd()
		return m.searchIn.Focus(), true
	case key.Matches(k, gk.Palette):
		m.ov = newPalette(m)
		return nil, true
	case key.Matches(k, gk.Help):
		m.ov = &helpOverlay{}
		return nil, true
	case key.Matches(k, gk.Reload):
		return m.reload("Reloaded"), true
	case key.Matches(k, gk.Add):
		f, cmd := newForm(m, nil)
		m.ov = f
		return cmd, true
	case key.Matches(k, gk.Edit):
		if t, ok := m.selected(); ok {
			f, cmd := newForm(m, &t)
			m.ov = f
			return cmd, true
		}
		return nil, true
	case key.Matches(k, gk.Delete):
		if t, ok := m.selected(); ok {
			m.ov = &confirmOverlay{task: t, dependents: m.board.Dependents(t.ID)}
		}
		return nil, true
	case key.Matches(k, gk.Status):
		if t, ok := m.selected(); ok {
			m.ov = newStatusPicker(t)
		}
		return nil, true
	case key.Matches(k, gk.Detail):
		if m.screen == screenBoard && m.paneMode() == paneNone {
			if t, ok := m.selected(); ok {
				m.ov = &detailOverlay{id: t.ID}
			}
			return nil, true
		}
		m.detail = !m.detail
		m.fixScroll()
		return nil, true
	}
	return nil, false
}

func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc":
			m.searching = false
			m.searchIn.Blur()
			m.q = query{}
			m.searchIn.SetValue("")
			m.fixScroll()
			return m, nil
		case "enter":
			m.searching = false
			m.searchIn.Blur()
			m.fixScroll()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.searchIn, cmd = m.searchIn.Update(msg)
	m.q = parseQuery(m.searchIn.Value())
	m.fixScroll()
	return m, cmd
}

// --- data ---

// visible reports whether t passes the search.
func (m Model) visible(t model.Task) bool {
	return m.q.empty() || m.q.match(t, m.now(), m.waiting)
}

// waiting reports whether t waits for a task that is not done.
func (m Model) waiting(t model.Task) bool {
	for _, id := range t.After {
		if o, ok := m.board.Get(id); ok && o.Status != model.StatusDone {
			return true
		}
	}
	return false
}

func (m Model) layout() timeline.Layout {
	var tasks []model.Task
	for _, t := range m.board.Info().Tasks {
		if m.visible(t) {
			tasks = append(tasks, t)
		}
	}
	return timeline.Build(tasks, m.now())
}

func (m Model) conflicts() map[string]bool {
	out := map[string]bool{}
	for _, w := range timeline.Build(m.board.Info().Tasks, m.now()).Warnings {
		out[w.TaskID] = true
	}
	return out
}

func (m Model) related(t model.Task) card.Related {
	var r card.Related
	for _, id := range t.After {
		if o, ok := m.board.Get(id); ok {
			r.After = append(r.After, o)
		}
	}
	for _, id := range m.board.Dependents(t.ID) {
		if o, ok := m.board.Get(id); ok {
			r.Dependents = append(r.Dependents, o)
		}
	}
	for _, w := range timeline.Build(m.board.Info().Tasks, m.now()).Warnings {
		if w.TaskID == t.ID {
			r.Conflicts = append(r.Conflicts, w.String())
		}
	}
	return r
}

// selected is the task under the cursor on the current screen.
func (m Model) selected() (model.Task, bool) {
	if m.screen == screenTimeline {
		if t, ok := m.board.Get(m.tl.sel); ok && m.visible(t) {
			return t, true
		}
		rows, l := m.tlRows()
		for _, r := range rows {
			if r.Bar >= 0 {
				return l.Bars[r.Bar].Task, true
			}
		}
		return model.Task{}, false
	}
	tasks := m.columnTasks(m.col)
	if i := indexOfID(tasks, m.sel[m.col]); i >= 0 {
		return tasks[i], true
	}
	if len(tasks) > 0 {
		return tasks[0], true
	}
	return model.Task{}, false
}

// selectTask focuses t on the board: its column and its card.
func (m *Model) selectTask(t model.Task) {
	for i, s := range columnOrder {
		if s == t.Status {
			m.col = i
			m.sel[i] = t.ID
		}
	}
	m.tl.sel = t.ID
}

func indexOfID(tasks []model.Task, id string) int {
	for i, t := range tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// --- writes ---

// write runs one board change, then shows its result as a toast. Every
// write in the TUI goes through here, so no error is dropped.
func (m *Model) write(okText string, fn func() error) tea.Cmd {
	if err := fn(); err != nil {
		return m.showToast(errText(err), toastErr)
	}
	m.mtime = m.fileMtime()
	m.fixScroll()
	return m.showToast(okText, toastOK)
}

func errText(err error) string {
	if errors.Is(err, board.ErrNotFound) {
		return "Task not found. It may have been deleted on disk; press r to reload"
	}
	return err.Error()
}

func (m *Model) showToast(text string, kind toastKind) tea.Cmd {
	m.seq++
	m.toast = &toast{text: text, kind: kind, seq: m.seq}
	seq := m.seq
	return tea.Tick(toastFor, func(time.Time) tea.Msg { return toastDoneMsg(seq) })
}

// --- live reload ---

func (m Model) fileMtime() time.Time {
	if fi, err := os.Stat(m.path); err == nil {
		return fi.ModTime()
	}
	return time.Time{}
}

func (m *Model) checkReload() tea.Cmd {
	mt := m.fileMtime()
	if mt.Equal(m.mtime) {
		return nil
	}
	if _, editing := m.ov.(*formOverlay); editing {
		if m.reloadDue {
			return nil
		}
		m.reloadDue = true
		return m.showToast("tasks.md changed on disk · reloads when you close the form", toastWarn)
	}
	return m.reload("Reloaded · tasks.md changed on disk")
}

// reload reads the board file again. Selections are task IDs, so they
// survive the reload.
func (m *Model) reload(text string) tea.Cmd {
	b, err := board.Open(storage.NewMarkdown(m.path))
	if err != nil {
		return m.showToast("Reload failed: "+err.Error(), toastErr)
	}
	m.board = b
	m.mtime = m.fileMtime()
	m.fixScroll()
	return m.showToast(text, toastInfo)
}

// --- frame ---

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.ReportFocus = true
	v.WindowTitle = "tasks · " + m.board.Info().Name
	if !m.look.Theme.Mono {
		v.BackgroundColor = m.look.Theme.Base
		v.ForegroundColor = m.look.Theme.Text
	}
	return v
}

// render draws the whole screen as a string of exactly width × height.
func (m Model) render() string {
	w, h := m.size()
	th := m.look.Theme
	var body string
	if m.screen == screenTimeline {
		body = m.renderTimeline(w, h-3)
	} else {
		body = m.renderBoard(w, h-3)
	}
	blank := lipgloss.NewStyle().Background(th.Base).Width(w).Render("")
	screen := strings.Join([]string{m.renderHeader(w), blank, fit(body, w, h-3, th.Base), m.renderStatus(w)}, "\n")

	var layers []*lipgloss.Layer
	if m.ov != nil || m.unfocused {
		screen = dim(screen, m.look)
	}
	layers = append(layers, lipgloss.NewLayer(screen))
	if m.ov != nil {
		box := m.overlayBox(w, h)
		bw, bh := lipgloss.Width(box), lipgloss.Height(box)
		layers = append(layers, lipgloss.NewLayer(box).X(max((w-bw)/2, 0)).Y(max((h-bh)/2, 0)).Z(1))
	}
	if m.toast != nil {
		t := m.renderToast()
		layers = append(layers, lipgloss.NewLayer(t).X(max(w-lipgloss.Width(t)-1, 0)).Y(max(h-2, 0)).Z(2))
	}
	if len(layers) == 1 {
		return screen
	}
	return lipgloss.NewCanvas(w, h).Compose(lipgloss.NewCompositor(layers...)).Render()
}

func (m Model) size() (int, int) {
	w, h := m.width, m.height
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 36
	}
	return max(w, 40), max(h, 12)
}

// fit pads or cuts s to exactly w × h on bg.
func fit(s string, w, h int, bg color.Color) string {
	p := card.Painter{BG: bg}
	lines := strings.Split(s, "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	for i, l := range lines {
		lines[i] = p.Pad(l, w)
	}
	return strings.Join(lines, "\n")
}

// dim redraws the screen in one faint color, under an overlay or when the
// terminal loses focus.
func dim(s string, look ui.Look) string {
	th := look.Theme
	st := lipgloss.NewStyle().Foreground(th.Overlay0).Background(th.Base)
	if th.Mono {
		st = lipgloss.NewStyle().Faint(true)
	}
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = st.Render(ansi.Strip(l))
	}
	return strings.Join(lines, "\n")
}
