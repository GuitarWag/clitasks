package tui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/storage"
	"github.com/GuitarWag/clitasks/internal/ui/tlgrid"
)

const (
	idDesign  = "T-MV1A2B-D01"
	idParser  = "T-MV1A2B-P02"
	idShip    = "T-MV1A2B-T03"
	idDeploy  = "T-MV1A2B-X04"
	idDocs    = "T-MV1A2B-W05"
	idFlaky   = "T-MV1A2B-R06"
	idMermaid = "T-MV1A2B-M07"
	idLegacy  = "T-MV1A2B-L09"
)

func plainView(m Model) string { return ansi.Strip(m.render()) }

func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	mm, cmd := m.Update(msg)
	return mm.(Model), cmd
}

func reopen(t *testing.T, m Model) *board.Board {
	t.Helper()
	b, err := board.Open(storage.NewMarkdown(m.path))
	require.NoError(t, err)
	return b
}

// --- frame ---

func TestRender_exactSizeEverywhere(t *testing.T) {
	for _, sz := range [][2]int{{80, 24}, {100, 30}, {140, 40}, {220, 60}} {
		for _, theme := range []string{"catppuccin-mocha", "mono"} {
			base := demoBoard(t, theme, sz[0], sz[1])
			frames := map[string]Model{
				"board":    base,
				"timeline": send(base, "tab"),
				"form":     send(base, "e"),
				"palette":  send(base, "ctrl+k"),
				"help":     send(base, "?"),
				"status":   send(base, "space"),
				"confirm":  send(base, "d"),
				"search":   send(base, "/", "p"),
			}
			for name, m := range frames {
				lines := strings.Split(m.render(), "\n")
				require.Len(t, lines, sz[1], "%s %s %v", theme, name, sz)
				for i, l := range lines {
					assert.Equal(t, sz[0], ansi.StringWidth(l), "%s %s %v line %d", theme, name, sz, i)
				}
			}
		}
	}
}

func TestView_settings(t *testing.T) {
	m := demoBoard(t, "catppuccin-mocha", 120, 36)
	v := m.View()
	assert.True(t, v.AltScreen)
	assert.Equal(t, tea.MouseModeCellMotion, v.MouseMode)
	assert.True(t, v.ReportFocus)
	assert.Equal(t, "tasks · Sprint 12", v.WindowTitle)
	assert.NotNil(t, v.BackgroundColor)

	mono := demoBoard(t, "mono", 120, 36)
	assert.Nil(t, mono.View().BackgroundColor, "NO_COLOR keeps the terminal colors")
}

func TestBackgroundColor_picksLightTheme(t *testing.T) {
	m := demoBoard(t, "", 120, 36)
	assert.Equal(t, "catppuccin-mocha", m.look.Theme.Name)
	m, _ = update(m, tea.BackgroundColorMsg{Color: color.White})
	assert.Equal(t, "catppuccin-latte", m.look.Theme.Name)
}

func TestHeader_alertsAndTabs(t *testing.T) {
	out := plainView(demoBoard(t, "mono", 160, 44))
	head := strings.Split(out, "\n")[0]
	assert.Contains(t, head, "Sprint 12")
	assert.Contains(t, head, "9 tasks")
	assert.Contains(t, head, "1 late")
	assert.Contains(t, head, "2 conflicts")
	assert.Contains(t, head, "Board")
	assert.Contains(t, head, "Timeline")
}

// --- board ---

func TestBoard_navigation(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	sel := func(m Model) string { s, _ := m.selected(); return s.ID }
	assert.Equal(t, idDocs, sel(m), "TODO is sorted as in the file")
	m = send(m, "j")
	assert.Equal(t, idFlaky, sel(m))
	m = send(m, "G")
	assert.Equal(t, idLegacy, sel(m))
	m = send(m, "g")
	assert.Equal(t, idDocs, sel(m))
	m = send(m, "l")
	assert.Equal(t, idParser, sel(m))
	m = send(m, "4")
	assert.Equal(t, idDesign, sel(m))
	m = send(m, "h", "h")
	assert.Equal(t, 1, m.col)
	assert.Equal(t, idParser, sel(m), "each column remembers its selection")
}

func TestBoard_fourColumnsWithPaneAt160(t *testing.T) {
	out := plainView(demoBoard(t, "mono", 160, 44))
	for _, c := range []string{"TODO", "IN PROGRESS", "BLOCKED", "DONE"} {
		assert.Contains(t, out, c)
	}
	assert.Contains(t, out, idDocs, "the detail pane shows the selected task's ID")
}

func TestBoard_moveWithShiftKeys(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	m = send(m, "L")
	got, _ := m.board.Get(idDocs)
	assert.Equal(t, model.StatusInProgress, got.Status)
	assert.Equal(t, 1, m.col, "the focus follows the moved task")
	s, _ := m.selected()
	assert.Equal(t, idDocs, s.ID)
	require.NotNil(t, m.toast)
	assert.Equal(t, "Moved to In progress", m.toast.text)

	disk, _ := reopen(t, m).Get(idDocs)
	assert.Equal(t, model.StatusInProgress, disk.Status, "the move is saved")

	m = send(m, "h", "H")
	assert.Equal(t, 0, m.col, "H at the first column does nothing")
}

func TestBoard_writeErrorShowsToast(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	dir := filepath.Dir(m.path)
	require.NoError(t, os.Chmod(dir, 0o500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
	m = send(m, "L")
	require.NotNil(t, m.toast)
	assert.Equal(t, toastErr, m.toast.kind)
}

func TestBoard_narrowDetailOpensOverlay(t *testing.T) {
	m := demoBoard(t, "mono", 90, 28)
	assert.Equal(t, paneNone, m.paneMode())
	m = send(m, "enter")
	_, ok := m.ov.(*detailOverlay)
	require.True(t, ok)
	assert.Contains(t, plainView(m), idDocs)
	m = send(m, "esc")
	assert.Nil(t, m.ov)
}

func TestBoard_enterTogglesPane(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	assert.Equal(t, paneRight, m.paneMode())
	m = send(m, "enter")
	assert.Equal(t, paneHidden, m.paneMode())
}

func TestBoard_scrollsToSelection(t *testing.T) {
	m := demoBoard(t, "mono", 120, 14) // compact cards, room for few
	for i := range 8 {
		_, err := m.board.Add(fmt.Sprintf("extra %d", i), board.AddInput{})
		require.NoError(t, err)
	}
	m = send(m, "G")
	assert.Positive(t, m.scroll[0], "the column scrolls to keep the last card visible")
	assert.Contains(t, plainView(m), "↑")
}

func TestBoard_emptyState(t *testing.T) {
	p := filepath.Join(t.TempDir(), "tasks.md")
	b, err := board.Open(storage.NewMarkdown(p))
	require.NoError(t, err)
	m := newModel(b, p, demoBoard(t, "mono", 1, 1).opts)
	m, _ = update(m, tea.WindowSizeMsg{Width: 100, Height: 30})
	assert.Contains(t, plainView(m), "No tasks yet")
	m = send(m, "e", "d", "space", "H")
	assert.Nil(t, m.ov, "actions on nothing do nothing")
}

// --- screens and search ---

func TestTab_keepsSelection(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	m = send(m, "l", "j") // IN PROGRESS, second card
	m = send(m, "tab")
	assert.Equal(t, screenTimeline, m.screen)
	assert.Equal(t, idShip, m.tl.sel)
	m = send(m, "k", "tab") // the timeline sorts by start: Mermaid is above Ship
	s, _ := m.selected()
	assert.Equal(t, idMermaid, s.ID, "the timeline selection comes back to the board")
	assert.Equal(t, 0, m.col)
}

func TestSearch_filtersAndClears(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	m = send(m, "/")
	assert.True(t, m.searching)
	for _, r := range "@bob" {
		m = send(m, string(r))
	}
	m = send(m, "enter")
	assert.False(t, m.searching)
	assert.Equal(t, 2, m.matchCount())
	out := plainView(m)
	assert.Contains(t, out, "Fix the flaky CSV")
	assert.NotContains(t, out, "Write the user docs")
	assert.Contains(t, out, "@bob 2 match")

	m = send(m, "esc")
	assert.True(t, m.q.empty())
	assert.Contains(t, plainView(m), "Write the user docs")
}

// --- overlays ---

func TestForm_addWithShortcutsAndLinks(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "a")
	f, ok := m.ov.(*formOverlay)
	require.True(t, ok)
	f.v.title = "Write release notes"
	f.v.start = "tomorrow"
	f.v.due = "+1w"
	f.v.after = []string{idDocs}
	f.v.status = model.StatusInProgress
	f.v.tags = " docs, release ,"
	m = send(m, "ctrl+s")
	assert.Nil(t, m.ov)
	require.NotNil(t, m.toast)
	assert.True(t, strings.HasPrefix(m.toast.text, "Added T-"), m.toast.text)

	s, _ := m.selected()
	assert.Equal(t, "Write release notes", s.Title)
	assert.Equal(t, "2026-10-09", s.Start)
	assert.Equal(t, "2026-10-15", s.DueDate)
	assert.Equal(t, []string{idDocs}, s.After)
	assert.Equal(t, model.StatusInProgress, s.Status)
	assert.Equal(t, []string{"docs", "release"}, s.Tags)
	assert.Equal(t, 1, m.col)
}

func TestForm_badDateKeepsFormOpen(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "a")
	f := m.ov.(*formOverlay)
	f.v.title = "x"
	f.v.due = "soon"
	m = send(m, "ctrl+s")
	require.NotNil(t, m.ov)
	assert.Contains(t, plainView(m), "invalid date")
}

func TestForm_cycleIsRejected(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "4") // DONE: Design
	m = send(m, "e")
	f := m.ov.(*formOverlay)
	f.v.after = []string{idDeploy} // Deploy → Ship → Parser → Design
	m = send(m, "ctrl+s")
	require.NotNil(t, m.ov)
	assert.Contains(t, plainView(m), "cycle")
}

func TestForm_editKeepsLegacyDue(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "G", "e") // Live reload task, due "next sprint"
	f := m.ov.(*formOverlay)
	f.v.title = "Live reload"
	m = send(m, "ctrl+s")
	assert.Nil(t, m.ov)
	got, _ := m.board.Get(idLegacy)
	assert.Equal(t, "Live reload", got.Title)
	assert.Equal(t, "next sprint", got.DueDate)
}

func TestForm_escAsksBeforeDroppingChanges(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "e", "esc")
	assert.Nil(t, m.ov, "an unchanged form closes at once")

	m = send(m, "e")
	m.ov.(*formOverlay).v.title = "changed"
	m = send(m, "esc")
	require.NotNil(t, m.ov)
	assert.Contains(t, plainView(m), "esc again to discard")
	m = send(m, "esc")
	assert.Nil(t, m.ov)
	got, _ := m.board.Get(idDocs)
	assert.Equal(t, "Write the user docs", got.Title)
}

func TestStatusPicker(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "space")
	require.IsType(t, &statusPicker{}, m.ov)
	m = send(m, "3")
	assert.Nil(t, m.ov)
	got, _ := m.board.Get(idDocs)
	assert.Equal(t, model.StatusBlocked, got.Status)
	assert.Equal(t, 2, m.col)
}

func TestDelete_confirmListsDependents(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "l", "j") // Ship the TUI view; Deploy waits for it
	m = send(m, "d")
	out := plainView(m)
	assert.Contains(t, out, "Delete task?")
	assert.Contains(t, out, "Deploy to production")
	m = send(m, "n")
	_, ok := m.board.Get(idShip)
	assert.True(t, ok, "n keeps the task")

	m = send(m, "d", "y")
	_, ok = m.board.Get(idShip)
	assert.False(t, ok)
	assert.Contains(t, m.toast.text, "unlinked 1")
	deploy, _ := m.board.Get(idDeploy)
	assert.Empty(t, deploy.After)
}

func TestPalette_runsActions(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "ctrl+k")
	for _, r := range "month" {
		m = send(m, string(r))
	}
	m = send(m, "enter")
	assert.Nil(t, m.ov)
	assert.Equal(t, screenTimeline, m.screen)
	assert.Equal(t, tlgrid.ZoomMonth, m.tl.zoom)

	m = send(m, "ctrl+k")
	for _, r := range "theme nord" {
		m = send(m, string(r))
	}
	m = send(m, "enter")
	assert.Equal(t, "nord", m.look.Theme.Name)

	m = send(m, "ctrl+k")
	for _, r := range "add task" {
		m = send(m, string(r))
	}
	m = send(m, "enter")
	assert.IsType(t, &formOverlay{}, m.ov, "an action that opens an overlay keeps it open")
}

func TestHelp_listsEveryBinding(t *testing.T) {
	m := send(demoBoard(t, "mono", 160, 50), "?")
	out := plainView(m)
	for _, b := range append(append(globalBindings(), boardBindings()...), timelineBindings()...) {
		assert.Contains(t, out, b.Help().Desc)
	}
}

// --- timeline ---

func TestTimeline_zoomScrollGroup(t *testing.T) {
	m := send(demoBoard(t, "mono", 160, 44), "tab")
	assert.Contains(t, plainView(m), "Day zoom")
	m = send(m, "z")
	assert.Equal(t, tlgrid.ZoomWeek, m.tl.zoom)

	m = send(m, "f")
	assert.Equal(t, tlgrid.ZoomAuto, m.tl.zoom)
	assert.True(t, m.tl.from.IsZero())

	m = send(m, "b")
	assert.Equal(t, groupStatus, m.tl.group)
	out := plainView(m)
	assert.Contains(t, out, "IN PROGRESS  2")
	assert.Contains(t, out, "group: status")
}

func TestTimeline_scrollMovesByZoomStep(t *testing.T) {
	m := send(demoBoard(t, "mono", 160, 44), "tab", "z") // week zoom
	l := m.layout()
	before := m.tlFrom(l, tlgrid.ZoomWeek)
	m = send(m, "l")
	assert.Equal(t, before.AddDate(0, 0, 7), m.tlFrom(l, tlgrid.ZoomWeek))
	m = send(m, "h", "h")
	assert.Equal(t, before.AddDate(0, 0, -7), m.tlFrom(l, tlgrid.ZoomWeek))
}

func TestTimeline_editDates(t *testing.T) {
	m := send(demoBoard(t, "mono", 160, 44), "tab")
	m.tl.sel = idDocs
	m = send(m, ">")
	got, _ := m.board.Get(idDocs)
	assert.Equal(t, "2026-10-14", got.Start)
	assert.Equal(t, "2026-10-24", got.DueDate, "a move keeps the length")

	m = send(m, "alt+<")
	got, _ = m.board.Get(idDocs)
	assert.Equal(t, "2026-10-23", got.DueDate)

	m = send(m, "[")
	got, _ = m.board.Get(idDocs)
	assert.Equal(t, "2026-10-08", got.Start)

	m.tl.sel = idDeploy // due Oct 16
	m = send(m, "]")
	got, _ = m.board.Get(idDeploy)
	assert.Equal(t, "2026-10-12", got.Start)
	assert.Equal(t, "2026-10-16", got.DueDate, "due today would be before start; the board rejects it")
	assert.Equal(t, toastErr, m.toast.kind)
}

func TestTimeline_unscheduledToggle(t *testing.T) {
	m := send(demoBoard(t, "mono", 160, 44), "tab")
	assert.Contains(t, plainView(m), "1 unscheduled")
	assert.NotContains(t, plainView(m), "due:next sprint")
	m = send(m, "u")
	assert.Contains(t, plainView(m), "due:next sprint")
}

func TestTimeline_selectionStaysVisible(t *testing.T) {
	m := send(demoBoard(t, "mono", 120, 14), "tab")
	for range 10 {
		m = send(m, "j")
	}
	assert.Positive(t, m.tl.offset)
	s, _ := m.selected()
	assert.Contains(t, plainView(m), s.Title[:10])
}

// --- mouse ---

func TestMouse_clickSelectsAndDoubleClickEdits(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	g := m.geometry(160, 41)
	pc := g.cols[0].cards[1]
	click := tea.MouseClickMsg{X: g.cols[0].x + 3, Y: bodyTop + pc.y0, Button: tea.MouseLeft}
	m, _ = update(m, click)
	s, _ := m.selected()
	assert.Equal(t, pc.id, s.ID)
	m, _ = update(m, click)
	assert.IsType(t, &formOverlay{}, m.ov)
}

func TestMouse_wheelAndTabs(t *testing.T) {
	m := demoBoard(t, "mono", 160, 44)
	m, _ = update(m, tea.MouseWheelMsg{X: 3, Y: 10, Button: tea.MouseWheelDown})
	s, _ := m.selected()
	assert.Equal(t, idFlaky, s.ID)
	m, _ = update(m, tea.MouseClickMsg{X: 155, Y: 0, Button: tea.MouseLeft})
	assert.Equal(t, screenTimeline, m.screen)
	m, _ = update(m, tea.MouseClickMsg{X: 3, Y: bodyTop + tlgrid.HeaderLines + 1, Button: tea.MouseLeft})
	assert.Equal(t, idParser, m.tl.sel, "row 2 of the timeline")
}

// --- live reload ---

func TestReload_pollsTheFile(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	other := reopen(t, m)
	_, err := other.Add("Added by an agent", board.AddInput{})
	require.NoError(t, err)
	future := time.Now().Add(time.Minute)
	require.NoError(t, os.Chtimes(m.path, future, future))

	m, _ = update(m, reloadTickMsg(time.Now()))
	assert.Contains(t, plainView(m), "Added by an agent")
	assert.Contains(t, m.toast.text, "changed on disk")
}

func TestReload_waitsForTheForm(t *testing.T) {
	m := demoBoard(t, "mono", 140, 42)
	m = send(m, "e")
	other := reopen(t, m)
	_, err := other.Add("Added by an agent", board.AddInput{})
	require.NoError(t, err)
	future := time.Now().Add(time.Minute)
	require.NoError(t, os.Chtimes(m.path, future, future))

	m, _ = update(m, reloadTickMsg(time.Now()))
	assert.True(t, m.reloadDue)
	assert.Equal(t, toastWarn, m.toast.kind)
	m = send(m, "esc")
	assert.False(t, m.reloadDue)
	_, ok := m.board.Get(idDocs)
	assert.True(t, ok)
	assert.Len(t, m.board.Info().Tasks, 10)
}

func TestToast_expires(t *testing.T) {
	m := send(demoBoard(t, "mono", 140, 42), "L")
	require.NotNil(t, m.toast)
	m, _ = update(m, toastDoneMsg(m.toast.seq))
	assert.Nil(t, m.toast)
}

func TestFocus_dimsTheScreen(t *testing.T) {
	m := demoBoard(t, "catppuccin-mocha", 140, 42)
	m, _ = update(m, tea.BlurMsg{})
	assert.True(t, m.unfocused)
	m, _ = update(m, tea.FocusMsg{})
	assert.False(t, m.unfocused)
}

// --- keymap ---

// A key does the same thing everywhere, or is used on one screen only.
func TestKeymap_noCollisions(t *testing.T) {
	keys := func(bs []key.Binding) map[string]string {
		out := map[string]string{}
		for _, b := range bs {
			for _, k := range b.Keys() {
				_, dup := out[k]
				assert.False(t, dup, "key %q bound twice in one map", k)
				out[k] = b.Help().Desc
			}
		}
		return out
	}
	global, brd, tl := keys(globalBindings()), keys(boardBindings()), keys(timelineBindings())
	for k := range brd {
		_, clash := global[k]
		assert.False(t, clash, "board key %q is also global", k)
	}
	for k := range tl {
		_, clash := global[k]
		assert.False(t, clash, "timeline key %q is also global", k)
	}
	for k, d := range brd {
		if td, ok := tl[k]; ok {
			assert.True(t, sameMotion(d, td), "board %q (%s) and timeline (%s) differ", k, d, td)
		}
	}
}

// sameMotion allows shared navigation keys whose meaning is the same kind.
func sameMotion(a, b string) bool {
	pairs := map[string]string{"up": "up", "down": "down", "column": "earlier"}
	return pairs[a] == b || a == b || (a == "column" && b == "later")
}
