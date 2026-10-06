package tui

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/GuitarWag/clitasks/v3/internal/model"
)

func TestQuery(t *testing.T) {
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)
	task := model.Task{
		ID: "T-1", Title: "Write the Parser", Description: "legacy boards",
		Status: model.StatusInProgress, Priority: model.PriorityHigh,
		Assignee: "Bob", Tags: []string{"Backend", "parser"}, DueDate: "2026-10-10",
	}
	waiting := func(model.Task) bool { return true }
	for q, want := range map[string]bool{
		"":                  true,
		"parser":            true,
		"parser legacy":     true,
		"parser frontend":   false,
		"t-1":               true,
		"@bo":               true,
		"@alice":            false,
		"@alice @bob":       true,
		"#back":             true,
		"#back #parser":     true,
		"#back #ops":        false,
		"!high":             true,
		"!hi":               true,
		"!crit":             false,
		"!crit !high":       true,
		"!nonsense":         false, // an unknown priority is a plain word
		"is:doing":          true,
		"is:open":           true,
		"is:done":           false,
		"is:overdue":        false,
		"is:waiting":        true,
		"is:unknown":        false,
		"due:<3d":           true,
		"due:<2d":           false,
		"due:>1d":           true,
		"due:today":         false,
		"due:none":          false,
		"due:<xd":           false, // not a filter, a word that does not match
		"parser !high @bob": true,
	} {
		assert.Equal(t, want, parseQuery(q).match(task, now, waiting), "query %q", q)
	}
	assert.Equal(t, []string{"parser", "legacy"}, parseQuery("parser @bob legacy #x").terms())
}
