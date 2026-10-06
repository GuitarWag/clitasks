package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/v3/internal/board"
	"github.com/GuitarWag/clitasks/v3/internal/model"
)

func newAddCmd() *cobra.Command {
	var desc, priority, assignee, tags, due, start, after string
	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := model.ParsePriority(priority)
			if err != nil {
				return err
			}
			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			t, err := b.Add(args[0], board.AddInput{
				Description: desc,
				Priority:    p,
				Assignee:    assignee,
				Tags:        splitTags(tags),
				DueDate:     due,
				Start:       start,
				After:       splitTags(after),
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Task added: "+t.ID))
			renderTask(cmd.OutOrStdout(), t, false)
			return nil
		},
	}
	cmd.Flags().StringVarP(&desc, "description", "d", "", "Task description")
	cmd.Flags().StringVarP(&priority, "priority", "p", "medium", "Priority (low|medium|high|critical)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assignee name")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "Comma-separated tags")
	cmd.Flags().StringVar(&due, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&start, "start", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&after, "after", "", "Comma-separated IDs of tasks this one waits for")
	return cmd
}
