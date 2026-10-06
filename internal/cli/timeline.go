package cli

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
	"github.com/GuitarWag/clitasks/internal/timeline"
)

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
			o := outputOf(cmd)
			if width > 0 {
				o.width = width
			}
			renderTimeline(cmd.OutOrStdout(), o, timeline.Build(b.List(f), timelineNow()))
			return nil
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (todo|in-progress|done|blocked)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Filter by assignee")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Filter by tags (comma-separated)")
	cmd.Flags().IntVarP(&width, "width", "w", 0, "Chart width in columns (default: terminal width, or 100)")
	return cmd
}

// badDates names the date fields that keep t off the chart.
func badDates(t model.Task) string {
	var bad []string
	if _, err := model.ParseDate(t.DueDate); t.DueDate != "" && err != nil {
		bad = append(bad, "due:"+t.DueDate)
	}
	if _, err := model.ParseDate(t.Start); t.Start != "" && err != nil {
		bad = append(bad, "start:"+t.Start)
	}
	return strings.Join(bad, " ")
}
