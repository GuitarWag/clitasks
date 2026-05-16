package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
)

type formField struct {
	label string
	input textinput.Model
}

type taskForm struct {
	editingID string
	fields    []formField
	step      int
	err       string
}

const (
	fieldTitle = iota
	fieldDesc
	fieldPriority
	fieldAssignee
	fieldTags
)

func newAddForm(_ styles) taskForm {
	return taskForm{fields: makeFields("", "", "medium", "", "")}
}

func newEditForm(_ styles, t model.Task) taskForm {
	return taskForm{
		editingID: t.ID,
		fields: makeFields(
			t.Title, t.Description, string(t.Priority), t.Assignee, strings.Join(t.Tags, ","),
		),
	}
}

func makeFields(title, desc, prio, assn, tags string) []formField {
	mk := func(label, val, placeholder string) formField {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.SetValue(val)
		ti.CharLimit = 256
		ti.Width = 60
		return formField{label: label, input: ti}
	}
	return []formField{
		mk("Title", title, "task title"),
		mk("Description", desc, "(optional)"),
		mk("Priority", prio, "low|medium|high|critical"),
		mk("Assignee", assn, "(optional)"),
		mk("Tags", tags, "comma,separated"),
	}
}

func (f *taskForm) focusFirst() tea.Cmd {
	if len(f.fields) == 0 {
		return nil
	}
	f.step = 0
	return f.fields[0].input.Focus()
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch km.String() {
	case "esc":
		m.mode = modeBoard
		return m, nil
	case "enter":
		if m.form.step < len(m.form.fields)-1 {
			m.form.fields[m.form.step].input.Blur()
			m.form.step++
			return m, m.form.fields[m.form.step].input.Focus()
		}
		return m.submitForm()
	case "tab":
		m.form.fields[m.form.step].input.Blur()
		m.form.step = (m.form.step + 1) % len(m.form.fields)
		return m, m.form.fields[m.form.step].input.Focus()
	case "shift+tab":
		m.form.fields[m.form.step].input.Blur()
		m.form.step = (m.form.step - 1 + len(m.form.fields)) % len(m.form.fields)
		return m, m.form.fields[m.form.step].input.Focus()
	}

	var cmd tea.Cmd
	m.form.fields[m.form.step].input, cmd = m.form.fields[m.form.step].input.Update(msg)
	return m, cmd
}

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.form.fields[fieldTitle].input.Value())
	if title == "" {
		m.form.err = "title is required"
		return m, nil
	}

	priorityStr := strings.TrimSpace(m.form.fields[fieldPriority].input.Value())
	if priorityStr == "" {
		priorityStr = "medium"
	}
	priority, err := model.ParsePriority(priorityStr)
	if err != nil {
		m.form.err = err.Error()
		return m, nil
	}

	desc := strings.TrimSpace(m.form.fields[fieldDesc].input.Value())
	assignee := strings.TrimSpace(m.form.fields[fieldAssignee].input.Value())
	tagsStr := strings.TrimSpace(m.form.fields[fieldTags].input.Value())
	var tags []string
	for _, t := range strings.Split(tagsStr, ",") {
		if v := strings.TrimSpace(t); v != "" {
			tags = append(tags, v)
		}
	}

	if m.form.editingID == "" {
		_, err = m.board.Add(title, board.AddInput{
			Description: desc,
			Priority:    priority,
			Assignee:    assignee,
			Tags:        tags,
			DueDate:     "",
		})
	} else {
		tagsVal := tags
		_, err = m.board.Update(m.form.editingID, board.UpdateInput{
			Title:       &title,
			Description: &desc,
			Priority:    &priority,
			Assignee:    &assignee,
			Tags:        &tagsVal,
		})
	}
	if err != nil {
		m.form.err = err.Error()
		return m, nil
	}
	m.mode = modeBoard
	m.flash = "saved"
	m.clampSelection()
	return m, nil
}

func (m Model) viewForm() string {
	title := "Add task"
	if m.form.editingID != "" {
		title = "Edit task " + m.form.editingID
	}
	var b strings.Builder
	fmt.Fprintln(&b, m.styles.modalLabel.Render(title))
	fmt.Fprintln(&b)
	for i, f := range m.form.fields {
		marker := "  "
		if i == m.form.step {
			marker = "▶ "
		}
		fmt.Fprintf(&b, "%s%s\n%s\n\n", marker, m.styles.modalLabel.Render(f.label), f.input.View())
	}
	if m.form.err != "" {
		fmt.Fprintln(&b, m.styles.error.Render("✗ "+m.form.err))
	}
	fmt.Fprintln(&b, m.styles.headerDim.Render("enter: next/save · tab/shift+tab: navigate · esc: cancel"))
	return lipgloss.NewStyle().Padding(1, 2).Render(m.styles.modalBox.Render(b.String()))
}

// --- delete ---

func (m Model) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "y", "Y":
		if t, ok := m.selectedTask(); ok {
			if err := m.board.Delete(t.ID); err == nil {
				m.flash = "deleted " + t.ID
			}
			m.clampSelection()
		}
		m.mode = modeBoard
	case "n", "N", "esc":
		m.mode = modeBoard
	}
	return m, nil
}

func (m Model) viewDelete() string {
	t, ok := m.selectedTask()
	if !ok {
		return m.viewBoard()
	}
	body := fmt.Sprintf("Delete task %s (%s)?\n\n[y] yes  [n] no",
		m.styles.modalLabel.Render(t.ID), t.Title)
	return lipgloss.NewStyle().Padding(1, 2).Render(m.styles.modalBox.Render(body))
}

// --- status menu ---

func (m Model) updateStatusMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch km.String() {
	case "up", "k":
		if m.statusSel > 0 {
			m.statusSel--
		}
	case "down", "j":
		if m.statusSel < len(columnOrder)-1 {
			m.statusSel++
		}
	case "enter":
		if t, ok := m.selectedTask(); ok {
			if _, err := m.board.Move(t.ID, columnOrder[m.statusSel]); err == nil {
				m.flash = "moved " + t.ID + " → " + string(columnOrder[m.statusSel])
			}
			m.clampSelection()
		}
		m.mode = modeBoard
	case "esc":
		m.mode = modeBoard
	}
	return m, nil
}

func (m Model) viewStatusMenu() string {
	var b strings.Builder
	fmt.Fprintln(&b, m.styles.modalLabel.Render("Move to status"))
	fmt.Fprintln(&b)
	for i, st := range columnOrder {
		marker := "  "
		label := columnLabel[st]
		if i == m.statusSel {
			marker = "▶ "
			label = m.styles.taskSel.Render(label)
		}
		fmt.Fprintf(&b, "%s%s\n", marker, label)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, m.styles.headerDim.Render("enter: confirm · esc: cancel"))
	return lipgloss.NewStyle().Padding(1, 2).Render(m.styles.modalBox.Render(b.String()))
}

// --- filter ---

func (m Model) updateFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			m.filterIn.SetValue("")
			m.filter = ""
			m.clampSelection()
			m.mode = modeBoard
			return m, nil
		case "enter":
			m.filter = strings.TrimSpace(m.filterIn.Value())
			m.clampSelection()
			m.mode = modeBoard
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.filterIn, cmd = m.filterIn.Update(msg)
	return m, cmd
}

func (m Model) viewFilter() string {
	var b strings.Builder
	fmt.Fprintln(&b, m.styles.modalLabel.Render("Filter"))
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, m.filterIn.View())
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, m.styles.headerDim.Render("enter: apply · esc: clear and close"))
	return lipgloss.NewStyle().Padding(1, 2).Render(m.styles.modalBox.Render(b.String()))
}

// --- help ---

func (m Model) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc", "q", "?":
			m.mode = modeBoard
		}
	}
	return m, nil
}

func (m Model) viewHelp() string {
	lines := []string{
		m.styles.modalLabel.Render("Keybindings"),
		"",
		"↑/k ↓/j         move selection within column",
		"←/h →/l         move between columns",
		"a               add new task",
		"e               edit selected task",
		"d               delete selected task",
		"s               move task to different status",
		"f               open filter",
		"r               reload from disk",
		"?               toggle this help",
		"esc             cancel modal · clear filter",
		"q / ctrl+c      quit",
		"",
		m.styles.headerDim.Render("press esc, q, or ? to close"),
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(m.styles.modalBox.Render(strings.Join(lines, "\n")))
}
