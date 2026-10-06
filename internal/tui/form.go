package tui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/ui"
	"github.com/GuitarWag/clitasks/internal/ui/card"
	"github.com/GuitarWag/clitasks/internal/ui/dates"
)

// formValues is what the form edits. huh writes through these pointers, so
// the struct lives on the heap behind formOverlay.
type formValues struct {
	title, desc, assignee, tags, start, due string
	status                                  model.TaskStatus
	priority                                model.TaskPriority
	after                                   []string
}

type formOverlay struct {
	form    *huh.Form
	editing *model.Task
	v       *formValues
	initial formValues
	err     string
	// escArmed is set by the first esc on a changed form; a second esc drops
	// the changes.
	escArmed bool
	width    int
}

const formWidthWide, formWidthNarrow = 96, 56

func newForm(m *Model, t *model.Task) (*formOverlay, tea.Cmd) {
	v := &formValues{status: model.StatusTodo, priority: model.PriorityMedium}
	if t != nil {
		*v = formValues{
			title: t.Title, desc: t.Description, assignee: t.Assignee,
			tags: strings.Join(t.Tags, ", "), start: t.Start, due: t.DueDate,
			status: t.Status, priority: t.Priority, after: slices.Clone(t.After),
		}
	}
	f := &formOverlay{editing: t, v: v, initial: *v}
	f.initial.after = slices.Clone(v.after)

	w, _ := m.size()
	f.width = formWidthNarrow
	layout := huh.LayoutDefault
	if w >= formWidthWide+10 {
		f.width = formWidthWide
		layout = huh.LayoutColumns(2)
	}
	id := ""
	if t != nil {
		id = t.ID
	}
	now := m.now()
	// An old free-form value that the user leaves as it is stays valid.
	dateCheck := func(orig string) func(string) error {
		return func(s string) error {
			if s == orig {
				return nil
			}
			_, err := dates.ParseShortcut(s, now)
			return err
		}
	}

	ic := m.look.Icons
	statusOpts := make([]huh.Option[model.TaskStatus], 0, len(columnOrder))
	for _, s := range columnOrder {
		statusOpts = append(statusOpts, huh.NewOption(ic.Status(s)+" "+card.StatusLabel(s), s))
	}
	var prioOpts []huh.Option[model.TaskPriority]
	for _, p := range []model.TaskPriority{model.PriorityCritical, model.PriorityHigh, model.PriorityMedium, model.PriorityLow} {
		prioOpts = append(prioOpts, huh.NewOption(ic.Priority(p)+" "+string(p), p))
	}
	var afterOpts []huh.Option[string]
	for _, o := range m.board.Info().Tasks {
		if o.ID == id {
			continue
		}
		label := fmt.Sprintf("%s %s", ic.Status(o.Status), ansi.Truncate(o.Title, 36, "…"))
		afterOpts = append(afterOpts, huh.NewOption(label, o.ID).Selected(slices.Contains(v.after, o.ID)))
	}

	details := huh.NewGroup(
		huh.NewInput().Title("Title").Value(&v.title).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("a title is required")
				}
				return nil
			}),
		huh.NewText().Title("Description").Value(&v.desc).Lines(3),
		huh.NewSelect[model.TaskStatus]().Title("Status").Options(statusOpts...).Value(&v.status).Inline(true),
		huh.NewSelect[model.TaskPriority]().Title("Priority").Options(prioOpts...).Value(&v.priority).Inline(true),
		huh.NewInput().Title("Assignee").Value(&v.assignee).Suggestions(m.assignees()).Placeholder("optional"),
		huh.NewInput().Title("Tags").Value(&v.tags).Placeholder(m.tagHint()),
	).Title("Task")

	scheduleFields := []huh.Field{
		huh.NewInput().Title("Start").Value(&v.start).Validate(dateCheck(f.initial.start)).
			Placeholder("today, +3d, fri, 2026-11-02"),
		huh.NewInput().Title("Due").Value(&v.due).Validate(dateCheck(f.initial.due)).
			Placeholder("tomorrow, +2w, next week"),
	}
	if len(afterOpts) > 0 {
		scheduleFields = append(scheduleFields, huh.NewMultiSelect[string]().Title("Waits for").
			Options(afterOpts...).Value(&v.after).Filterable(true).Height(min(len(afterOpts)+2, 8)).
			Validate(func(ids []string) error { return m.board.CheckAfter(id, ids) }))
	}
	schedule := huh.NewGroup(scheduleFields...).Title("Schedule")

	keys := huh.NewDefaultKeyMap()
	keys.Quit = key.NewBinding(key.WithKeys("ctrl+c"))
	f.form = huh.NewForm(details, schedule).
		WithTheme(huhTheme(m.look)).
		WithKeyMap(keys).
		WithLayout(layout).
		WithWidth(f.width).
		WithShowHelp(false).
		WithShowErrors(true)
	return f, f.form.Init()
}

func (f *formOverlay) dirty() bool {
	a, b := *f.v, f.initial
	return a.title != b.title || a.desc != b.desc || a.assignee != b.assignee || a.tags != b.tags ||
		a.start != b.start || a.due != b.due || a.status != b.status || a.priority != b.priority ||
		!slices.Equal(a.after, b.after)
}

func (f *formOverlay) update(m *Model, msg tea.Msg) (bool, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "esc":
			if !f.dirty() || f.escArmed {
				return true, nil
			}
			f.escArmed = true
			return false, nil
		case "ctrl+s":
			return f.submit(m)
		}
		f.escArmed = false
	}
	fm, cmd := f.form.Update(msg)
	if ff, ok := fm.(*huh.Form); ok {
		f.form = ff
	}
	switch f.form.State {
	case huh.StateCompleted:
		return f.submit(m)
	case huh.StateAborted:
		return true, nil
	}
	return false, cmd
}

// submit converts the form values and writes them through the board.
func (f *formOverlay) submit(m *Model) (bool, tea.Cmd) {
	v := f.v
	now := m.now()
	parse := func(s, orig string) (string, error) {
		if s == orig {
			return s, nil
		}
		return dates.ParseShortcut(s, now)
	}
	start, err := parse(v.start, f.initial.start)
	if err == nil {
		v.due, err = parse(v.due, f.initial.due)
	}
	title := strings.TrimSpace(v.title)
	if err == nil && title == "" {
		err = errors.New("a title is required")
	}
	if err != nil {
		f.err = err.Error()
		f.reopen()
		return false, nil
	}
	var tags []string
	for _, t := range strings.Split(v.tags, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	desc, assignee := strings.TrimSpace(v.desc), strings.TrimSpace(v.assignee)
	after := slices.Clone(v.after)

	var saved model.Task
	write := func() error {
		if f.editing == nil {
			t, err := m.board.Add(title, board.AddInput{
				Description: desc, Priority: v.priority, Assignee: assignee, Tags: tags,
				DueDate: v.due, Start: start, After: after,
			})
			if err != nil {
				return err
			}
			saved = t
			if v.status != model.StatusTodo {
				saved, err = m.board.Move(t.ID, v.status)
			}
			return err
		}
		in := board.UpdateInput{
			Title: &title, Description: &desc, Priority: &v.priority, Assignee: &assignee,
			Tags: &tags, Status: &v.status,
		}
		// Send dates and links only when they changed, so an old free-form
		// due date that the user did not touch still saves.
		if v.due != f.editing.DueDate {
			in.DueDate = &v.due
		}
		if start != f.editing.Start {
			in.Start = &start
		}
		if !slices.Equal(after, f.editing.After) {
			in.After = &after
		}
		var err error
		saved, err = m.board.Update(f.editing.ID, in)
		return err
	}
	if err := write(); err != nil {
		f.err = err.Error()
		f.reopen()
		return false, nil
	}
	verb := "Added "
	if f.editing != nil {
		verb = "Saved "
	}
	cmd := m.write(verb+saved.ID, func() error { return nil })
	m.selectTask(saved)
	return true, cmd
}

// reopen puts a completed form back into editing after a failed save.
func (f *formOverlay) reopen() {
	f.form.State = huh.StateNormal
}

func (f *formOverlay) view(m *Model) (string, string) {
	th := m.look.Theme
	p := card.Painter{BG: th.Mantle}
	title := "New task"
	if f.editing != nil {
		title = "Edit " + f.editing.ID
	}
	body := f.form.View()
	var extra []string
	if f.err != "" {
		extra = append(extra, p.Fg(th.Error).Render(m.look.Icons.Warn+" "+f.err))
	}
	if f.escArmed {
		extra = append(extra, p.Fg(th.Warn).Render("Unsaved changes · esc again to discard · ctrl+s to save"))
	} else {
		extra = append(extra, m.footer("enter", "next", "shift+tab", "back", "ctrl+s", "save", "esc", "cancel"))
	}
	return title, body + "\n\n" + strings.Join(extra, "\n")
}

func (m *Model) assignees() []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range m.board.Info().Tasks {
		if t.Assignee != "" && !seen[t.Assignee] {
			seen[t.Assignee] = true
			out = append(out, t.Assignee)
		}
	}
	slices.Sort(out)
	return out
}

// tagHint lists the most used tags as the placeholder of the tags field.
func (m *Model) tagHint() string {
	count := map[string]int{}
	for _, t := range m.board.Info().Tasks {
		for _, tg := range t.Tags {
			count[tg]++
		}
	}
	var tags []string
	for tg := range count {
		tags = append(tags, tg)
	}
	slices.SortFunc(tags, func(a, b string) int {
		if count[a] != count[b] {
			return count[b] - count[a]
		}
		return strings.Compare(a, b)
	})
	if len(tags) == 0 {
		return "comma, separated"
	}
	return strings.Join(tags[:min(len(tags), 4)], ", ")
}

// huhTheme maps our tokens onto huh's styles.
func huhTheme(look ui.Look) huh.Theme {
	th := look.Theme
	return huh.ThemeFunc(func(bool) *huh.Styles {
		s := huh.ThemeBase(th.Dark)
		bg := lipgloss.NewStyle().Background(th.Mantle)
		fg := bg.Foreground

		s.Form.Base = bg
		s.Group.Base = bg
		s.Group.Title = fg(th.Subtext).Bold(true).MarginBottom(1)
		s.Group.Description = fg(th.Muted)
		s.FieldSeparator = bg.SetString("\n\n")

		s.Focused.Base = s.Focused.Base.Background(th.Mantle).BorderForeground(th.Accent).BorderBackground(th.Mantle)
		s.Focused.Card = s.Focused.Base
		s.Focused.Title = fg(th.Accent).Bold(true)
		s.Focused.NoteTitle = fg(th.Accent).Bold(true)
		s.Focused.Description = fg(th.Muted)
		s.Focused.ErrorIndicator = fg(th.Error).SetString(" *")
		s.Focused.ErrorMessage = fg(th.Error)
		s.Focused.SelectSelector = fg(th.Accent).SetString("› ")
		s.Focused.NextIndicator = fg(th.Accent).MarginLeft(1).SetString("›")
		s.Focused.PrevIndicator = fg(th.Accent).MarginRight(1).SetString("‹")
		s.Focused.Option = fg(th.Text)
		s.Focused.MultiSelectSelector = fg(th.Accent).SetString("› ")
		s.Focused.SelectedOption = fg(th.Success)
		s.Focused.SelectedPrefix = fg(th.Success).SetString("◆ ")
		s.Focused.UnselectedOption = fg(th.Text)
		s.Focused.UnselectedPrefix = fg(th.Overlay0).SetString("◇ ")
		s.Focused.TextInput.Cursor = fg(th.Accent)
		s.Focused.TextInput.Placeholder = fg(th.Overlay0)
		s.Focused.TextInput.Prompt = fg(th.Accent)
		s.Focused.TextInput.Text = fg(th.Text)

		s.Blurred = s.Focused
		s.Blurred.Base = s.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
		s.Blurred.Card = s.Blurred.Base
		s.Blurred.Title = fg(th.Subtext)
		s.Blurred.TextInput.Text = fg(th.Subtext)
		s.Blurred.SelectSelector = bg.SetString("  ")
		s.Blurred.MultiSelectSelector = bg.SetString("  ")
		s.Blurred.NextIndicator = bg
		s.Blurred.PrevIndicator = bg

		s.Help.ShortKey = fg(th.Accent)
		s.Help.ShortDesc = fg(th.Muted)
		s.Help.ShortSeparator = fg(th.Overlay0)
		s.Help.FullKey = fg(th.Accent)
		s.Help.FullDesc = fg(th.Muted)
		s.Help.FullSeparator = fg(th.Overlay0)
		s.Help.Ellipsis = fg(th.Overlay0)
		return s
	})
}
