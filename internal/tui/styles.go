package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/theme"
)

type styles struct {
	header     lipgloss.Style
	headerDim  lipgloss.Style
	colTitle   lipgloss.Style
	colBorder  lipgloss.Style
	colActive  lipgloss.Style
	task       lipgloss.Style
	taskSel    lipgloss.Style
	taskDim    lipgloss.Style
	footer     lipgloss.Style
	modalBox   lipgloss.Style
	modalLabel lipgloss.Style
	error      lipgloss.Style
	tag        lipgloss.Style
	assignee   lipgloss.Style
	due        lipgloss.Style
}

func newStyles() styles {
	return styles{
		header:     lipgloss.NewStyle().Bold(true).Foreground(theme.ColorCyan),
		headerDim:  theme.Dim,
		colTitle:   theme.Bold,
		colBorder:  lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		colActive:  lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(theme.ColorCyan).Padding(0, 1),
		task:       lipgloss.NewStyle(),
		taskSel:    lipgloss.NewStyle().Reverse(true).Bold(true),
		taskDim:    theme.Dim,
		footer:     lipgloss.NewStyle().Faint(true).MarginTop(1),
		modalBox:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		modalLabel: theme.Bold,
		error:      theme.Error,
		tag:        lipgloss.NewStyle().Foreground(theme.ColorTag),
		assignee:   lipgloss.NewStyle().Foreground(theme.ColorAssignee),
		due:        lipgloss.NewStyle().Foreground(theme.ColorDue),
	}
}

func (s styles) priority(p model.TaskPriority) lipgloss.Style {
	return theme.PriorityStyle(p)
}
