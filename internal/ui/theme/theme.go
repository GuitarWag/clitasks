// Package theme holds the semantic color tokens of the UI and the preset
// palettes. Components use tokens (Accent, StatusDone, Surface1), never raw
// colors, so a preset change restyles everything.
package theme

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/GuitarWag/clitasks/internal/model"
)

type Theme struct {
	Name string
	Dark bool
	// Mono is true for the NO_COLOR theme: every color is NoColor, and
	// components must use bold, faint and reverse instead.
	Mono bool

	Base, Mantle, Crust          color.Color
	Surface0, Surface1, Surface2 color.Color
	Overlay0, Overlay1           color.Color
	Text, Subtext, Muted         color.Color
	Accent                       color.Color

	StatusTodo, StatusInProgress, StatusBlocked, StatusDone color.Color
	PrioCritical, PrioHigh, PrioMedium, PrioLow             color.Color
	Success, Warn, Error, Info, Today                       color.Color
}

func (t Theme) Status(s model.TaskStatus) color.Color {
	switch s {
	case model.StatusInProgress:
		return t.StatusInProgress
	case model.StatusBlocked:
		return t.StatusBlocked
	case model.StatusDone:
		return t.StatusDone
	default:
		return t.StatusTodo
	}
}

func (t Theme) Priority(p model.TaskPriority) color.Color {
	switch p {
	case model.PriorityCritical:
		return t.PrioCritical
	case model.PriorityHigh:
		return t.PrioHigh
	case model.PriorityLow:
		return t.PrioLow
	default:
		return t.PrioMedium
	}
}

// Fg is a style with a foreground color.
func Fg(c color.Color) lipgloss.Style { return lipgloss.NewStyle().Foreground(c) }

const (
	NameMocha     = "catppuccin-mocha"
	NameLatte     = "catppuccin-latte"
	NameTokyo     = "tokyo-night"
	NameNord      = "nord"
	NameMono      = "mono"
	NameAuto      = "auto"
	DefaultDark   = NameMocha
	DefaultLight  = NameLatte
	envNoColorKey = "NO_COLOR"
)

var presets = map[string]Theme{
	NameMocha: build(NameMocha, true, palette{
		base: "#1e1e2e", mantle: "#181825", crust: "#11111b",
		s0: "#313244", s1: "#45475a", s2: "#585b70", o0: "#6c7086", o1: "#7f849c",
		text: "#cdd6f4", sub: "#bac2de", accent: "#cba6f7",
		blue: "#89b4fa", yellow: "#f9e2af", red: "#f38ba8", green: "#a6e3a1",
		orange: "#fab387", info: "#74c7ec", today: "#f5c2e7",
	}),
	NameLatte: build(NameLatte, false, palette{
		base: "#eff1f5", mantle: "#e6e9ef", crust: "#dce0e8",
		// Latte's yellow, green, orange, sapphire, pink and surface1 are darkened to reach the
		// contrast that TestPresets_contrast requires on a light background.
		s0: "#dce0e8", s1: "#c6cad5", s2: "#acb0be", o0: "#9ca0b0", o1: "#7c7f93",
		text: "#4c4f69", sub: "#5c5f77", accent: "#8839ef",
		blue: "#1e66f5", yellow: "#b86e00", red: "#d20f39", green: "#368a22",
		orange: "#d45500", info: "#17859b", today: "#c14a9b",
	}),
	NameTokyo: build(NameTokyo, true, palette{
		base: "#1a1b26", mantle: "#16161e", crust: "#0f0f14",
		s0: "#24283b", s1: "#2f354f", s2: "#414868", o0: "#565f89", o1: "#737aa2",
		text: "#c0caf5", sub: "#a9b1d6", accent: "#bb9af7",
		blue: "#7aa2f7", yellow: "#e0af68", red: "#f7768e", green: "#9ece6a",
		orange: "#ff9e64", info: "#7dcfff", today: "#2ac3de",
	}),
	NameNord: build(NameNord, true, palette{
		base: "#2e3440", mantle: "#2a2f3a", crust: "#242933",
		s0: "#3b4252", s1: "#434c5e", s2: "#4c566a", o0: "#616e88", o1: "#8590a6",
		text: "#eceff4", sub: "#d8dee9", accent: "#88c0d0",
		blue: "#81a1c1", yellow: "#ebcb8b", red: "#bf616a", green: "#a3be8c",
		orange: "#d08770", info: "#8fbcbb", today: "#b48ead",
	}),
}

// Names lists the selectable themes, including auto and mono.
func Names() []string {
	out := []string{NameAuto}
	for n := range presets {
		out = append(out, n)
	}
	slices.Sort(out[1:])
	return append(out, NameMono)
}

// Resolve returns the theme for a name. "" and "auto" pick the dark or light
// default from the terminal background. noColor (NO_COLOR set) always wins.
func Resolve(name string, dark, noColor bool) (Theme, error) {
	if noColor {
		return Mono(dark), nil
	}
	switch n := strings.ToLower(strings.TrimSpace(name)); n {
	case "", NameAuto:
		if dark {
			return presets[DefaultDark], nil
		}
		return presets[DefaultLight], nil
	case NameMono:
		return Mono(dark), nil
	default:
		t, ok := presets[n]
		if !ok {
			return Theme{}, fmt.Errorf("unknown theme %q (want %s)", name, strings.Join(Names(), "|"))
		}
		return t, nil
	}
}

// NoColorEnv reports whether NO_COLOR is set to a non-empty value.
func NoColorEnv(getenv func(string) string) bool { return getenv(envNoColorKey) != "" }

// Mono is the theme with no colors.
func Mono(dark bool) Theme {
	n := lipgloss.NoColor{}
	return Theme{
		Name: NameMono, Dark: dark, Mono: true,
		Base: n, Mantle: n, Crust: n, Surface0: n, Surface1: n, Surface2: n,
		Overlay0: n, Overlay1: n, Text: n, Subtext: n, Muted: n, Accent: n,
		StatusTodo: n, StatusInProgress: n, StatusBlocked: n, StatusDone: n,
		PrioCritical: n, PrioHigh: n, PrioMedium: n, PrioLow: n,
		Success: n, Warn: n, Error: n, Info: n, Today: n,
	}
}

type palette struct {
	base, mantle, crust, s0, s1, s2, o0, o1, text, sub, accent string
	blue, yellow, red, green, orange, info, today              string
}

func build(name string, dark bool, p palette) Theme {
	c := lipgloss.Color
	return Theme{
		Name: name, Dark: dark,
		Base: c(p.base), Mantle: c(p.mantle), Crust: c(p.crust),
		Surface0: c(p.s0), Surface1: c(p.s1), Surface2: c(p.s2),
		Overlay0: c(p.o0), Overlay1: c(p.o1),
		Text: c(p.text), Subtext: c(p.sub), Muted: c(p.o1), Accent: c(p.accent),
		StatusTodo: c(p.blue), StatusInProgress: c(p.yellow), StatusBlocked: c(p.red), StatusDone: c(p.green),
		PrioCritical: c(p.red), PrioHigh: c(p.orange), PrioMedium: c(p.blue), PrioLow: c(p.o1),
		Success: c(p.green), Warn: c(p.yellow), Error: c(p.red), Info: c(p.info), Today: c(p.today),
	}
}
