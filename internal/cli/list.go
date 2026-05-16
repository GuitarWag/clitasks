package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
)

func newListCmd() *cobra.Command {
	var status, priority, assignee, tags string
	var detailed bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := board.Filter{Assignee: assignee, Tags: splitTags(tags)}
			if status != "" {
				s, err := model.ParseStatus(status)
				if err != nil {
					return err
				}
				f.Status = &s
			}
			if priority != "" {
				p, err := model.ParsePriority(priority)
				if err != nil {
					return err
				}
				f.Priority = &p
			}

			b, err := openBoard()
			if err != nil {
				return err
			}
			tasks := b.List(f)
			out := cmd.OutOrStdout()
			if len(tasks) == 0 {
				fmt.Fprintln(out, styleYellow.Render("No tasks found"))
				return nil
			}
			fmt.Fprintln(out, styleBold.Render(fmt.Sprintf("\nFound %d task(s):", len(tasks))))
			for _, t := range tasks {
				renderTask(out, t, detailed)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status (todo|in-progress|done|blocked)")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "Filter by priority")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Filter by assignee")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Filter by tags (comma-separated)")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed information")
	return cmd
}
