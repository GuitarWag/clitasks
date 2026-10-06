// Package dates formats task dates for people and parses the date shortcuts
// that the task form accepts. Storage keeps YYYY-MM-DD; this package only
// converts at the edges.
package dates

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/timeline"
)

const day = 24 * time.Hour

// DaysUntil returns the number of calendar days from today to d. Both are
// dates as model.ParseDate returns them; today may be any time of that day.
func DaysUntil(d, now time.Time) int {
	return int(d.Sub(timeline.Day(now)) / day)
}

// Relative describes d against today: "today", "tomorrow", "in 4 days",
// "yesterday", "3 days ago".
func Relative(d, now time.Time) string {
	n := DaysUntil(d, now)
	switch {
	case n == 0:
		return "today"
	case n == 1:
		return "tomorrow"
	case n == -1:
		return "yesterday"
	case n > 1:
		return fmt.Sprintf("in %d days", n)
	default:
		return fmt.Sprintf("%d days ago", -n)
	}
}

// Short is a compact date for chips: "Oct 9", or "Oct 9 2027" outside the
// current year.
func Short(d, now time.Time) string {
	if d.Year() != timeline.Day(now).Year() {
		return d.Format("Jan 2 2006")
	}
	return d.Format("Jan 2")
}

// Urgency classifies a due date for coloring.
type Urgency int

const (
	UrgencyNone Urgency = iota
	UrgencySoon         // within 3 days
	UrgencyToday
	UrgencyLate
)

func DueUrgency(t model.Task, now time.Time) Urgency {
	if t.Status == model.StatusDone {
		return UrgencyNone
	}
	d, err := model.ParseDate(t.DueDate)
	if err != nil {
		return UrgencyNone
	}
	switch n := DaysUntil(d, now); {
	case n < 0:
		return UrgencyLate
	case n == 0:
		return UrgencyToday
	case n <= 3:
		return UrgencySoon
	}
	return UrgencyNone
}

// DueLabel is the text of a due chip: "today", "2d late", "tomorrow" or the
// short date.
func DueLabel(t model.Task, now time.Time) string {
	d, err := model.ParseDate(t.DueDate)
	if err != nil {
		return t.DueDate
	}
	n := DaysUntil(d, now)
	switch {
	case t.Status != model.StatusDone && n < 0:
		return fmt.Sprintf("%dd late", -n)
	case n == 0:
		return "today"
	case n == 1:
		return "tomorrow"
	}
	return Short(d, now)
}

var weekdays = map[string]time.Weekday{
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday,
	"thu": time.Thursday, "fri": time.Friday, "sat": time.Saturday,
}

// ParseShortcut turns a form entry into YYYY-MM-DD. It accepts "" (no date),
// YYYY-MM-DD, "today", "tomorrow", "+3d", "+2w", a weekday ("fri", "friday",
// the next one after today) and "next week" (next Monday).
func ParseShortcut(s string, now time.Time) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	today := timeline.Day(now)
	out := func(d time.Time) (string, error) { return d.Format(time.DateOnly), nil }
	switch s {
	case "":
		return "", nil
	case "today", "tod":
		return out(today)
	case "tomorrow", "tom", "tmr":
		return out(today.Add(day))
	case "next week", "nw":
		return out(nextWeekday(today, time.Monday))
	}
	if strings.HasPrefix(s, "+") && len(s) > 2 {
		n, err := strconv.Atoi(s[1 : len(s)-1])
		if err == nil && n >= 0 {
			switch s[len(s)-1] {
			case 'd':
				return out(today.AddDate(0, 0, n))
			case 'w':
				return out(today.AddDate(0, 0, 7*n))
			case 'm':
				return out(today.AddDate(0, n, 0))
			}
		}
	}
	if len(s) >= 3 {
		if wd, ok := weekdays[s[:3]]; ok && strings.HasPrefix(wd.String(), strings.ToUpper(s[:1])+s[1:]) {
			return out(nextWeekday(today, wd))
		}
	}
	if d, err := model.ParseDate(s); err == nil {
		return out(d)
	}
	return "", fmt.Errorf("invalid date %q (YYYY-MM-DD, today, tomorrow, +3d, +2w, fri, next week)", s)
}

// nextWeekday returns the first wd strictly after today.
func nextWeekday(today time.Time, wd time.Weekday) time.Time {
	n := (int(wd) - int(today.Weekday()) + 7) % 7
	if n == 0 {
		n = 7
	}
	return today.AddDate(0, 0, n)
}
