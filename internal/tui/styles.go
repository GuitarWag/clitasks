package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/GuitarWag/clitasks/internal/model"
)

type styles struct {
	header      lipgloss.Style
	headerDim   lipgloss.Style
	colTitle    lipgloss.Style
	colBorder   lipgloss.Style
	colActive   lipgloss.Style
	task        lipgloss.Style
	taskSel     lipgloss.Style
	taskDim     lipgloss.Style
	footer      lipgloss.Style
	modalBox    lipgloss.Style
	modalLabel  lipgloss.Style
	error       lipgloss.Style
	prioCrit    lipgloss.Style
	prioHigh    lipgloss.Style
	prioMed     lipgloss.Style
	prioLow     lipgloss.Style
	tag         lipgloss.Style
	assignee    lipgloss.Style
	due         lipgloss.Style
}

func newStyles() styles {
	return styles{
		header:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")),
		headerDim:  lipgloss.NewStyle().Faint(true),
		colTitle:   lipgloss.NewStyle().Bold(true),
		colBorder:  lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
		colActive:  lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("6")).Padding(0, 1),
		task:       lipgloss.NewStyle(),
		taskSel:    lipgloss.NewStyle().Reverse(true).Bold(true),
		taskDim:    lipgloss.NewStyle().Faint(true),
		footer:     lipgloss.NewStyle().Faint(true).MarginTop(1),
		modalBox:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2),
		modalLabel: lipgloss.NewStyle().Bold(true),
		error:      lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		prioCrit:   lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
		prioHigh:   lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		prioMed:    lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
		prioLow:    lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		tag:        lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		assignee:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		due:        lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
	}
}

func (s styles) priority(p model.TaskPriority) lipgloss.Style {
	switch p {
	case model.PriorityCritical:
		return s.prioCrit
	case model.PriorityHigh:
		return s.prioHigh
	case model.PriorityMedium:
		return s.prioMed
	default:
		return s.prioLow
	}
}
