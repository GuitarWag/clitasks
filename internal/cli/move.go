package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
	"github.com/GuitarWag/clitasks/internal/model"
)

func runMove(cmd *cobra.Command, id string, s model.TaskStatus, successPrefix string, style func(string) string) error {
	b, err := openBoard()
	if err != nil {
		return err
	}
	t, err := b.Move(id, s)
	if err != nil {
		if errors.Is(err, board.ErrNotFound) {
			fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+id))
			return nil
		}
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), style(successPrefix+id))
	renderTask(cmd.OutOrStdout(), t, false)
	return nil
}

func newMoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "move <id> <status>",
		Short:     "Move task to different status (todo|in-progress|done|blocked)",
		Args:      cobra.ExactArgs(2),
		ValidArgs: []string{"todo", "in-progress", "done", "blocked"},
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := model.ParseStatus(args[1])
			if err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ "+err.Error()))
				return nil
			}
			return runMove(cmd, args[0], s, "✓ Task moved to "+string(s)+": ",
				func(t string) string { return styleSuccess.Render(t) })
		},
	}
}
