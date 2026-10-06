package tui

import "charm.land/bubbles/v2/key"

// Keymap rule: a key does the same thing on every screen, or it is used on
// one screen only. TestKeymap_noCollisions enforces it.

type globalKeys struct {
	Add, Edit, Delete, Status, Detail, Search, Palette, Help, Screen, Reload, Quit, Esc key.Binding
}

type boardKeys struct {
	Up, Down, Left, Right, MoveLeft, MoveRight, Top, Bottom, Col1, Col2, Col3, Col4 key.Binding
}

type timelineKeys struct {
	Up, Down, ScrollLeft, ScrollRight, PageLeft, PageRight,
	Zoom, ZoomIn, ZoomOut, Today, Fit, Group, Unscheduled,
	Earlier, Later, DueEarlier, DueLater, StartToday, DueToday key.Binding
}

func kb(help string, desc string, keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(help, desc))
}

var gk = globalKeys{
	Add:     kb("a", "add", "a"),
	Edit:    kb("e", "edit", "e"),
	Delete:  kb("d", "delete", "d"),
	Status:  kb("space", "status", "space"),
	Detail:  kb("enter", "details", "enter"),
	Search:  kb("/", "search", "/"),
	Palette: kb("ctrl+k", "commands", "ctrl+k", ":"),
	Help:    kb("?", "help", "?"),
	Screen:  kb("tab", "board/timeline", "tab", "shift+tab"),
	Reload:  kb("r", "reload", "r"),
	Quit:    kb("q", "quit", "q", "ctrl+c"),
	Esc:     kb("esc", "clear", "esc"),
}

var bk = boardKeys{
	Up:        kb("↑/k", "up", "up", "k"),
	Down:      kb("↓/j", "down", "down", "j"),
	Left:      kb("←/h", "column", "left", "h"),
	Right:     kb("→/l", "column", "right", "l"),
	MoveLeft:  kb("H", "move left", "H", "shift+left"),
	MoveRight: kb("L", "move right", "L", "shift+right"),
	Top:       kb("g", "top", "g", "home"),
	Bottom:    kb("G", "bottom", "G", "end"),
	Col1:      kb("1", "todo", "1"),
	Col2:      kb("2", "in progress", "2"),
	Col3:      kb("3", "blocked", "3"),
	Col4:      kb("4", "done", "4"),
}

var tk = timelineKeys{
	Up:          kb("↑/k", "up", "up", "k"),
	Down:        kb("↓/j", "down", "down", "j"),
	ScrollLeft:  kb("←/h", "earlier", "left", "h"),
	ScrollRight: kb("→/l", "later", "right", "l"),
	PageLeft:    kb("{", "page back", "{", "pgup"),
	PageRight:   kb("}", "page forward", "}", "pgdown"),
	Zoom:        kb("z", "zoom", "z"),
	ZoomIn:      kb("+", "zoom in", "+", "="),
	ZoomOut:     kb("-", "zoom out", "-"),
	Today:       kb(".", "today", "."),
	Fit:         kb("f", "fit", "f"),
	Group:       kb("b", "group by", "b"),
	Unscheduled: kb("u", "unscheduled", "u"),
	Earlier:     kb("<", "bar −1d", "<"),
	Later:       kb(">", "bar +1d", ">"),
	DueEarlier:  kb("alt+<", "due −1d", "alt+<", "alt+,"),
	DueLater:    kb("alt+>", "due +1d", "alt+>", "alt+."),
	StartToday:  kb("[", "start today", "["),
	DueToday:    kb("]", "due today", "]"),
}

func globalBindings() []key.Binding {
	g := gk
	return []key.Binding{g.Add, g.Edit, g.Delete, g.Status, g.Detail, g.Search, g.Palette, g.Help, g.Screen, g.Reload, g.Quit, g.Esc}
}

func boardBindings() []key.Binding {
	b := bk
	return []key.Binding{b.Up, b.Down, b.Left, b.Right, b.MoveLeft, b.MoveRight, b.Top, b.Bottom, b.Col1, b.Col2, b.Col3, b.Col4}
}

func timelineBindings() []key.Binding {
	t := tk
	return []key.Binding{t.Up, t.Down, t.ScrollLeft, t.ScrollRight, t.PageLeft, t.PageRight,
		t.Zoom, t.ZoomIn, t.ZoomOut, t.Today, t.Fit, t.Group, t.Unscheduled,
		t.Earlier, t.Later, t.DueEarlier, t.DueLater, t.StartToday, t.DueToday}
}

// hints are the short help of the status bar, per screen.
func (m Model) hints() []key.Binding {
	if m.screen == screenTimeline {
		return []key.Binding{tk.Up, tk.ScrollLeft, tk.Zoom, tk.Today, tk.Group, tk.Earlier, gk.Edit, gk.Search, gk.Palette, gk.Help}
	}
	return []key.Binding{bk.Left, bk.Up, bk.MoveRight, gk.Add, gk.Edit, gk.Status, gk.Search, gk.Palette, gk.Help}
}
