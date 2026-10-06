// Package timeline computes the bars of a Gantt chart from tasks. It does no
// I/O. The CLI, the TUI and the Mermaid export all render the same Layout.
package timeline

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/v3/internal/model"
)

// Bar is one task on the chart. Start and End are dates (midnight UTC) and
// both days are inside the bar.
type Bar struct {
	Task  model.Task
	Start time.Time
	End   time.Time
	// Open is true when the task has no due date and is not done, so the bar
	// runs to today.
	Open bool
}

// Warning is an after link that the dates do not respect: Task starts before
// the task it waits for ends.
type Warning struct {
	TaskID   string
	AfterID  string
	Start    time.Time
	AfterEnd time.Time
}

func (w Warning) String() string {
	return fmt.Sprintf("%s starts %s, before %s ends %s",
		w.TaskID, w.Start.Format(time.DateOnly), w.AfterID, w.AfterEnd.Format(time.DateOnly))
}

type Layout struct {
	Bars []Bar
	// Unscheduled holds tasks whose due or start is not a YYYY-MM-DD date.
	Unscheduled []model.Task
	Warnings    []Warning
	// From and To are the first and last day of all bars. Both are zero when
	// there are no bars.
	From, To time.Time
}

// Build lays out tasks. today is the date that open bars end on.
func Build(tasks []model.Task, today time.Time) Layout {
	var l Layout
	day := Day(today)
	for _, t := range tasks {
		b, ok := barFor(t, day)
		if !ok {
			l.Unscheduled = append(l.Unscheduled, t)
			continue
		}
		l.Bars = append(l.Bars, b)
	}
	slices.SortFunc(l.Bars, func(a, b Bar) int {
		if c := a.Start.Compare(b.Start); c != 0 {
			return c
		}
		return strings.Compare(a.Task.ID, b.Task.ID)
	})

	byID := make(map[string]Bar, len(l.Bars))
	for i, b := range l.Bars {
		byID[b.Task.ID] = b
		if i == 0 || b.Start.Before(l.From) {
			l.From = b.Start
		}
		if i == 0 || b.End.After(l.To) {
			l.To = b.End
		}
	}
	for _, b := range l.Bars {
		for _, dep := range b.Task.After {
			d, ok := byID[dep]
			if ok && b.Start.Before(d.End) {
				l.Warnings = append(l.Warnings, Warning{
					TaskID: b.Task.ID, AfterID: dep, Start: b.Start, AfterEnd: d.End,
				})
			}
		}
	}
	return l
}

func barFor(t model.Task, today time.Time) (Bar, bool) {
	b := Bar{Task: t, Start: Day(t.CreatedAt)}
	if t.Start != "" {
		s, err := model.ParseDate(t.Start)
		if err != nil {
			return Bar{}, false
		}
		b.Start = s
	}
	switch {
	case t.DueDate != "":
		d, err := model.ParseDate(t.DueDate)
		if err != nil {
			return Bar{}, false
		}
		b.End = d
	case t.Status == model.StatusDone:
		b.End = Day(t.UpdatedAt)
	default:
		b.End = today
		b.Open = true
	}
	if b.End.Before(b.Start) {
		b.End = b.Start
	}
	return b, true
}

// Day returns the local calendar date of t as midnight UTC, the same form
// that model.ParseDate returns.
func Day(t time.Time) time.Time {
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
