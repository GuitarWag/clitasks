package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
	"github.com/GuitarWag/clitasks/internal/ui"
	"github.com/GuitarWag/clitasks/internal/ui/card"
	"github.com/GuitarWag/clitasks/internal/ui/dates"
	"github.com/GuitarWag/clitasks/internal/ui/icons"
	"github.com/GuitarWag/clitasks/internal/ui/theme"
	"github.com/GuitarWag/clitasks/internal/ui/tlgrid"
)

// output describes where a command prints. rich is true for a terminal:
// commands then use the card, detail and grid renderers. Pipes and agents get
// the plain text formats, which stay stable for parsing.
type output struct {
	rich  bool
	width int
	look  ui.Look
}

type outputKey struct{}

// forceRich makes tests render the terminal output into a buffer.
var forceRich = false

// setupOutput runs before every command. It checks for a terminal before it
// wraps stdout in the color writer, then picks the look.
func setupOutput(cmd *cobra.Command) error {
	r := cmd.Root()
	raw := r.OutOrStdout()
	o := output{rich: forceRich, width: 100}
	if f, ok := raw.(*os.File); ok && term.IsTerminal(f.Fd()) {
		o.rich = true
		if w, _, err := term.GetSize(f.Fd()); err == nil && w > 0 {
			o.width = w
		}
	}
	opts, err := lookOptions(cmd)
	if err != nil {
		return err
	}
	// The CLI does not query the terminal for its background: a terminal that
	// does not answer would stall every command for the 2 s query timeout.
	// The TUI asks without blocking instead.
	o.look = opts.Resolve(darkFromEnv(os.Getenv("COLORFGBG")))
	if !o.rich {
		o.look.Icons = icons.ASCII // pipes and agents get plain glyphs
	}
	applyLook(o.look.Theme)
	r.SetOut(colorprofile.NewWriter(raw, os.Environ()))
	cmd.SetContext(context.WithValue(cmd.Context(), outputKey{}, o))
	return nil
}

func outputOf(cmd *cobra.Command) output {
	if o, ok := cmd.Context().Value(outputKey{}).(output); ok {
		return o
	}
	return output{width: 100, look: ui.Options{}.Resolve(true)}
}

func (o output) ctx() card.Ctx { return card.Ctx{Look: o.look, Now: timelineNow()} }

// applyLook points the plain-output styles at the theme tokens.
func applyLook(th theme.Theme) {
	fg := theme.Fg
	styleSuccess, styleError, styleWarn = fg(th.Success), fg(th.Error), fg(th.Warn)
	styleCyan, styleGreen, styleYellow = fg(th.Info), fg(th.Success), fg(th.Warn)
	styleBlue, styleMagenta, styleRed, styleGray = fg(th.StatusTodo), fg(th.Accent), fg(th.Error), fg(th.Muted)
	styleDim = lipgloss.NewStyle().Faint(true)
	styleBold = lipgloss.NewStyle().Bold(true)
	priorityTheme = th
}

var priorityTheme = theme.Mono(true)

// --- rich renderers ---

// richBoard prints the four status columns of cards side by side.
func richBoard(w io.Writer, o output, info model.Board, by map[model.TaskStatus][]model.Task) {
	th, ic := o.look.Theme, o.look.Icons
	order := []model.TaskStatus{model.StatusTodo, model.StatusInProgress, model.StatusBlocked, model.StatusDone}
	n := min(max((o.width+2)/26, 1), 4)
	cw := (o.width - 2*(n-1)) / n
	fmt.Fprintln(w, card.Pill(info.Name, th.Accent, th.Base, th.Mono)+"  "+theme.Fg(th.Muted).Render(fmt.Sprintf("%d tasks", len(info.Tasks))))
	if info.Description != "" {
		fmt.Fprintln(w, theme.Fg(th.Subtext).Render(info.Description))
	}
	for start := 0; start < len(order); start += n {
		var cols []string
		for _, st := range order[start:min(start+n, len(order))] {
			lines := []string{
				theme.Fg(th.Status(st)).Bold(true).Render(ic.Status(st)+" "+statusTitle(st)) +
					theme.Fg(th.Muted).Render(fmt.Sprintf("  %d", len(by[st]))),
				theme.Fg(th.Overlay0).Render(strings.Repeat("─", cw)),
			}
			if len(by[st]) == 0 {
				lines = append(lines, theme.Fg(th.Muted).Render("Empty"))
			}
			for _, t := range by[st] {
				lines = append(lines, card.Card(t, o.ctx(), card.Opts{Width: cw, ShowID: true}), "")
			}
			cols = append(cols, lipgloss.NewStyle().Width(cw).Render(strings.Join(lines, "\n")))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, lipgloss.JoinHorizontal(lipgloss.Top, joinWithGap(cols, 2)...))
	}
}

func joinWithGap(cols []string, gap int) []string {
	var out []string
	for i, c := range cols {
		if i > 0 {
			out = append(out, strings.Repeat(" ", gap))
		}
		out = append(out, c)
	}
	return out
}

func statusTitle(s model.TaskStatus) string {
	switch s {
	case model.StatusInProgress:
		return "IN PROGRESS"
	case model.StatusBlocked:
		return "BLOCKED"
	case model.StatusDone:
		return "DONE"
	}
	return "TODO"
}

// richRow is one task on one line: status, ID, title and chips.
func richRow(o output, t model.Task) string {
	th, ic := o.look.Theme, o.look.Icons
	parts := []string{
		theme.Fg(th.Status(t.Status)).Render(ic.Status(t.Status)),
		theme.Fg(th.Muted).Render(t.ID),
		theme.Fg(th.Text).Bold(true).Render(t.Title),
		theme.Fg(th.Priority(t.Priority)).Render(ic.Priority(t.Priority) + " " + string(t.Priority)),
	}
	if t.DueDate != "" {
		c := th.Subtext
		switch dates.DueUrgency(t, timelineNow()) {
		case dates.UrgencyLate:
			c = th.Error
		case dates.UrgencyToday:
			c = th.Today
		case dates.UrgencySoon:
			c = th.Warn
		}
		parts = append(parts, theme.Fg(c).Render(ic.Due+" "+dates.DueLabel(t, timelineNow())))
	}
	if t.Assignee != "" {
		parts = append(parts, theme.Fg(th.Subtext).Render(ic.Assignee+t.Assignee))
	}
	for _, tg := range t.Tags {
		parts = append(parts, theme.Fg(th.Info).Render(ic.Tag+tg))
	}
	return strings.Join(parts, " ")
}

// renderTimeline prints the chart, the conflicts and the unscheduled tasks.
// The plain path uses the same grid with no color and ASCII icons.
func renderTimeline(w io.Writer, o output, l timeline.Layout) {
	if len(l.Bars) == 0 && len(l.Unscheduled) == 0 {
		fmt.Fprintln(w, styleYellow.Render("No tasks found"))
		return
	}
	c := o.ctx()
	if !o.rich {
		c.Look.Theme = theme.Mono(true)
	}
	if len(l.Bars) > 0 {
		out := tlgrid.Render(l, tlgrid.BarRows(l), c, tlgrid.Opts{Width: o.width, Selected: -1, ShowID: true})
		for _, ln := range out.Lines {
			fmt.Fprintln(w, strings.TrimRight(ln, " "))
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, styleDim.Render(fmt.Sprintf("%s zoom · %s to %s · │ today",
			out.Zoom, out.From.Format(time.DateOnly), out.To.Format(time.DateOnly))))
		if out.Clipped {
			fmt.Fprintln(w, styleDim.Render(fmt.Sprintf("Range %s to %s does not fit; use --width or --status to show all of it.",
				l.From.Format(time.DateOnly), l.To.Format(time.DateOnly))))
		}
	}
	if len(l.Warnings) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, styleWarn.Render("Warnings:"))
		for _, wn := range l.Warnings {
			fmt.Fprintln(w, styleWarn.Render("  ! "+wn.String()))
		}
	}
	if len(l.Unscheduled) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, styleDim.Render("Unscheduled (due or start is not YYYY-MM-DD):"))
		for _, t := range l.Unscheduled {
			fmt.Fprintf(w, "  %s %s %s\n", styleCyan.Render(t.ID), t.Title, styleDim.Render(badDates(t)))
		}
	}
}

// bar is a horizontal bar of n out of total cells for stats.
func bar(n, total, width int, c lipgloss.Style) string {
	if total == 0 {
		return ""
	}
	filled := n * width / total
	if n > 0 && filled == 0 {
		filled = 1
	}
	return c.Render(strings.Repeat("█", filled)) + theme.Fg(priorityTheme.Overlay0).Render(strings.Repeat("░", width-filled))
}

// darkFromEnv reads COLORFGBG ("fg;bg" or "fg;default;bg"), which iTerm2,
// Konsole and rxvt set. ANSI backgrounds 7 and 15 are light. Without it the
// CLI assumes a dark terminal; TASKS_THEME overrides either way.
func darkFromEnv(colorfgbg string) bool {
	parts := strings.Split(colorfgbg, ";")
	if colorfgbg == "" || len(parts) < 2 {
		return true
	}
	bg := parts[len(parts)-1]
	return bg != "7" && bg != "15"
}
