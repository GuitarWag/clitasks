package card

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/GuitarWag/clitasks/v3/internal/model"
	"github.com/GuitarWag/clitasks/v3/internal/timeline"
	"github.com/GuitarWag/clitasks/v3/internal/ui/dates"
)

// Related is the dependency context of one task.
type Related struct {
	After      []model.Task // tasks it waits for that exist on the board
	Dependents []model.Task // tasks that wait for it
	// Conflicts are human sentences from timeline warnings for this task.
	Conflicts []string
}

// Pill renders " text " on a solid color, or in reverse video for Mono.
func Pill(text string, bg, fg color.Color, mono bool) string {
	s := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if mono {
		return s.Reverse(true).Render(text)
	}
	return s.Background(bg).Foreground(fg).Render(text)
}

// StatusLabel is the human name of a status.
func StatusLabel(s model.TaskStatus) string {
	switch s {
	case model.StatusInProgress:
		return "In progress"
	case model.StatusBlocked:
		return "Blocked"
	case model.StatusDone:
		return "Done"
	}
	return "Todo"
}

// Detail renders every field of t as lines exactly width columns wide on the
// Mantle background. The caller cuts or scrolls the lines to its height.
func Detail(t model.Task, rel Related, c Ctx, width int) []string {
	th, ic := c.Look.Theme, c.Look.Icons
	p := Painter{BG: th.Mantle, Mono: th.Mono}
	inner := max(width-4, 10)
	var out []string
	add := func(s string) { out = append(out, p.Pad(p.Style().Render("  ")+s, width)) }
	blank := func() { add("") }
	muted, sub, text := p.Fg(th.Muted), p.Fg(th.Subtext), p.Fg(th.Text)
	label := func(l string) string { return muted.Render(fmt.Sprintf("%-9s", l)) }

	blank()
	add(muted.Render(t.ID))
	for _, l := range strings.Split(ansi.Wordwrap(t.Title, inner, " -"), "\n") {
		add(text.Bold(true).Render(l))
	}
	blank()
	add(Pill(ic.Status(t.Status)+" "+StatusLabel(t.Status), th.Status(t.Status), th.Base, th.Mono) +
		p.Style().Render(" ") +
		Pill(ic.Priority(t.Priority)+" "+string(t.Priority), th.Priority(t.Priority), th.Base, th.Mono))
	blank()

	dateLine := func(name, icon, v string, fg color.Color) {
		if v == "" {
			return
		}
		d, err := model.ParseDate(v)
		if err != nil {
			add(label(name) + p.Fg(th.Error).Render(v+"  not a date"))
			return
		}
		add(label(name) + p.Fg(fg).Render(icon+" "+d.Format("Mon Jan 2, 2006")) +
			muted.Render("  "+dates.Relative(d, c.Now)))
	}
	dateLine("Start", ic.Start, t.Start, th.Text)
	dueFg := th.Text
	switch dates.DueUrgency(t, c.Now) {
	case dates.UrgencyLate:
		dueFg = th.Error
	case dates.UrgencyToday:
		dueFg = th.Today
	case dates.UrgencySoon:
		dueFg = th.Warn
	}
	dateLine("Due", ic.Due, t.DueDate, dueFg)
	if t.Assignee != "" {
		add(label("Assignee") + sub.Render(ic.Assignee+t.Assignee))
	}
	if len(t.Tags) > 0 {
		var tags []string
		for _, tg := range t.Tags {
			tags = append(tags, ic.Tag+tg)
		}
		add(label("Tags") + p.Fg(th.Info).Render(ansi.Truncate(strings.Join(tags, " "), inner-9, "…")))
	}

	link := func(o model.Task) {
		add(p.Fg(th.Status(o.Status)).Render(" "+ic.Status(o.Status)+" ") +
			text.Render(ansi.Truncate(o.Title, inner-4, "…")))
	}
	if len(t.After) > 0 {
		blank()
		add(sub.Bold(true).Render(ic.After + " Waits for"))
		for _, o := range rel.After {
			link(o)
		}
		if n := len(t.After) - len(rel.After); n > 0 {
			add(muted.Render(fmt.Sprintf("   %d missing task(s)", n)))
		}
	}
	if len(rel.Dependents) > 0 {
		blank()
		add(sub.Bold(true).Render(ic.After + " Blocks"))
		for _, o := range rel.Dependents {
			link(o)
		}
	}
	for _, cf := range rel.Conflicts {
		blank()
		for _, l := range strings.Split(ansi.Wordwrap(ic.Warn+" "+cf, inner, " "), "\n") {
			add(p.Fg(th.Error).Render(l))
		}
	}
	if t.Description != "" {
		blank()
		add(sub.Bold(true).Render("Description"))
		for _, l := range strings.Split(ansi.Wordwrap(t.Description, inner, " -"), "\n") {
			add(text.Render(l))
		}
	}
	blank()
	if !t.CreatedAt.IsZero() {
		add(muted.Render("Created " + dates.Relative(timeline.Day(t.CreatedAt), c.Now) +
			" · updated " + dates.Relative(timeline.Day(t.UpdatedAt), c.Now)))
	}
	return out
}
