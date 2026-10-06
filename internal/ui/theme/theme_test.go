package theme

import (
	"image/color"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	lin := func(v uint32) float64 {
		s := float64(v) / 0xffff
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(r) + 0.7152*lin(g) + 0.0722*lin(b)
}

func contrast(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// WCAG AA: 4.5:1 for body text. Muted text, borders and bars need 3:1.
func TestPresets_contrast(t *testing.T) {
	for name, th := range presets {
		for _, bg := range []struct {
			n string
			c color.Color
		}{{"Base", th.Base}, {"Mantle", th.Mantle}} {
			assert.GreaterOrEqual(t, contrast(th.Text, bg.c), 4.5, "%s Text on %s", name, bg.n)
			assert.GreaterOrEqual(t, contrast(th.Subtext, bg.c), 4.5, "%s Subtext on %s", name, bg.n)
		}
		assert.GreaterOrEqual(t, contrast(th.Text, th.Surface1), 4.5, "%s Text on Surface1 (selected card)", name)
		for tok, c := range map[string]color.Color{
			"Muted": th.Muted, "Accent": th.Accent,
			"StatusTodo": th.StatusTodo, "StatusInProgress": th.StatusInProgress,
			"StatusBlocked": th.StatusBlocked, "StatusDone": th.StatusDone,
			"Today": th.Today, "Error": th.Error, "Warn": th.Warn, "Info": th.Info,
			"PrioCritical": th.PrioCritical, "PrioHigh": th.PrioHigh, "PrioMedium": th.PrioMedium, "PrioLow": th.PrioLow,
		} {
			assert.GreaterOrEqual(t, contrast(c, th.Base), 3.0, "%s %s on Base", name, tok)
		}
	}
}

func TestResolve(t *testing.T) {
	th, err := Resolve("", true, false)
	require.NoError(t, err)
	assert.Equal(t, DefaultDark, th.Name)

	th, _ = Resolve("auto", false, false)
	assert.Equal(t, DefaultLight, th.Name)

	th, _ = Resolve("Tokyo-Night", false, false)
	assert.Equal(t, NameTokyo, th.Name, "a named preset ignores the background")

	th, _ = Resolve("nord", true, true)
	assert.True(t, th.Mono, "NO_COLOR wins over a named theme")

	_, err = Resolve("solarized", true, false)
	assert.ErrorContains(t, err, "catppuccin-mocha")
}

func TestNames(t *testing.T) {
	assert.Equal(t, []string{"auto", "catppuccin-latte", "catppuccin-mocha", "nord", "tokyo-night", "mono"}, Names())
}
