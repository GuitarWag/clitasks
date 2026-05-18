package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/storage"
)

func Run(filePath string) error {
	b, err := board.Open(storage.NewMarkdown(filePath))
	if err != nil {
		return err
	}
	m := newModel(b, filePath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
