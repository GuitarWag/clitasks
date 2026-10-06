package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/GuitarWag/clitasks/v3/internal/board"
	"github.com/GuitarWag/clitasks/v3/internal/storage"
	"github.com/GuitarWag/clitasks/v3/internal/ui"
)

func Run(filePath string, opts ui.Options) error {
	b, err := board.Open(storage.NewMarkdown(filePath))
	if err != nil {
		return err
	}
	m := newModel(b, filePath, opts)
	p := tea.NewProgram(m)
	_, err = p.Run()
	return err
}
