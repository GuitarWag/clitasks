package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/GuitarWag/clitasks/internal/board"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := openBoard()
			if err != nil {
				return err
			}
			if err := b.Delete(args[0]); err != nil {
				if errors.Is(err, board.ErrNotFound) {
					fmt.Fprintln(cmd.OutOrStdout(), styleError.Render("✗ Task not found: "+args[0]))
					return nil
				}
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), styleSuccess.Render("✓ Task deleted: "+args[0]))
			return nil
		},
	}
}
