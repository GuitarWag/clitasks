package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
)

type mode int

const (
	modeBoard mode = iota
	modeAdd
	modeEdit
	modeDelete
	modeStatus
	modeFilter
	modeHelp
)

var columnOrder = []model.TaskStatus{
	model.StatusTodo,
	model.StatusInProgress,
	model.StatusBlocked,
	model.StatusDone,
}

var columnLabel = map[model.TaskStatus]string{
	model.StatusTodo:       "TODO",
	model.StatusInProgress: "IN PROGRESS",
	model.StatusBlocked:    "BLOCKED",
	model.StatusDone:       "DONE",
}

type Model struct {
	board    *board.Board
	filePath string

	colIdx int
	rowIdx int
	filter string

	mode      mode
	form      taskForm
	statusSel int

	filterIn textinput.Model

	keys   keyMap
	styles styles

	width, height int
	flash         string
}

func newModel(b *board.Board, filePath string) Model {
	fi := textinput.New()
	fi.Placeholder = "filter text"
	return Model{
		board:    b,
		filePath: filePath,
		keys:     defaultKeys(),
		styles:   newStyles(),
		filterIn: fi,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	switch m.mode {
	case modeBoard:
		return m.updateBoard(msg)
	case modeAdd, modeEdit:
		return m.updateForm(msg)
	case modeDelete:
		return m.updateDelete(msg)
	case modeStatus:
		return m.updateStatusMenu(msg)
	case modeFilter:
		return m.updateFilter(msg)
	case modeHelp:
		return m.updateHelp(msg)
	}
	return m, nil
}

// --- board mode ---

func (m Model) tasksInColumn() []model.Task {
	by := m.board.ByStatus()
	return applyFilter(by[columnOrder[m.colIdx]], m.filter)
}

func applyFilter(tasks []model.Task, filter string) []model.Task {
	if filter == "" {
		return tasks
	}
	q := strings.ToLower(filter)
	out := make([]model.Task, 0, len(tasks))
	for _, t := range tasks {
		if matchesFilter(t, q) {
			out = append(out, t)
		}
	}
	return out
}

func matchesFilter(t model.Task, q string) bool {
	if strings.Contains(strings.ToLower(t.Title), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Description), q) {
		return true
	}
	if strings.Contains(strings.ToLower(t.Assignee), q) {
		return true
	}
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			return true
		}
	}
	return false
}

func (m Model) selectedTask() (model.Task, bool) {
	tasks := m.tasksInColumn()
	if m.rowIdx < 0 || m.rowIdx >= len(tasks) {
		return model.Task{}, false
	}
	return tasks[m.rowIdx], true
}

func (m *Model) clampSelection() {
	tasks := m.tasksInColumn()
	if m.rowIdx >= len(tasks) {
		m.rowIdx = len(tasks) - 1
	}
	if m.rowIdx < 0 {
		m.rowIdx = 0
	}
}

func (m Model) updateBoard(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(km, m.keys.Up):
		if m.rowIdx > 0 {
			m.rowIdx--
		}
	case key.Matches(km, m.keys.Down):
		tasks := m.tasksInColumn()
		if m.rowIdx < len(tasks)-1 {
			m.rowIdx++
		}
	case key.Matches(km, m.keys.Left):
		if m.colIdx > 0 {
			m.colIdx--
			m.rowIdx = 0
		}
	case key.Matches(km, m.keys.Right):
		if m.colIdx < len(columnOrder)-1 {
			m.colIdx++
			m.rowIdx = 0
		}
	case key.Matches(km, m.keys.Add):
		m.mode = modeAdd
		m.form = newAddForm(m.styles)
		return m, m.form.focusFirst()
	case key.Matches(km, m.keys.Edit):
		if t, ok := m.selectedTask(); ok {
			m.mode = modeEdit
			m.form = newEditForm(m.styles, t)
			return m, m.form.focusFirst()
		}
	case key.Matches(km, m.keys.Delete):
		if _, ok := m.selectedTask(); ok {
			m.mode = modeDelete
		}
	case key.Matches(km, m.keys.Status):
		if _, ok := m.selectedTask(); ok {
			m.mode = modeStatus
			m.statusSel = 0
		}
	case key.Matches(km, m.keys.Filter):
		m.mode = modeFilter
		m.filterIn.SetValue(m.filter)
		return m, m.filterIn.Focus()
	case key.Matches(km, m.keys.Refresh):
		if b, err := board.Open(storage.NewMarkdown(m.filePath)); err == nil {
			m.board = b
			m.clampSelection()
			m.flash = "reloaded"
		}
	case key.Matches(km, m.keys.Help):
		m.mode = modeHelp
	case key.Matches(km, m.keys.Esc):
		if m.filter != "" {
			m.filter = ""
			m.clampSelection()
		}
	}
	return m, nil
}

// --- view ---

func (m Model) View() string {
	switch m.mode {
	case modeAdd, modeEdit:
		return m.viewForm()
	case modeDelete:
		return m.viewDelete()
	case modeStatus:
		return m.viewStatusMenu()
	case modeFilter:
		return m.viewFilter()
	case modeHelp:
		return m.viewHelp()
	}
	return m.viewBoard()
}

func (m Model) viewBoard() string {
	info := m.board.Info()
	var b strings.Builder

	fmt.Fprintf(&b, "%s — %d tasks — %s\n",
		m.styles.header.Render(info.Name), len(info.Tasks),
		m.styles.headerDim.Render(m.filePath))
	if info.Description != "" {
		fmt.Fprintln(&b, m.styles.headerDim.Render(info.Description))
	}
	filterLine := "filter: (none — press f)"
	if m.filter != "" {
		filterLine = "filter: " + m.filter + " (esc to clear)"
	}
	fmt.Fprintln(&b, m.styles.headerDim.Render(filterLine))
	if m.flash != "" {
		fmt.Fprintln(&b, m.styles.headerDim.Render(m.flash))
	}
	b.WriteString("\n")

	cols := m.renderColumns()
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cols...))
	b.WriteString("\n")
	b.WriteString(m.styles.footer.Render("←/h →/l move column · ↑/k ↓/j move task · a add · e edit · d delete · s status · f filter · r refresh · ? help · q quit"))
	return b.String()
}

func (m Model) renderColumns() []string {
	by := m.board.ByStatus()
	colW := 30
	if m.width > 0 {
		colW = (m.width - 8) / 4
		if colW < 24 {
			colW = 24
		}
	}

	colH := 0
	if m.height > 0 {
		reserved := 6 // header + description + filter + blank + footer + margin
		if m.flash != "" {
			reserved++
		}
		colH = m.height - reserved
		if colH < 8 {
			colH = 8
		}
	}

	out := make([]string, len(columnOrder))
	for i, st := range columnOrder {
		tasks := applyFilter(by[st], m.filter)

		var inner strings.Builder
		fmt.Fprintln(&inner, m.styles.colTitle.Render(fmt.Sprintf("%s (%d)", columnLabel[st], len(tasks))))
		if len(tasks) == 0 {
			inner.WriteString(m.styles.taskDim.Render("(empty)"))
		} else {
			for j, t := range tasks {
				m.renderTaskLine(&inner, t, i == m.colIdx && j == m.rowIdx)
			}
		}

		box := m.styles.colBorder
		if i == m.colIdx {
			box = m.styles.colActive
		}
		box = box.Width(colW)
		if colH > 0 {
			box = box.Height(colH)
		}
		out[i] = box.Render(inner.String())
	}
	return out
}

func (m Model) renderTaskLine(b *strings.Builder, t model.Task, selected bool) {
	dot := m.styles.priority(t.Priority).Render("●")
	prefix := "  "
	titleStyle := m.styles.task
	if selected {
		prefix = "▶ "
		titleStyle = m.styles.taskSel
	}
	fmt.Fprintf(b, "%s%s %s\n", prefix, dot, titleStyle.Render(t.Title))
	if !selected {
		return
	}
	if t.Assignee != "" {
		fmt.Fprintf(b, "    %s\n", m.styles.assignee.Render("@"+t.Assignee))
	}
	if len(t.Tags) > 0 {
		var tagged []string
		for _, tg := range t.Tags {
			tagged = append(tagged, m.styles.tag.Render("#"+tg))
		}
		fmt.Fprintf(b, "    %s\n", strings.Join(tagged, " "))
	}
	if t.DueDate != "" {
		fmt.Fprintf(b, "    %s\n", m.styles.due.Render("due "+t.DueDate))
	}
	if t.Description != "" {
		desc := t.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		fmt.Fprintf(b, "    %s\n", m.styles.taskDim.Render(desc))
	}
}

