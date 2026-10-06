package dates

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/v3/internal/model"
)

// Thursday 2026-10-08, local noon.
var now = time.Date(2026, 10, 8, 12, 0, 0, 0, time.Local)

func d(s string) time.Time {
	t, err := model.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestRelative(t *testing.T) {
	for in, want := range map[string]string{
		"2026-10-08": "today", "2026-10-09": "tomorrow", "2026-10-07": "yesterday",
		"2026-10-12": "in 4 days", "2026-10-05": "3 days ago",
	} {
		assert.Equal(t, want, Relative(d(in), now), in)
	}
}

func TestDueLabelAndUrgency(t *testing.T) {
	task := func(due string, st model.TaskStatus) model.Task { return model.Task{DueDate: due, Status: st} }
	cases := []struct {
		t     model.Task
		label string
		u     Urgency
	}{
		{task("2026-10-06", model.StatusTodo), "2d late", UrgencyLate},
		{task("2026-10-06", model.StatusDone), "Oct 6", UrgencyNone},
		{task("2026-10-08", model.StatusTodo), "today", UrgencyToday},
		{task("2026-10-09", model.StatusTodo), "tomorrow", UrgencySoon},
		{task("2026-10-11", model.StatusTodo), "Oct 11", UrgencySoon},
		{task("2026-10-20", model.StatusTodo), "Oct 20", UrgencyNone},
		{task("2027-01-05", model.StatusTodo), "Jan 5 2027", UrgencyNone},
		{task("next friday", model.StatusTodo), "next friday", UrgencyNone},
	}
	for _, c := range cases {
		assert.Equal(t, c.label, DueLabel(c.t, now), c.t.DueDate)
		assert.Equal(t, c.u, DueUrgency(c.t, now), c.t.DueDate)
	}
}

func TestParseShortcut(t *testing.T) {
	for in, want := range map[string]string{
		"":           "",
		"2026-12-01": "2026-12-01",
		"today":      "2026-10-08",
		"Tomorrow":   "2026-10-09",
		"+3d":        "2026-10-11",
		"+2w":        "2026-10-22",
		"+1m":        "2026-11-08",
		"fri":        "2026-10-09",
		"friday":     "2026-10-09",
		"thu":        "2026-10-15", // the next Thursday, not today
		"mon":        "2026-10-12",
		"next week":  "2026-10-12",
	} {
		got, err := ParseShortcut(in, now)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, bad := range []string{"soon", "+d", "+3x", "frix", "2026-13-01", "-3d"} {
		_, err := ParseShortcut(bad, now)
		assert.Error(t, err, bad)
	}
}
