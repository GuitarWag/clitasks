package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	env := map[string]string{"TASKS_THEME": "nord", "TERM_PROGRAM": "WezTerm"}
	o := OptionsFromEnv(func(k string) string { return env[k] })
	assert.NoError(t, o.Validate())
	l := o.Resolve(true)
	assert.Equal(t, "nord", l.Theme.Name)
	assert.Equal(t, "nerd", l.Icons.Name)

	o = Options{NoColor: true}
	l = o.Resolve(false)
	assert.True(t, l.Theme.Mono)
	assert.Equal(t, "ascii", l.Icons.Name, "NO_COLOR without an icon choice gives plain ASCII")

	assert.Error(t, Options{Theme: "x"}.Validate())
	assert.Error(t, Options{Icons: "x"}.Validate())
}
