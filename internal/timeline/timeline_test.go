package timeline

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GuitarWag/clitasks/v3/internal/model"
)

func date(s string) time.Time {
	t, err := model.ParseDate(s)
	if err != nil {
		panic(err)
	}
	return t
}

// noon is local noon, so Day() returns the same date in any time zone.
func noon(s string) time.Time {
	d := date(s)
	return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.Local)
}

var today = noon("2026-06-15")

func task(id string, f func(*model.Task)) model.Task {
	t := model.Task{
		ID: id, Title: id, Status: model.StatusTodo, Priority: model.PriorityMedium,
		CreatedAt: noon("2026-06-01"), UpdatedAt: noon("2026-06-01"),
	}
	if f != nil {
		f(&t)
	}
	return t
}

func TestBuild_barEnds(t *testing.T) {
	tasks := []model.Task{
		task("due", func(t *model.Task) { t.Start = "2026-06-02"; t.DueDate = "2026-06-20" }),
		task("done", func(t *model.Task) { t.Status = model.StatusDone; t.UpdatedAt = noon("2026-06-10") }),
		task("open", nil),
	}
	l := Build(tasks, today)
	require.Len(t, l.Bars, 3)
	by := map[string]Bar{}
	for _, b := range l.Bars {
		by[b.Task.ID] = b
	}

	assert.Equal(t, date("2026-06-02"), by["due"].Start)
	assert.Equal(t, date("2026-06-20"), by["due"].End)
	assert.False(t, by["due"].Open)

	assert.Equal(t, date("2026-06-01"), by["done"].Start, "no start falls back to CreatedAt")
	assert.Equal(t, date("2026-06-10"), by["done"].End)
	assert.False(t, by["done"].Open)

	assert.Equal(t, date("2026-06-15"), by["open"].End)
	assert.True(t, by["open"].Open)

	assert.Equal(t, date("2026-06-01"), l.From)
	assert.Equal(t, date("2026-06-20"), l.To)
}

func TestBuild_futureOpenBarIsOneDay(t *testing.T) {
	l := Build([]model.Task{task("f", func(t *model.Task) { t.Start = "2026-07-01" })}, today)
	require.Len(t, l.Bars, 1)
	assert.Equal(t, l.Bars[0].Start, l.Bars[0].End)
	assert.True(t, l.Bars[0].Open)
}

func TestBuild_unscheduled(t *testing.T) {
	l := Build([]model.Task{
		task("free", func(t *model.Task) { t.DueDate = "next friday" }),
		task("badstart", func(t *model.Task) { t.Start = "soon" }),
		task("ok", nil),
	}, today)
	require.Len(t, l.Bars, 1)
	assert.Equal(t, "ok", l.Bars[0].Task.ID)
	require.Len(t, l.Unscheduled, 2)
	assert.Equal(t, "free", l.Unscheduled[0].ID)
}

func TestBuild_afterConflictWarnsWithoutMoving(t *testing.T) {
	l := Build([]model.Task{
		task("A", func(t *model.Task) { t.Start = "2026-06-01"; t.DueDate = "2026-06-10" }),
		task("B", func(t *model.Task) { t.Start = "2026-06-05"; t.DueDate = "2026-06-12"; t.After = []string{"A"} }),
		task("C", func(t *model.Task) { t.Start = "2026-06-10"; t.DueDate = "2026-06-12"; t.After = []string{"A", "gone"} }),
	}, today)
	require.Len(t, l.Warnings, 1)
	w := l.Warnings[0]
	assert.Equal(t, "B", w.TaskID)
	assert.Equal(t, "A", w.AfterID)
	assert.Equal(t, "B starts 2026-06-05, before A ends 2026-06-10", w.String())
	assert.Equal(t, date("2026-06-05"), l.Bars[1].Start, "warning must not move the bar")
}

func TestBuild_sortsByStartThenID(t *testing.T) {
	l := Build([]model.Task{
		task("b", func(t *model.Task) { t.Start = "2026-06-03" }),
		task("c", func(t *model.Task) { t.Start = "2026-06-02" }),
		task("a", func(t *model.Task) { t.Start = "2026-06-03" }),
	}, today)
	var ids []string
	for _, b := range l.Bars {
		ids = append(ids, b.Task.ID)
	}
	assert.Equal(t, []string{"c", "a", "b"}, ids)
}

func TestBuild_empty(t *testing.T) {
	l := Build(nil, today)
	assert.Empty(t, l.Bars)
	assert.True(t, l.From.IsZero())
	assert.True(t, l.To.IsZero())
}
