package theme

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/GuitarWag/clitasks/internal/model"
)

const (
	ColorError    = lipgloss.Color("1")
	ColorSuccess  = lipgloss.Color("2")
	ColorWarn     = lipgloss.Color("3")
	ColorBlue     = lipgloss.Color("4")
	ColorMagenta  = lipgloss.Color("5")
	ColorCyan     = lipgloss.Color("6")
	ColorGray     = lipgloss.Color("8")

	ColorPriorityCritical = ColorError
	ColorPriorityHigh     = ColorWarn
	ColorPriorityMedium   = ColorBlue
	ColorPriorityLow      = ColorGray

	ColorAssignee = ColorSuccess
	ColorTag      = ColorMagenta
	ColorDue      = ColorWarn
)

var (
	Success = lipgloss.NewStyle().Foreground(ColorSuccess)
	Error   = lipgloss.NewStyle().Foreground(ColorError)
	Warn    = lipgloss.NewStyle().Foreground(ColorWarn)
	Dim     = lipgloss.NewStyle().Faint(true)
	Bold    = lipgloss.NewStyle().Bold(true)
	Cyan    = lipgloss.NewStyle().Foreground(ColorCyan)
	Green   = lipgloss.NewStyle().Foreground(ColorSuccess)
	Yellow  = lipgloss.NewStyle().Foreground(ColorWarn)
	Blue    = lipgloss.NewStyle().Foreground(ColorBlue)
	Magenta = lipgloss.NewStyle().Foreground(ColorMagenta)
	Red     = lipgloss.NewStyle().Foreground(ColorError)
	Gray    = lipgloss.NewStyle().Foreground(ColorGray)
)

func PriorityColor(p model.TaskPriority) lipgloss.Color {
	switch p {
	case model.PriorityCritical:
		return ColorPriorityCritical
	case model.PriorityHigh:
		return ColorPriorityHigh
	case model.PriorityMedium:
		return ColorPriorityMedium
	default:
		return ColorPriorityLow
	}
}

func PriorityStyle(p model.TaskPriority) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(PriorityColor(p))
}
