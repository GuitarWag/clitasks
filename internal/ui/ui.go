// Package ui holds the parts of the interface that do not depend on Bubble
// Tea: themes, icons and renderers. The TUI and the CLI output share them.
package ui

import (
	"github.com/GuitarWag/clitasks/v3/internal/ui/icons"
	"github.com/GuitarWag/clitasks/v3/internal/ui/theme"
)

// Options are the user's look settings, from flags or env vars. Theme is
// resolved later, when the terminal background is known.
type Options struct {
	Theme   string
	Icons   string
	NoColor bool
	// TermProgram is $TERM_PROGRAM, used to pick the auto icon set.
	TermProgram string
}

// OptionsFromEnv reads TASKS_THEME, TASKS_ICONS, NO_COLOR and TERM_PROGRAM.
func OptionsFromEnv(getenv func(string) string) Options {
	return Options{
		Theme:       getenv("TASKS_THEME"),
		Icons:       getenv("TASKS_ICONS"),
		NoColor:     theme.NoColorEnv(getenv),
		TermProgram: getenv("TERM_PROGRAM"),
	}
}

// Validate reports an unknown theme or icon set before the UI starts.
func (o Options) Validate() error {
	if _, err := theme.Resolve(o.Theme, true, false); err != nil {
		return err
	}
	_, err := icons.Resolve(o.Icons, o.TermProgram)
	return err
}

// Look is a resolved theme and icon set.
type Look struct {
	Theme theme.Theme
	Icons icons.Set
}

// Resolve picks the theme for the given background. Call Validate first.
func (o Options) Resolve(dark bool) Look {
	th, _ := theme.Resolve(o.Theme, dark, o.NoColor)
	ic, _ := icons.Resolve(o.Icons, o.TermProgram)
	if o.NoColor && o.Icons == "" {
		ic = icons.ASCII
	}
	return Look{Theme: th, Icons: ic}
}
