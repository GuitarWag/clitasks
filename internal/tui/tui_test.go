package tui

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
)

func setupBoard(t *testing.T) (*board.Board, string) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "b.md")
	s := storage.NewMarkdown(p)
	clk := func() time.Time { return time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC) }
	b, err := board.Open(s, board.WithClock(clk), board.WithRand(rand.New(rand.NewPCG(1, 2))))
	require.NoError(t, err)
	return b, p
}

func k(s string) tea.KeyPressMsg {
	switch s {
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	r := []rune(s)
	return tea.KeyPressMsg{Code: r[0], Text: s}
}

func send(m Model, k tea.KeyPressMsg) Model {
	mm, _ := m.Update(k)
	return mm.(Model)
}

func TestView_emptyBoard(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	out := m.render()
	assert.Contains(t, out, "My Board")
	assert.Contains(t, out, "TODO (0)")
	assert.Contains(t, out, "IN PROGRESS (0)")
	assert.Contains(t, out, "DONE (0)")
	assert.Contains(t, out, "BLOCKED (0)")
	assert.Contains(t, out, "(empty)")
}

func TestView_withTasks(t *testing.T) {
	b, p := setupBoard(t)
	_, _ = b.Add("first task", board.AddInput{})
	_, _ = b.Add("second", board.AddInput{Priority: model.PriorityHigh, Assignee: "alice"})
	m := newModel(b, p)

	out := m.render()
	assert.Contains(t, out, "TODO (2)")
	assert.Contains(t, out, "first task")
	assert.Contains(t, out, "second")
}

func TestNavigation_movesSelection(t *testing.T) {
	b, p := setupBoard(t)
	_, _ = b.Add("a", board.AddInput{})
	_, _ = b.Add("b", board.AddInput{})
	m := newModel(b, p)

	assert.Equal(t, 0, m.colIdx)
	assert.Equal(t, 0, m.rowIdx)

	m = send(m, k("j"))
	assert.Equal(t, 1, m.rowIdx, "down moves row")

	m = send(m, k("k"))
	assert.Equal(t, 0, m.rowIdx, "up moves row back")

	m = send(m, k("l"))
	assert.Equal(t, 1, m.colIdx, "right moves column")

	m = send(m, k("h"))
	assert.Equal(t, 0, m.colIdx, "left moves column")
}

func TestQuit_returnsQuitCmd(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	_, cmd := m.Update(k("q"))
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	assert.True(t, ok)
}

func TestAdd_modalFlow(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)

	m = send(m, k("a"))
	assert.Equal(t, modeAdd, m.mode)

	m.form.fields[fieldTitle].input.SetValue("via tui")
	m.form.fields[fieldPriority].input.SetValue("high")
	for range m.form.fields {
		m = send(m, k("enter"))
	}

	assert.Equal(t, modeBoard, m.mode)
	tasks := b.List(board.Filter{})
	require.Len(t, tasks, 1)
	assert.Equal(t, "via tui", tasks[0].Title)
	assert.Equal(t, model.PriorityHigh, tasks[0].Priority)
}

func TestAdd_requiresTitle(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("a"))
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	assert.Equal(t, modeAdd, m.mode, "form stays open on validation error")
	assert.Contains(t, m.form.err, "title is required")
}

func TestEdit_modalPrefilled(t *testing.T) {
	b, p := setupBoard(t)
	tk, _ := b.Add("orig", board.AddInput{Priority: model.PriorityMedium})
	m := newModel(b, p)

	m = send(m, k("e"))
	assert.Equal(t, modeEdit, m.mode)
	assert.Equal(t, "orig", m.form.fields[fieldTitle].input.Value())
	assert.Equal(t, "medium", m.form.fields[fieldPriority].input.Value())

	m.form.fields[fieldTitle].input.SetValue("renamed")
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	got, _ := b.Get(tk.ID)
	assert.Equal(t, "renamed", got.Title)
}

func TestDelete_modalConfirm(t *testing.T) {
	b, p := setupBoard(t)
	tk, _ := b.Add("x", board.AddInput{})
	m := newModel(b, p)

	m = send(m, k("d"))
	assert.Equal(t, modeDelete, m.mode)

	m = send(m, k("y"))
	assert.Equal(t, modeBoard, m.mode)
	_, ok := b.Get(tk.ID)
	assert.False(t, ok, "task should be gone after y")
}

func TestDelete_cancel(t *testing.T) {
	b, p := setupBoard(t)
	tk, _ := b.Add("x", board.AddInput{})
	m := newModel(b, p)
	m = send(m, k("d"))
	m = send(m, k("n"))
	assert.Equal(t, modeBoard, m.mode)
	_, ok := b.Get(tk.ID)
	assert.True(t, ok, "task still present after n")
}

func TestStatusMenu_movesTask(t *testing.T) {
	b, p := setupBoard(t)
	tk, _ := b.Add("x", board.AddInput{})
	m := newModel(b, p)

	m = send(m, k("s"))
	assert.Equal(t, modeStatus, m.mode)
	m = send(m, k("j")) // move selection to IN PROGRESS
	m = send(m, k("enter"))

	got, _ := b.Get(tk.ID)
	assert.Equal(t, model.StatusInProgress, got.Status)
}

func TestFilter_modalAndClear(t *testing.T) {
	b, p := setupBoard(t)
	_, _ = b.Add("alpha", board.AddInput{})
	_, _ = b.Add("beta", board.AddInput{})
	m := newModel(b, p)

	m = send(m, k("f"))
	assert.Equal(t, modeFilter, m.mode)
	m.filterIn.SetValue("alph")
	m = send(m, k("enter"))
	assert.Equal(t, "alph", m.filter)
	view := m.render()
	assert.Contains(t, view, "alpha")
	assert.NotContains(t, view, "beta")

	m = send(m, k("esc"))
	assert.Empty(t, m.filter)
}

func TestHelp_modal(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("?"))
	assert.Equal(t, modeHelp, m.mode)
	assert.Contains(t, m.render(), "Keybindings")
	m = send(m, k("esc"))
	assert.Equal(t, modeBoard, m.mode)
}

func TestRefresh_reloadsFile(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)

	// Simulate an out-of-band change
	b2, _ := board.Open(storage.NewMarkdown(p))
	_, _ = b2.Add("external", board.AddInput{})

	assert.Equal(t, 0, len(m.board.List(board.Filter{})), "model is stale before refresh")
	m = send(m, k("r"))
	assert.Equal(t, 1, len(m.board.List(board.Filter{})), "model has new task after refresh")
}

func TestRenderColumn_taskCountInTitle(t *testing.T) {
	b, p := setupBoard(t)
	tk1, _ := b.Add("a", board.AddInput{})
	_, _ = b.Move(tk1.ID, model.StatusDone)
	_, _ = b.Add("b", board.AddInput{})

	m := newModel(b, p)
	v := m.render()
	assert.Contains(t, v, "TODO (1)")
	assert.Contains(t, v, "DONE (1)")
}

func TestView_helpFooterShown(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	v := m.render()
	assert.Contains(t, v, "a add")
	assert.Contains(t, v, "q quit")
}

func TestForm_tabNavigation(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("a"))
	assert.Equal(t, 0, m.form.step)

	m = send(m, k("tab"))
	assert.Equal(t, 1, m.form.step)

	m = send(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Equal(t, 0, m.form.step)

	// wrap backwards from step 0
	m = send(m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift})
	assert.Equal(t, len(m.form.fields)-1, m.form.step)
}

func TestFilter_noMatches(t *testing.T) {
	b, p := setupBoard(t)
	_, _ = b.Add("apple", board.AddInput{})
	m := newModel(b, p)
	m = send(m, k("f"))
	m.filterIn.SetValue("zzz")
	m = send(m, k("enter"))

	view := m.render()
	assert.Contains(t, view, "TODO (0)")
	assert.Contains(t, view, "(empty)", "no-match column should render empty placeholder")
}

func TestWindowResize_changesLayout(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)

	mm, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 50})
	m = mm.(Model)
	assert.Equal(t, 200, m.width)
	assert.Equal(t, 50, m.height)
	// View should not panic at the new size.
	assert.NotEmpty(t, m.render())
}

func TestSubmitForm_invalidPriority(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("a"))
	m.form.fields[fieldTitle].input.SetValue("x")
	m.form.fields[fieldPriority].input.SetValue("bogus")
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	assert.Equal(t, modeAdd, m.mode)
	assert.Contains(t, strings.ToLower(m.form.err), "invalid priority")
}

func timelineModel(t *testing.T) (Model, *board.Board, []model.Task) {
	t.Helper()
	b, p := setupBoard(t)
	a, err := b.Add("design", board.AddInput{Start: "2026-06-01", DueDate: "2026-06-05"})
	require.NoError(t, err)
	c, err := b.Add("build", board.AddInput{Start: "2026-06-08", DueDate: "2026-06-12"})
	require.NoError(t, err)
	m := newModel(b, p)
	m.now = func() time.Time { return time.Date(2026, 6, 15, 12, 0, 0, 0, time.Local) }
	m.width = 80
	return m, b, []model.Task{a, c}
}

func TestTimeline_toggleAndView(t *testing.T) {
	m, _, tasks := timelineModel(t)
	m = send(m, k("t"))
	require.True(t, m.showTimeline)

	out := m.render()
	assert.Contains(t, out, tasks[0].ID+" design")
	assert.Contains(t, out, "█████")
	assert.Contains(t, out, "1 column = 1 day")
	assert.Contains(t, out, "t/esc board")

	m = send(m, k("t"))
	assert.False(t, m.showTimeline)
	assert.Contains(t, m.render(), "TODO (2)")

	m = send(m, k("t"))
	m = send(m, k("esc"))
	assert.False(t, m.showTimeline, "esc also goes back")
}

func TestTimeline_selectionDrivesActions(t *testing.T) {
	m, b, tasks := timelineModel(t)
	m = send(m, k("t"))
	m = send(m, k("j"))
	sel, ok := m.selectedTask()
	require.True(t, ok)
	assert.Equal(t, tasks[1].ID, sel.ID)

	m = send(m, k("j"))
	assert.Equal(t, 1, m.tlRow, "selection stops at the last row")

	// Edit from the timeline and return to it.
	m = send(m, k("e"))
	require.Equal(t, modeEdit, m.mode)
	assert.Equal(t, "2026-06-08", m.form.fields[fieldStart].input.Value())
	m.form.fields[fieldStart].input.SetValue("2026-06-09")
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	assert.Equal(t, modeBoard, m.mode)
	assert.True(t, m.showTimeline)
	got, _ := b.Get(tasks[1].ID)
	assert.Equal(t, "2026-06-09", got.Start)

	// Delete from the timeline clamps the selection.
	m = send(m, k("d"))
	m = send(m, k("y"))
	assert.Equal(t, 0, m.tlRow)
	assert.Len(t, b.List(board.Filter{}), 1)
}

func TestTimeline_scrollKeepsScale(t *testing.T) {
	m, _, _ := timelineModel(t)
	m = send(m, k("t"))
	first := m.timelineASCII(m.timelineLayout())
	m = send(m, k("l"))
	assert.Equal(t, first.Scale, m.tlScale)
	assert.Equal(t, first.From.AddDate(0, 0, 7), m.tlFrom)
	m = send(m, k("h"))
	m = send(m, k("h"))
	assert.Equal(t, first.From.AddDate(0, 0, -7), m.tlFrom)
}

func TestTimeline_resizeLimitsRows(t *testing.T) {
	m, b, _ := timelineModel(t)
	for range 10 {
		_, err := b.Add("more", board.AddInput{Start: "2026-06-02", DueDate: "2026-06-03"})
		require.NoError(t, err)
	}
	m = send(m, k("t"))
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: timelineChrome + 3})
	m = mm.(Model)
	for range 11 {
		m = send(m, k("j"))
	}
	assert.Contains(t, m.render(), "rows 10-12 of 12")
}

func TestForm_rejectsBadDate(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("a"))
	m.form.fields[fieldTitle].input.SetValue("x")
	m.form.fields[fieldDue].input.SetValue("tomorrow")
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	assert.Equal(t, modeAdd, m.mode)
	assert.Contains(t, m.form.err, "YYYY-MM-DD")
}

func TestForm_editKeepsLegacyDue(t *testing.T) {
	_, p := setupBoard(t)
	require.NoError(t, storage.NewMarkdown(p).Write(&model.Board{Name: "B", Tasks: []model.Task{
		{ID: "T-OLD-001", Title: "old", Status: model.StatusTodo, Priority: model.PriorityLow, DueDate: "next friday"},
	}}))
	b, err := board.Open(storage.NewMarkdown(p))
	require.NoError(t, err)
	m := newModel(b, p)
	m = send(m, k("e"))
	m.form.fields[fieldTitle].input.SetValue("renamed")
	for range m.form.fields {
		m = send(m, k("enter"))
	}
	assert.Empty(t, m.form.err)
	got, _ := b.Get("T-OLD-001")
	assert.Equal(t, "renamed", got.Title)
	assert.Equal(t, "next friday", got.DueDate)
}

func TestStatusMenu_showsMoveError(t *testing.T) {
	b, p := setupBoard(t)
	_, err := b.Add("x", board.AddInput{})
	require.NoError(t, err)
	m := newModel(b, p)

	dir := filepath.Dir(p)
	require.NoError(t, os.Chmod(dir, 0o500)) // the atomic write cannot create its temp file
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	m = send(m, k("s"))
	m = send(m, k("j"))
	m = send(m, k("enter"))
	assert.True(t, strings.HasPrefix(m.flash, "✗ "), "flash was %q", m.flash)
}
