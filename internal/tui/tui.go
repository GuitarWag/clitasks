package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/storage"
)

func Run(filePath string) error {
	b, err := board.Open(storage.NewMarkdown(filePath))
	if err != nil {
		return err
	}
	m := newModel(b, filePath)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
