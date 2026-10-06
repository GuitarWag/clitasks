package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/ui/dates"
)

// query is a parsed search. Free words must all appear in the title,
// description or ID. Tokens narrow further:
//
//	@alice  #backend  !high  is:overdue|blocked|open|done|todo|doing|waiting
//	due:<7d  due:>3d  due:today  due:none
type query struct {
	raw      string
	words    []string
	assignee []string
	tags     []string
	prio     []model.TaskPriority
	is       []string
	due      []func(model.Task, time.Time) bool
}

var prioAlias = map[string]model.TaskPriority{
	"crit": model.PriorityCritical, "critical": model.PriorityCritical,
	"high": model.PriorityHigh, "hi": model.PriorityHigh,
	"medium": model.PriorityMedium, "med": model.PriorityMedium,
	"low": model.PriorityLow, "lo": model.PriorityLow,
}

func parseQuery(s string) query {
	q := query{raw: strings.TrimSpace(s)}
	for _, f := range strings.Fields(strings.ToLower(s)) {
		switch {
		case len(f) > 1 && f[0] == '@':
			q.assignee = append(q.assignee, f[1:])
		case len(f) > 1 && f[0] == '#':
			q.tags = append(q.tags, f[1:])
		case len(f) > 1 && f[0] == '!':
			if p, ok := prioAlias[f[1:]]; ok {
				q.prio = append(q.prio, p)
				continue
			}
			q.words = append(q.words, f)
		case strings.HasPrefix(f, "is:") && len(f) > 3:
			q.is = append(q.is, f[3:])
		case strings.HasPrefix(f, "due:") && len(f) > 4:
			if fn := dueFilter(f[4:]); fn != nil {
				q.due = append(q.due, fn)
				continue
			}
			q.words = append(q.words, f)
		default:
			q.words = append(q.words, f)
		}
	}
	return q
}

func dueFilter(v string) func(model.Task, time.Time) bool {
	days := func(t model.Task, now time.Time) (int, bool) {
		d, err := model.ParseDate(t.DueDate)
		if err != nil {
			return 0, false
		}
		return dates.DaysUntil(d, now), true
	}
	switch v {
	case "none":
		return func(t model.Task, _ time.Time) bool { return t.DueDate == "" }
	case "today":
		return func(t model.Task, now time.Time) bool { n, ok := days(t, now); return ok && n == 0 }
	}
	if len(v) < 3 || (v[0] != '<' && v[0] != '>') || v[len(v)-1] != 'd' {
		return nil
	}
	n, err := strconv.Atoi(v[1 : len(v)-1])
	if err != nil {
		return nil
	}
	if v[0] == '<' {
		return func(t model.Task, now time.Time) bool { d, ok := days(t, now); return ok && d < n }
	}
	return func(t model.Task, now time.Time) bool { d, ok := days(t, now); return ok && d > n }
}

func (q query) empty() bool { return q.raw == "" }

// terms are the free words, for highlighting matches.
func (q query) terms() []string { return q.words }

// waiting reports whether t waits for a task that is not done.
type waitFunc func(model.Task) bool

func (q query) match(t model.Task, now time.Time, waiting waitFunc) bool {
	hay := strings.ToLower(t.Title + "\n" + t.Description + "\n" + t.ID)
	for _, w := range q.words {
		if !strings.Contains(hay, w) {
			return false
		}
	}
	if len(q.assignee) > 0 && !anyPrefix(strings.ToLower(t.Assignee), q.assignee) {
		return false
	}
	for _, tg := range q.tags {
		found := false
		for _, have := range t.Tags {
			if strings.HasPrefix(strings.ToLower(have), tg) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	if len(q.prio) > 0 {
		ok := false
		for _, p := range q.prio {
			ok = ok || t.Priority == p
		}
		if !ok {
			return false
		}
	}
	for _, is := range q.is {
		if !isMatch(is, t, now, waiting) {
			return false
		}
	}
	for _, f := range q.due {
		if !f(t, now) {
			return false
		}
	}
	return true
}

func isMatch(is string, t model.Task, now time.Time, waiting waitFunc) bool {
	switch is {
	case "overdue", "late":
		return dates.DueUrgency(t, now) == dates.UrgencyLate
	case "blocked":
		return t.Status == model.StatusBlocked
	case "done":
		return t.Status == model.StatusDone
	case "open":
		return t.Status != model.StatusDone
	case "todo":
		return t.Status == model.StatusTodo
	case "doing", "wip", "in-progress":
		return t.Status == model.StatusInProgress
	case "waiting":
		return waiting != nil && waiting(t)
	}
	return false
}

func anyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}
