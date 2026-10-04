package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/theme"
	"github.com/GuitarWag/clitasks/internal/timeline"
)

const defaultTimelineWidth = 100

// timelineNow is the clock for `tasks timeline`. Tests replace it.
var timelineNow = time.Now

func newTimelineCmd() *cobra.Command {
	var status, assignee, tags string
	var width int
	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Show tasks as a Gantt chart",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := board.Filter{Assignee: assignee, Tags: splitTags(tags)}
			if status != "" {
				s, err := model.ParseStatus(status)
				if err != nil {
					return err
				}
				f.Status = &s
			}
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			if width <= 0 {
				width = terminalWidth(cmd.OutOrStdout())
			}
			now := timelineNow()
			l := timeline.Build(b.List(f), now)
			renderTimeline(cmd.OutOrStdout(), l, width, now)
			return nil
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (todo|in-progress|done|blocked)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Filter by assignee")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Filter by tags (comma-separated)")
	cmd.Flags().IntVarP(&width, "width", "w", 0, "Chart width in columns (default: terminal width, or 100)")
	return cmd
}

func renderTimeline(w io.Writer, l timeline.Layout, width int, now time.Time) {
	if len(l.Bars) == 0 && len(l.Unscheduled) == 0 {
		fmt.Fprintln(w, styleYellow.Render("No tasks found"))
		return
	}
	if len(l.Bars) > 0 {
		out := timeline.RenderASCII(l, timeline.ASCIIOptions{Width: width, Today: now})
		fmt.Fprintln(w, styleDim.Render(out.Header))
		for i, row := range out.Rows {
			fmt.Fprintln(w, theme.StatusStyle(l.Bars[i].Task.Status).Render(row))
		}
		scale := "1 column = 1 day"
		if out.Scale == timeline.ScaleWeek {
			scale = "1 column = 1 week"
		}
		fmt.Fprintln(w)
		fmt.Fprintln(w, styleDim.Render(fmt.Sprintf("%s · %s to %s · | today · > open end",
			scale, out.From.Format(time.DateOnly), out.To.Format(time.DateOnly))))
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
			fmt.Fprintf(w, "  %s %s %s\n", styleCyan.Render(t.ID), t.Title, styleDim.Render("due:"+t.DueDate))
		}
	}
}

func terminalWidth(w io.Writer) int {
	if f, ok := w.(*os.File); ok && term.IsTerminal(f.Fd()) {
		if cols, _, err := term.GetSize(f.Fd()); err == nil && cols > 0 {
			return cols
		}
	}
	return defaultTimelineWidth
}
