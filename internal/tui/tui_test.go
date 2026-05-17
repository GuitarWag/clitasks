package tui

import (
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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

func k(s string) tea.KeyMsg {
	switch s {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func send(m Model, k tea.KeyMsg) Model {
	mm, _ := m.Update(k)
	return mm.(Model)
}

func TestView_emptyBoard(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	out := m.View()
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

	out := m.View()
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
	for i := 0; i < 5; i++ {
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
	for i := 0; i < 5; i++ {
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
	for i := 0; i < 5; i++ {
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
	view := m.View()
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
	assert.Contains(t, m.View(), "Keybindings")
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
	v := m.View()
	assert.Contains(t, v, "TODO (1)")
	assert.Contains(t, v, "DONE (1)")
}

func TestView_helpFooterShown(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	v := m.View()
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

	m = send(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, 0, m.form.step)

	// wrap backwards from step 0
	m = send(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.Equal(t, len(m.form.fields)-1, m.form.step)
}

func TestFilter_noMatches(t *testing.T) {
	b, p := setupBoard(t)
	_, _ = b.Add("apple", board.AddInput{})
	m := newModel(b, p)
	m = send(m, k("f"))
	m.filterIn.SetValue("zzz")
	m = send(m, k("enter"))

	view := m.View()
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
	assert.NotEmpty(t, m.View())
}

func TestSubmitForm_invalidPriority(t *testing.T) {
	b, p := setupBoard(t)
	m := newModel(b, p)
	m = send(m, k("a"))
	m.form.fields[fieldTitle].input.SetValue("x")
	m.form.fields[fieldPriority].input.SetValue("bogus")
	for i := 0; i < 5; i++ {
		m = send(m, k("enter"))
	}
	assert.Equal(t, modeAdd, m.mode)
	assert.Contains(t, strings.ToLower(m.form.err), "invalid priority")
}
