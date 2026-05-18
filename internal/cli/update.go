package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
)

func newUpdateCmd() *cobra.Command {
	var title, desc, priority, assignee, tags, due string
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			in := board.UpdateInput{}
			if cmd.Flags().Changed("title") {
				in.Title = &title
			}
			if cmd.Flags().Changed("description") {
				in.Description = &desc
			}
			if cmd.Flags().Changed("priority") {
				p, err := model.ParsePriority(priority)
				if err != nil {
					return err
				}
				in.Priority = &p
			}
			if cmd.Flags().Changed("assignee") {
				in.Assignee = &assignee
			}
			if cmd.Flags().Changed("tags") {
				v := splitTags(tags)
				in.Tags = &v
			}
			if cmd.Flags().Changed("due") {
				in.DueDate = &due
			}

			b, err := openBoard(cmd)
			if err != nil {
				return err
			}
			t, err := b.Update(args[0], in)
			if err != nil {
				if errors.Is(err, board.ErrNotFound) {
					fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+args[0]))
					return nil
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Task updated: "+args[0]))
			renderTask(cmd.OutOrStdout(), t, true)
			return nil
		},
	}
	cmd.Flags().StringVarP(&title, "title", "t", "", "New title")
	cmd.Flags().StringVarP(&desc, "description", "d", "", "New description")
	cmd.Flags().StringVarP(&priority, "priority", "p", "", "New priority")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "New assignee")
	cmd.Flags().StringVar(&tags, "tags", "", "New tags (comma-separated)")
	cmd.Flags().StringVar(&due, "due", "", "New due date")
	return cmd
}
