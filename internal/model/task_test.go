package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseStatus(t *testing.T) {
	for _, s := range []string{"todo", "in-progress", "done", "blocked"} {
		got, err := ParseStatus(s)
		require.NoError(t, err, s)
		assert.Equal(t, TaskStatus(s), got)
	}

	_, err := ParseStatus("bogus")
	assert.Error(t, err)

	_, err = ParseStatus("")
	assert.Error(t, err)
}

func TestParsePriority(t *testing.T) {
	for _, p := range []string{"low", "medium", "high", "critical"} {
		got, err := ParsePriority(p)
		require.NoError(t, err, p)
		assert.Equal(t, TaskPriority(p), got)
	}

	_, err := ParsePriority("nope")
	assert.Error(t, err)

	_, err = ParsePriority("HIGH") // case-sensitive
	assert.Error(t, err)
}

func TestStatusValid(t *testing.T) {
	assert.True(t, StatusTodo.Valid())
	assert.True(t, StatusInProgress.Valid())
	assert.True(t, StatusDone.Valid())
	assert.True(t, StatusBlocked.Valid())
	assert.False(t, TaskStatus("").Valid())
	assert.False(t, TaskStatus("pending").Valid())
}

func TestPriorityValid(t *testing.T) {
	assert.True(t, PriorityLow.Valid())
	assert.True(t, PriorityCritical.Valid())
	assert.False(t, TaskPriority("").Valid())
	assert.False(t, TaskPriority("urgent").Valid())
}
