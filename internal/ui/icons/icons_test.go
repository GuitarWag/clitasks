package icons

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	for _, tc := range []struct{ name, term, want string }{
		{"", "WezTerm", NameNerd},
		{"auto", "ghostty", NameNerd},
		{"auto", "Apple_Terminal", NameUnicode},
		{"", "", NameUnicode},
		{"NERD", "", NameNerd},
		{"ascii", "WezTerm", NameASCII},
	} {
		s, err := Resolve(tc.name, tc.term)
		require.NoError(t, err)
		assert.Equal(t, tc.want, s.Name, "%q in %q", tc.name, tc.term)
	}
	_, err := Resolve("emoji", "")
	assert.Error(t, err)
}

// Every field of Nerd and Unicode must be set, so no icon renders as "".
func TestSets_complete(t *testing.T) {
	for _, s := range []Set{Nerd, Unicode} {
		v := reflect.ValueOf(s)
		for i := range v.NumField() {
			assert.NotEmpty(t, v.Field(i).String(), "%s.%s", s.Name, v.Type().Field(i).Name)
		}
	}
}

// Statuses must differ inside each set, or the board loses information.
func TestSets_statusesDistinct(t *testing.T) {
	for _, s := range []Set{Nerd, Unicode, ASCII} {
		seen := map[string]bool{}
		for _, g := range []string{s.Todo, s.InProgress, s.Blocked, s.Done} {
			assert.False(t, seen[g], "%s repeats %q", s.Name, g)
			seen[g] = true
		}
	}
}
