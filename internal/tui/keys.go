package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up, Down, Left, Right                       key.Binding
	Quit, Add, Edit, Delete, Status, Filter, Refresh, Help, Esc, Enter key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Left:    key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "left")),
		Right:   key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "right")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Status:  key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status")),
		Filter:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "filter")),
		Refresh: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Esc:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	}
}
